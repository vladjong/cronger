package cronger

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vladjong/cronger/model"
	"github.com/vladjong/cronger/repository"
	"github.com/vladjong/cronger/repository/sqlx_repository"
)

const (
	SqlxName      = "sqlx"
	UndefinedName = "undefined"
)

const (
	Sqlx = iota
)

type Cronger struct {
	cfg      *Config
	schedule *gocron.Scheduler
	repo     repository.Repository
	cache    map[string]model.Job
	mu       sync.Mutex
}

type Config struct {
	Loc        *time.Location
	TypeClient int
	Client     interface{}
}

func New(cfg *Config) (*Cronger, error) {
	schedule := gocron.NewScheduler(cfg.Loc)
	schedule.TagsUnique()

	c := &Cronger{
		cfg:      cfg,
		schedule: schedule,
	}

	if err := c.checkDriver(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cronger) AddJob(tag, expression string, task func()) error {
	job := model.Job{
		Id:         uuid.New(),
		Tag:        tag,
		Expression: tag,
	}

	c.mu.Lock()
	if _, ok := c.cache[tag]; ok {
		return fmt.Errorf("job tag=%s is exist", tag)
	}
	c.cache[tag] = job
	c.mu.Unlock()

	if err := c.addJob(job); err != nil {
		return err
	}

	if _, err := c.schedule.Cron(expression).Tag(tag).Do(task); err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (c *Cronger) RemoveJob(tag string) error {
	c.mu.Lock()
	if _, ok := c.cache[tag]; !ok {
		return fmt.Errorf("job tag=%s don't exist", tag)
	}
	delete(c.cache, tag)
	c.mu.Unlock()

	if err := c.removeJob(tag); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) GetAllJob() []model.Job {
	c.mu.Lock()
	defer c.mu.Unlock()

	jobs := make([]model.Job, 0, len(c.cache))
	for _, job := range c.cache {
		jobs = append(jobs, job)
	}
	return jobs
}

func (c *Cronger) addJob(job model.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.repo.AddJob(ctx, job); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) removeJob(tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.repo.RemoveJob(ctx, tag); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) checkDriver() error {
	switch c.cfg.TypeClient {
	case Sqlx:
		db, ok := c.cfg.Client.(*sqlx.DB)
		if !ok {
			typeOf := reflect.TypeOf(c.cfg.Client).String()
			return NewIncorrectClientError(SqlxName, typeOf)
		}
		c.repo = sqlx_repository.New(db)

	default:
		typeOf := reflect.TypeOf(c.cfg.Client).String()
		return NewIncorrectClientError(UndefinedName, typeOf)
	}
	return nil
}

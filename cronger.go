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

const (
	timeOut = 5 * time.Second
)

type cronger struct {
	cfg        *Config
	schedule   *gocron.Scheduler
	repo       repository.Repository
	backupJobs map[string]model.Job
	mu         sync.Mutex
}

type Config struct {
	Loc        *time.Location
	TypeClient int
	Client     interface{}
	IsMigrate  bool
}

func New(cfg *Config) (*cronger, error) {
	schedule := gocron.NewScheduler(cfg.Loc)
	schedule.TagsUnique()

	c := &cronger{
		cfg:      cfg,
		schedule: schedule,
	}

	if err := c.checkDriver(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	if !cfg.IsMigrate {
		if err := c.repo.Create(ctx); err != nil {
			return nil, err
		}
	}

	jobs, err := c.repo.BackupJobs(ctx)
	if err != nil {
		return nil, err
	}

	backupJobs := make(map[string]model.Job)

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, job := range jobs {
		backupJobs[job.Tag] = job
	}
	c.backupJobs = backupJobs

	return c, nil
}

func (c *cronger) AddJob(title, tag, expression string, task func()) error {
	if _, err := c.schedule.Cron(expression).Tag(tag).Do(task); err != nil {
		return fmt.Errorf("create job: %w", err)
	}

	job := model.Job{
		Id:         uuid.New(),
		Title:      title,
		Tag:        tag,
		Expression: expression,
		IsWork:     true,
	}

	if err := c.addJob(job); err != nil {
		return err
	}

	c.deleteJobInBackup(tag)
	return nil
}

func (c *cronger) RemoveJob(tag string) error {
	if err := c.removeJob(tag); err != nil {
		return err
	}

	if err := c.schedule.RemoveByTag(tag); err != nil {
		return err
	}

	c.deleteJobInBackup(tag)
	return nil
}

func (c *cronger) Jobs() ([]model.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	jobs, err := c.repo.Jobs(ctx)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *cronger) BackupJobs() []model.Job {
	c.mu.Lock()
	defer c.mu.Unlock()

	jobs := make([]model.Job, 0, len(c.backupJobs))
	for _, job := range c.backupJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (c *cronger) StartAsync() {
	c.schedule.StartAsync()
}

func (c *cronger) Stop() {
	c.schedule.Stop()
}

func (c *cronger) addJob(job model.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	if err := c.repo.AddJob(ctx, job); err != nil {
		return err
	}
	return nil
}

func (c *cronger) removeJob(tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	if err := c.repo.RemoveJob(ctx, tag); err != nil {
		return err
	}
	return nil
}

func (c *cronger) checkDriver() error {
	switch c.cfg.TypeClient {
	case Sqlx:
		db, ok := c.cfg.Client.(*sqlx.DB)
		if !ok {
			typeOf := reflect.TypeOf(c.cfg.Client)
			return newIncorrectClientError(SqlxName, typeOf)
		}
		c.repo = sqlx_repository.New(db)

	default:
		typeOf := reflect.TypeOf(c.cfg.Client)
		return newIncorrectClientError(UndefinedName, typeOf)
	}
	return nil
}

func (c *cronger) deleteJobInBackup(tag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.backupJobs[tag]; ok {
		delete(c.backupJobs, tag)
	}
}

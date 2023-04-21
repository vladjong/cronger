package cronger

import (
	"context"
	"fmt"
	"reflect"
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

type cronger struct {
	cfg      *Config
	schedule *gocron.Scheduler
	repo     repository.Repository
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

	if !cfg.IsMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.repo.Create(ctx); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *cronger) AddJob(tag, expression string, task func()) error {
	job := model.Job{
		Id:         uuid.New(),
		Tag:        tag,
		Expression: expression,
	}

	if err := c.addJob(job); err != nil {
		return err
	}

	if _, err := c.schedule.Cron(expression).Tag(tag).Do(task); err != nil {
		return fmt.Errorf("create job: %w", err)
	}
	return nil
}

func (c *cronger) RemoveJob(tag string) error {
	if err := c.removeJob(tag); err != nil {
		return err
	}
	return nil
}

func (c *cronger) GetAllJob() ([]model.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobs, err := c.repo.Jobs(ctx)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *cronger) StartAsync() {
	c.schedule.StartAsync()
}

func (c *cronger) addJob(job model.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.repo.AddJob(ctx, job); err != nil {
		return err
	}
	return nil
}

func (c *cronger) removeJob(tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			typeOf := reflect.TypeOf(c.cfg.Client).String()
			return newIncorrectClientError(SqlxName, typeOf)
		}
		c.repo = sqlx_repository.New(db)

	default:
		typeOf := reflect.TypeOf(c.cfg.Client).String()
		return newIncorrectClientError(UndefinedName, typeOf)
	}
	return nil
}

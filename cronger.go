package cronger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
)

var (
	ErrJobIntervalNotFound = errors.New("job interval not found")
)

var validate *validator.Validate

const (
	Unlimited = 0
)

const (
	_timeOut = time.Second * 5
)

type Cronger struct {
	cfg           *Config
	schedule      *gocron.Scheduler
	mu            sync.Mutex
	suspendedJobs map[string]Job
}

type Config struct {
	Loc *time.Location
	// Client type for working with persistent storage.
	Repository Repository
	// Time interval for starting tasks.
	JobIntervals map[string]time.Duration
}

type Job struct {
	// Unique job ID for the gocron library.
	Tag string `db:"tag" validate:"required,uuid"`
	// ID of the object.
	ID string `db:"id" validate:"required,uuid"`
	// Expression in cron format.
	Expression     string         `db:"expression" validate:"required,cron"`
	FunctionName   string         `db:"function_name" validate:"required"`
	FunctionFields FunctionFields `db:"function_fields" validate:"required"`
	// Limit run job.
	Limit             uint      `db:"limit" validate:"required,gte=0,lte=100"`
	Status            Status    `db:"status"`
	StatusDescription string    `db:"status_description"`
	CreatedAt         time.Time `db:"created_at" goqu:"skipupdate"`
}

func (j Job) CheckUpdate() error {
	if err := validate.Var(j.Tag, "required,uuid"); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	if len(j.ID) != 0 {
		if err := validate.Var(j.ID, "uuid"); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}
	if len(j.Expression) != 0 {
		if err := validate.Var(j.Expression, "cron"); err != nil {
			return fmt.Errorf("validate: %w", err)
		}
	}
	if err := validate.Var(j.Limit, "gte=0,lte=100"); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

type Fields struct {
	Job `validate:"required"`

	Task func() error `validate:"required"`
}

type FunctionFields []interface{}

func (f *FunctionFields) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to cast value to []byte: %v", value)
	}

	return json.Unmarshal(bytes, &f)
}

func (f FunctionFields) Value() (driver.Value, error) {
	return json.Marshal(f)
}

type Status string

const (
	Working   Status = "working"
	Suspended Status = "suspended"
	Done      Status = "done"
	Failed    Status = "failed"
	Cancelled Status = "cancelled"
)

func (s Status) String() string {
	return string(s)
}

func init() {
	validate = validator.New()
}

func New(cfg *Config) (*Cronger, error) {
	schedule := gocron.NewScheduler(cfg.Loc)
	schedule.TagsUnique()

	c := &Cronger{
		cfg:      cfg,
		schedule: schedule,
	}

	if err := c.setSuspendJob(); err != nil {
		return nil, err
	}

	if _, err := schedule.Cron("*/1 * * * *").Do(c.jobUpdateStatusDone); err != nil {
		return nil, err
	}

	schedule.StartAsync()
	return c, nil
}

func (c *Cronger) setSuspendJob() error {
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	jobs, err := c.cfg.Repository.SuspendJobs(ctx)
	if err != nil {
		return err
	}

	suspendedJobs := make(map[string]Job, len(jobs))
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, job := range jobs {
		suspendedJobs[job.Tag] = job
	}
	c.suspendedJobs = suspendedJobs
	return nil
}

func (c *Cronger) jobUpdateStatusDone() {
	data, err := c.Jobs()
	if err != nil {
		return
	}

	jobs := make(map[string]Job, len(data))
	for _, job := range data {
		jobs[job.Tag] = job
	}

	for _, job := range c.schedule.Jobs() {
		tags := job.Tags()
		runCount := job.FinishedRunCount()
		if len(tags) == 0 || runCount == 0 || job.IsRunning() {
			continue
		}

		data, ok := jobs[tags[0]]
		if !ok {
			continue
		}

		if int(data.Limit) != runCount || data.Limit == Unlimited {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
		if err := c.cfg.Repository.UpdateStatus(ctx, data.Tag, Done); err != nil {
			log.Println(err)
		}
		cancel()
	}
}

func (c *Cronger) Jobs() ([]Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	jobs, err := c.cfg.Repository.Jobs(ctx)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *Cronger) SuspendJobs() []Job {
	c.mu.Lock()
	defer c.mu.Unlock()

	jobs := make([]Job, 0, len(c.suspendedJobs))
	for _, job := range c.suspendedJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (c *Cronger) Recover(tag string) error {
	if err := c.schedule.RunByTag(tag); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) Add(in Fields) error {
	if err := validate.Struct(&in); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	job := in.Job
	job.Status = Working

	schedule := c.schedule.Cron(job.Expression).Tag(job.Tag)
	if job.Limit != Unlimited {
		schedule.LimitRunsTo(int(job.Limit))
	}
	if _, err := schedule.Do(func() {
		c.Template(job, in.Task)
	}); err != nil {
		return fmt.Errorf("create job: %w", err)
	}

	if err := c.add(job); err != nil {
		if err := c.schedule.RemoveByTag(job.Tag); err != nil {
			return fmt.Errorf("remove job: %w", err)
		}
		return err
	}

	c.deleteSuspendJob(in.Tag)
	return nil
}

func (c *Cronger) add(in Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	if err := c.cfg.Repository.Add(ctx, in); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) deleteSuspendJob(tag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.suspendedJobs, tag)
}

func (c *Cronger) GetExpression(time time.Time) string {
	_, month, day := time.Date()
	hour, minute, _ := time.Clock()
	return fmt.Sprintf("%d %d %d %d *", minute, hour, day, month)
}

func (c *Cronger) JobInterval(job string) (time.Duration, error) {
	interval, ok := c.cfg.JobIntervals[job]
	if !ok {
		return 0, ErrJobIntervalNotFound
	}
	return interval, nil
}

func (c *Cronger) SetStatusCancelled(ids []string, functionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	tags, err := c.cfg.Repository.SetStatusCancelled(ctx, ids, functionName)
	if err != nil {
		return fmt.Errorf("SuspendJobs in repo: %w", err)
	}

	for _, tag := range tags {
		if err := c.schedule.RemoveByTag(tag); err != nil {
			return fmt.Errorf("SuspendJobs in schedule: %w", err)
		}
	}
	return nil
}

func (c *Cronger) Remove(tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	if err := c.cfg.Repository.Remove(ctx, tag); err != nil {
		return err
	}

	if err := c.schedule.RemoveByTag(tag); err != nil {
		return fmt.Errorf("remove job: %w", err)
	}

	c.deleteSuspendJob(tag)
	return nil
}

func (c *Cronger) Update(in Job) error {
	if err := in.CheckUpdate(); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	value := structToMap(in)
	ctx, cancel := context.WithTimeout(context.Background(), _timeOut)
	defer cancel()

	if err := c.cfg.Repository.Update(ctx, in.Tag, value); err != nil {
		return err
	}
	return nil
}

func (c *Cronger) Template(job Job, fnc func() error) {
	if err := fnc(); err != nil {
		job.Status = Failed
		job.StatusDescription = err.Error()
		if err := c.Update(job); err != nil {
			log.Printf("set done: %v\n", err)
		}
		return
	}

	job.Status = Done
	job.StatusDescription = ""
	if err := c.Update(job); err != nil {
		log.Printf("set done: %v\n", err)
	}
}

func structToMap(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(s)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("db")
		if !field.IsZero() {
			result[tag] = field.Interface()
		}
	}
	delete(result, _tag)
	return result
}

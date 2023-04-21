package sqlx_repository

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vladjong/cronger/model"
)

const (
	jobsTable = "jobs"
)

type Job struct {
	Id         uuid.UUID `db:"id"`
	Tag        string    `db:"tag"`
	Expression string    `db:"expression"`
}

type sqlxRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *sqlxRepository {
	return &sqlxRepository{
		db: db,
	}
}

func (r *sqlxRepository) AddJob(ctx context.Context, in model.Job) error {
	job := Job{
		Id:         in.Id,
		Tag:        in.Tag,
		Expression: in.Expression,
	}

	query, _, err := goqu.Insert(jobsTable).Rows(job).ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func (r *sqlxRepository) GetJobs(ctx context.Context) ([]model.Job, error) {
	query, _, err := goqu.From(jobsTable).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	jobs := []Job{}
	if err := r.db.SelectContext(ctx, &jobs, query); err != nil {
		return nil, fmt.Errorf("insert job: %w", err)
	}

	out := make([]model.Job, len(jobs))
	for i, val := range jobs {
		job := model.Job{
			Id:         val.Id,
			Tag:        val.Tag,
			Expression: val.Expression,
		}
		out[i] = job
	}
	return out, nil
}

func (r *sqlxRepository) RemoveJob(ctx context.Context, tag string) error {
	query, _, err := goqu.Delete(jobsTable).
		Where(goqu.C("tag").Eq(tag)).
		Returning("id").ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}

	id := ""
	row := r.db.QueryRowContext(ctx, query)
	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("tag=%s don't exist: %w", tag, err)
	}
	return nil
}

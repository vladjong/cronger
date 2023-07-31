package cronger

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
)

const (
	_jobsTable    = "jobs"
	_tag          = "tag"
	_status       = "status"
	_functionName = "function_name"
	_id           = "id"
)

type SqlxRepository struct {
	db *sqlx.DB
}

func NewSqlx(db *sqlx.DB) *SqlxRepository {
	return &SqlxRepository{
		db: db,
	}
}

func (r *SqlxRepository) Jobs(ctx context.Context) ([]Job, error) {
	query, _, err := goqu.From(_jobsTable).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	var jobs []Job
	if err := r.db.SelectContext(ctx, &jobs, query); err != nil {
		return nil, fmt.Errorf("select jobs: %w", err)
	}
	return jobs, nil
}

func (r *SqlxRepository) JobsByStatus(ctx context.Context, status Status) ([]Job, error) {
	query, _, err := goqu.From(_jobsTable).
		Where(goqu.C(_status).Eq(status.String())).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	var jobs []Job
	if err := r.db.SelectContext(ctx, &jobs, query); err != nil {
		return nil, fmt.Errorf("select jobs by statust=%s: %w", status.String(), err)
	}
	return jobs, nil
}

func (r *SqlxRepository) Add(ctx context.Context, in Job) error {
	query, _, err := goqu.Insert(_jobsTable).
		Rows(in).
		OnConflict(goqu.DoUpdate(_tag, in)).
		ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func (r *SqlxRepository) SuspendJobs(ctx context.Context) ([]Job, error) {
	tx, err := r.db.Beginx()
	defer func() {
		_ = tx.Rollback()
	}()
	if err != nil {
		return nil, err
	}

	queryUpdate, _, err := goqu.Update(_jobsTable).
		Where(goqu.C(_status).Eq(Working)).
		Set(goqu.Record{
			_status: Suspended.String(),
		}).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, queryUpdate); err != nil {
		return nil, fmt.Errorf("update jobs: %w", err)
	}

	query, _, err := goqu.From(_jobsTable).
		Where(goqu.C(_status).Eq(Suspended)).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	var jobs []Job
	if err := tx.SelectContext(ctx, &jobs, query); err != nil {
		return nil, fmt.Errorf("select jobs: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *SqlxRepository) SetStatusCancelled(ctx context.Context, ids []string, functionName string) ([]string, error) {
	tx, err := r.db.Beginx()
	defer func() {
		_ = tx.Rollback()
	}()
	if err != nil {
		return nil, err
	}

	getQuery, _, err := goqu.From(_jobsTable).
		Where(
			goqu.C(_status).Eq(Working),
			goqu.C(_functionName).Eq(functionName),
			goqu.C(_id).In(ids),
		).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	var jobs []Job
	if err := tx.SelectContext(ctx, &jobs, getQuery); err != nil {
		return nil, fmt.Errorf("select jobs: %w", err)
	}

	tags := make([]string, len(jobs))
	for i, job := range jobs {
		tags[i] = job.Tag
	}

	updateQuery, _, err := goqu.Update(_jobsTable).
		Where(goqu.C("tag").In(tags)).
		Set(goqu.Record{
			_status: Cancelled.String(),
		}).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("configure query: %w", err)
	}

	if _, err := tx.ExecContext(ctx, updateQuery); err != nil {
		return nil, fmt.Errorf("update jobs: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *SqlxRepository) Remove(ctx context.Context, tag string) error {
	query, _, err := goqu.Delete(_jobsTable).
		Where(goqu.C(_tag).Eq(tag)).
		Returning("tag").ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("delete job = %s: %w", tag, err)
	}
	return nil
}

func (r *SqlxRepository) Update(ctx context.Context, tag string, in map[string]interface{}) error {
	updateQuery, _, err := goqu.Update(_jobsTable).
		Where(goqu.C("tag").Eq(tag)).Set(in).ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, updateQuery); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

func (r *SqlxRepository) UpdateStatus(ctx context.Context, tag string, status Status) error {
	updateQuery, _, err := goqu.Update(_jobsTable).
		Where(goqu.C("tag").Eq(tag)).
		Set(goqu.Record{
			_status: status.String(),
		}).ToSQL()
	if err != nil {
		return fmt.Errorf("configure query: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, updateQuery); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

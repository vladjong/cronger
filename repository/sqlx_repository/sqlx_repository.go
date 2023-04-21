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

const (
	queryCreate = `
	CREATE TABLE product (
		id uuid primary key,
		tag text unique not null,
		expression varchar(25) not null
	);`
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

func (r *sqlxRepository) Create(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, queryCreate); err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
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

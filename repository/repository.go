package repository

import (
	"context"

	"github.com/vladjong/cronger/model"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.0  --name Repository
type Repository interface {
	AddJob(ctx context.Context, in model.Job) error
	Jobs(ctx context.Context) ([]model.Job, error)
	BackupJobs(ctx context.Context) ([]model.Job, error)
	RemoveJob(ctx context.Context, tag string) error
	Create(ctx context.Context) error
}

package repository

import (
	"context"

	"github.com/vladjong/cronger/model"
)

type Repository interface {
	AddJob(ctx context.Context, in model.Job) error
	GetJobs(ctx context.Context) ([]model.Job, error)
	RemoveJob(ctx context.Context, tag string) error
}

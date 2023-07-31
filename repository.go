package cronger

import (
	"context"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.0  --name Repository
type Repository interface {
	Add(ctx context.Context, in Job) error
	Jobs(ctx context.Context) ([]Job, error)
	JobsByStatus(ctx context.Context, status Status) ([]Job, error)
	Remove(ctx context.Context, tag string) error
	Update(ctx context.Context, tag string, in map[string]interface{}) error
	UpdateStatus(ctx context.Context, tag string, status Status) error
	SuspendJobs(ctx context.Context) ([]Job, error)
	SetStatusCancelled(ctx context.Context, ids []string, functionName string) ([]string, error)
}

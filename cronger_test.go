package cronger

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vladjong/cronger/model"
	"github.com/vladjong/cronger/repository/mocks"
)

func TestCheckDriver(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "incorrect_client_undefined_nil",
			args: args{
				cfg: &Config{
					TypeClient: 22,
					Client:     nil,
				},
			},
			wantErr: true,
		},
		{
			name: "error_if_failed_check_driver_client_exist",
			args: args{
				cfg: &Config{
					TypeClient: Sqlx,
					Client:     Config{},
				},
			},
			wantErr: true,
		},
		{
			name: "successful_sqlx",
			args: args{
				cfg: &Config{
					TypeClient: Sqlx,
					Client:     &sqlx.DB{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cronger{
				cfg: tt.args.cfg,
			}
			err := c.checkDriver()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.ErrorIs(t, err, nil)
			}
		})
	}
}

func TestDeleteJobInBackup(t *testing.T) {
	c := &cronger{
		backupJobs: map[string]model.Job{
			"job1": {},
			"job2": {},
		},
	}

	c.deleteJobInBackup("job1")

	if _, ok := c.backupJobs["job1"]; ok {
		t.Error("job1 should have been deleted from backup jobs")
	}

	if _, ok := c.backupJobs["job2"]; !ok {
		t.Error("job2 should still be in backup jobs")
	}

	c.deleteJobInBackup("job3")

	if _, ok := c.backupJobs["job3"]; ok {
		t.Error("job3 should not have been added to backup jobs")
	}
}

func TestAddJob(t *testing.T) {
	type args struct {
		ctx context.Context
		job model.Job
	}
	tests := []struct {
		name    string
		args    args
		mock    func(repo *mocks.Repository, job model.Job)
		wantErr bool
	}{
		{
			name: "successful",
			args: args{
				job: model.Job{
					Id:         uuid.New(),
					Title:      "job",
					Tag:        "tag",
					Expression: "0 * * * * *",
					IsWork:     true,
				},
			},
			mock: func(repo *mocks.Repository, job model.Job) {
				repo.On("AddJob", mock.Anything, job).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				job: model.Job{
					Id:         uuid.New(),
					Title:      "job",
					Tag:        "tag",
					Expression: "0 * * * * *",
					IsWork:     true,
				},
			},
			mock: func(repo *mocks.Repository, job model.Job) {
				repo.On("AddJob", mock.Anything, job).Return(fmt.Errorf("error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.mock(repo, tt.args.job)

			c := &cronger{
				repo: repo,
			}

			err := c.addJob(tt.args.job)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.ErrorIs(t, err, nil)
		})
	}
}

func TestRemoveJob(t *testing.T) {
	type fields struct {
		schedule   *gocron.Scheduler
		backupJobs map[string]model.Job
	}
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func(repo *mocks.Repository, tag string)
		wantErr bool
	}{
		{
			name: "successful_dont_exist_backup",
			fields: fields{
				schedule:   gocron.NewScheduler(time.UTC),
				backupJobs: map[string]model.Job{},
			},
			args: args{
				tag: "test",
			},
			mock: func(repo *mocks.Repository, tag string) {
				repo.On("RemoveJob", mock.Anything, tag).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "successful_exist_backup",
			fields: fields{
				schedule:   gocron.NewScheduler(time.UTC),
				backupJobs: map[string]model.Job{"test": {}},
			},
			args: args{
				tag: "test",
			},
			mock: func(repo *mocks.Repository, tag string) {
				repo.On("RemoveJob", mock.Anything, tag).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error_repository_side",
			fields: fields{
				schedule:   gocron.NewScheduler(time.UTC),
				backupJobs: map[string]model.Job{"test": {}},
			},
			args: args{
				tag: "test",
			},
			mock: func(repo *mocks.Repository, tag string) {
				repo.On("RemoveJob", mock.Anything, tag).Return(fmt.Errorf("repository error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.mock(repo, tt.args.tag)
			c := &cronger{
				schedule:   tt.fields.schedule,
				repo:       repo,
				backupJobs: tt.fields.backupJobs,
			}
			c.StartAsync()
			if _, err := c.schedule.Cron("* * * * *").Tag(tt.args.tag).Do(func() {}); err != nil {
				assert.Error(t, err)
			}
			err := c.RemoveJob(tt.args.tag)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestJobs(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(repo *mocks.Repository)
		want    []model.Job
		wantErr bool
	}{
		{
			name: "successful",
			mock: func(repo *mocks.Repository) {
				repo.On("Jobs", mock.Anything).Return([]model.Job{{Tag: "test1"}, {Tag: "test2"}}, nil)
			},
			want:    []model.Job{{Tag: "test1"}, {Tag: "test2"}},
			wantErr: false,
		},
		{
			name: "error_repository",
			mock: func(repo *mocks.Repository) {
				repo.On("Jobs", mock.Anything).Return(nil, fmt.Errorf("error repository"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.mock(repo)
			c := &cronger{
				repo: repo,
			}
			got, err := c.Jobs()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestBackupJobs(t *testing.T) {
	type fields struct {
		backupJobs map[string]model.Job
	}
	tests := []struct {
		name   string
		fields fields
		want   []model.Job
	}{
		{
			name: "successful",
			fields: fields{
				backupJobs: map[string]model.Job{"test1": {Tag: "test1"}, "test2": {Tag: "test2"}, "test3": {Tag: "test3"}},
			},
			want: []model.Job{{Tag: "test1"}, {Tag: "test2"}, {Tag: "test3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cronger{
				backupJobs: tt.fields.backupJobs,
			}
			got := c.BackupJobs()
			sort.Slice(got, func(i, j int) bool {
				return got[i].Tag < got[j].Tag
			})
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestBackup(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(repo *mocks.Repository)
		wantErr bool
	}{
		{
			name: "successful",
			mock: func(repo *mocks.Repository) {
				repo.On("BackupJobs", mock.Anything).Return([]model.Job{{Tag: "test1"}, {Tag: "test2"}}, nil)
			},
			wantErr: false,
		},
		{
			name: "error_repository",
			mock: func(repo *mocks.Repository) {
				repo.On("BackupJobs", mock.Anything).Return(nil, fmt.Errorf("error repository"))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			tt.mock(repo)
			c := &cronger{
				repo: repo,
			}
			err := c.backup()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
		})
	}
}

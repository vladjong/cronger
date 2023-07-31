package cronger_test

//func TestAdd(t *testing.T) {
//	tag := uuid.NewString()
//	id := uuid.NewString()
//	type args struct {
//		job   cronger.Job
//		field cronger.Fields
//	}
//	tests := []struct {
//		name                     string
//		args                     args
//		mockAdd                  func(repo *mocks.Repository, job cronger.Job)
//		mockWorkingUpdateSuspend func(repo *mocks.Repository)
//		wantErr                  bool
//	}{
//		{
//			name: "successful",
//			args: args{
//				field: cronger.Fields{
//					TaskDetail: cronger.TaskDetail{
//						Tag:            tag,
//						ID:             id,
//						Expression:     "* * * * *",
//						FunctionName:   "test",
//						FunctionFields: []interface{}{1, "test"},
//						Limit:          1,
//					},
//					Task: func() {
//						fmt.Println("test")
//					},
//				},
//				job: cronger.Job{
//					TaskDetail: cronger.TaskDetail{
//						Tag:            tag,
//						ID:             id,
//						Expression:     "* * * * *",
//						FunctionName:   "test",
//						FunctionFields: []interface{}{1, "test"},
//						Limit:          1,
//					},
//					Status: cronger.Working,
//				},
//			},
//			mockWorkingUpdateSuspend: func(repo *mocks.Repository) {
//				repo.On("WorkingUpdateSuspend", mock.Anything).Return([]cronger.Job{}, nil)
//			},
//			mockAdd: func(repo *mocks.Repository, job cronger.Job) {
//				repo.On("Add", mock.Anything, job).Return(nil)
//			},
//			wantErr: false,
//		},
//		{
//			name: "error fields",
//			args: args{
//				field: cronger.Fields{
//					TaskDetail: cronger.TaskDetail{
//						Tag:            "",
//						ID:             "",
//						Expression:     "",
//						FunctionName:   "",
//						FunctionFields: nil,
//						Limit:          0,
//					},
//				},
//				job: cronger.Job{},
//			},
//			mockAdd: func(repo *mocks.Repository, job cronger.Job) {
//			},
//			mockWorkingUpdateSuspend: func(repo *mocks.Repository) {
//				repo.On("WorkingUpdateSuspend", mock.Anything).Return([]cronger.Job{}, nil)
//			},
//			wantErr: true,
//		},
//		{
//			name: "error",
//			args: args{
//				field: cronger.Fields{
//					TaskDetail: cronger.TaskDetail{
//						Tag:            tag,
//						ID:             id,
//						Expression:     "* * * * *",
//						FunctionName:   "test",
//						FunctionFields: []interface{}{1, "test"},
//						Limit:          1,
//					},
//					Task: func() {
//						fmt.Println("test")
//					},
//				},
//				job: cronger.Job{
//					TaskDetail: cronger.TaskDetail{
//						Tag:            tag,
//						ID:             id,
//						Expression:     "* * * * *",
//						FunctionName:   "test",
//						FunctionFields: []interface{}{1, "test"},
//						Limit:          1,
//					},
//					Status: cronger.Working,
//				},
//			},
//			mockWorkingUpdateSuspend: func(repo *mocks.Repository) {
//				repo.On("WorkingUpdateSuspend", mock.Anything).Return([]cronger.Job{}, nil)
//			},
//			mockAdd: func(repo *mocks.Repository, job cronger.Job) {
//				repo.On("Add", mock.Anything, job).Return(fmt.Errorf("error"))
//			},
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			repo := mocks.NewRepository(t)
//			tt.mockAdd(repo, tt.args.job)
//			tt.mockWorkingUpdateSuspend(repo)
//
//			c, err := cronger.New(&cronger.Config{
//				Repository: repo,
//				Loc:        time.UTC,
//			})
//			assert.Nil(t, err)
//
//			err = c.Add(tt.args.field)
//			if tt.wantErr {
//				assert.Error(t, err)
//				return
//			}
//			assert.Nil(t, err)
//		})
//	}
//}

//
//func TestRemoveJob(t *testing.T) {
//	type fields struct {
//		schedule   *gocron.Scheduler
//		backupJobs map[string]model.Job
//	}
//	type args struct {
//		tag string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		mock    func(repo *mocks.Repository, tag string)
//		wantErr bool
//	}{
//		{
//			name: "successful_dont_exist_backup",
//			fields: fields{
//				schedule:   gocron.NewScheduler(time.UTC),
//				backupJobs: map[string]model.Job{},
//			},
//			args: args{
//				tag: "test",
//			},
//			mock: func(repo *mocks.Repository, tag string) {
//				repo.On("RemoveJob", mock.Anything, tag).Return(nil)
//			},
//			wantErr: false,
//		},
//		{
//			name: "successful_exist_backup",
//			fields: fields{
//				schedule:   gocron.NewScheduler(time.UTC),
//				backupJobs: map[string]model.Job{"test": {}},
//			},
//			args: args{
//				tag: "test",
//			},
//			mock: func(repo *mocks.Repository, tag string) {
//				repo.On("RemoveJob", mock.Anything, tag).Return(nil)
//			},
//			wantErr: false,
//		},
//		{
//			name: "error_repository_side",
//			fields: fields{
//				schedule:   gocron.NewScheduler(time.UTC),
//				backupJobs: map[string]model.Job{"test": {}},
//			},
//			args: args{
//				tag: "test",
//			},
//			mock: func(repo *mocks.Repository, tag string) {
//				repo.On("RemoveJob", mock.Anything, tag).Return(fmt.Errorf("repository error"))
//			},
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			repo := mocks.NewRepository(t)
//			tt.mock(repo, tt.args.tag)
//			c := &cronger{
//				schedule:   tt.fields.schedule,
//				repo:       repo,
//				backupJobs: tt.fields.backupJobs,
//			}
//			c.StartAsync()
//			if _, err := c.schedule.Cron("* * * * *").Tag(tt.args.tag).Do(func() {}); err != nil {
//				assert.Error(t, err)
//			}
//			err := c.RemoveJob(tt.args.tag)
//			if tt.wantErr {
//				assert.Error(t, err)
//			}
//		})
//	}
//}
//
//func TestJobs(t *testing.T) {
//	tests := []struct {
//		name    string
//		mock    func(repo *mocks.Repository)
//		want    []model.Job
//		wantErr bool
//	}{
//		{
//			name: "successful",
//			mock: func(repo *mocks.Repository) {
//				repo.On("Jobs", mock.Anything).Return([]model.Job{{Tag: "test1"}, {Tag: "test2"}}, nil)
//			},
//			want:    []model.Job{{Tag: "test1"}, {Tag: "test2"}},
//			wantErr: false,
//		},
//		{
//			name: "error_repository",
//			mock: func(repo *mocks.Repository) {
//				repo.On("Jobs", mock.Anything).Return(nil, fmt.Errorf("error repository"))
//			},
//			want:    nil,
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			repo := mocks.NewRepository(t)
//			tt.mock(repo)
//			c := &cronger{
//				repo: repo,
//			}
//			got, err := c.Jobs()
//			if tt.wantErr {
//				assert.Error(t, err)
//				return
//			}
//			assert.Equal(t, got, tt.want)
//		})
//	}
//}
//
//func TestBackupJobs(t *testing.T) {
//	type fields struct {
//		backupJobs map[string]model.Job
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		want   []model.Job
//	}{
//		{
//			name: "successful",
//			fields: fields{
//				backupJobs: map[string]model.Job{"test1": {Tag: "test1"}, "test2": {Tag: "test2"}, "test3": {Tag: "test3"}},
//			},
//			want: []model.Job{{Tag: "test1"}, {Tag: "test2"}, {Tag: "test3"}},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &cronger{
//				backupJobs: tt.fields.backupJobs,
//			}
//			got := c.BackupJobs()
//			sort.Slice(got, func(i, j int) bool {
//				return got[i].Tag < got[j].Tag
//			})
//			assert.Equal(t, got, tt.want)
//		})
//	}
//}
//
//func TestBackup(t *testing.T) {
//	tests := []struct {
//		name    string
//		mock    func(repo *mocks.Repository)
//		wantErr bool
//	}{
//		{
//			name: "successful",
//			mock: func(repo *mocks.Repository) {
//				repo.On("BackupJobs", mock.Anything).Return([]model.Job{{Tag: "test1"}, {Tag: "test2"}}, nil)
//			},
//			wantErr: false,
//		},
//		{
//			name: "error_repository",
//			mock: func(repo *mocks.Repository) {
//				repo.On("BackupJobs", mock.Anything).Return(nil, fmt.Errorf("error repository"))
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			repo := mocks.NewRepository(t)
//			tt.mock(repo)
//			c := &cronger{
//				repo: repo,
//			}
//			err := c.backup()
//			if tt.wantErr {
//				assert.Error(t, err)
//				return
//			}
//		})
//	}
//}

package cronger

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *cronger
		wantErr error
	}{
		{
			name: "error_if_failed_check_driver",
			args: args{
				cfg: &Config{
					TypeClient: 22,
					Client:     nil,
					Loc:        time.UTC,
				},
			},
			want:    nil,
			wantErr: errors.New("check driver error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg)
			assert.Error(t, tt.wantErr, err)
			fmt.Println(err)
			assert.Equal(t, tt.want, got)
		})
	}
}

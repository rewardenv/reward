package core_test

import (
	"testing"

	reward "github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
)

func TestRunDockerComposeCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		args             []string
		suppressOsStdOut bool
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Return version string",
			args{
				[]string{"version", "--short"},
				true,
			},
			"1",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()
				got, err := reward.RunDockerComposeCommand(tt.args.args, tt.args.suppressOsStdOut)

				log.Printf("%T, %v", got, got)
				if (err != nil) != tt.wantErr {
					t.Errorf("RunDockerComposeCommand() error = %v, wantErr %v", err, tt.wantErr)

					return
				}

				if got[:1] != tt.want {
					t.Errorf("RunDockerComposeCommand() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

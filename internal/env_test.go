package internal_test

import (
	"path/filepath"
	"testing"

	. "reward/internal"

	log "github.com/sirupsen/logrus"
)

func initEnvTests() {
}

func TestEnvInitCmd(t *testing.T) { //nolint:paralleltest
	type args []string

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"Create default magento2 environment",
			args{"magento2-project-default"},
			true,
			false,
		},
		{
			"Create magento1 environment",
			args{"magento-project", "magento1"},
			true,
			false,
		},
		{
			"Create magento 2 environment",
			args{"magento2-project", "magento2"},
			true,
			false,
		},
		{
			"Returns error for wrong environment type",
			args{"want-error-project", "test"},
			true,
			true,
		},
	}

	for _, tt := range tests { //nolint:paralleltest
		tt := tt
		errorOnInit := false

		t.Run(tt.name, func(t *testing.T) {
			// t.Helper()
			_ = AFS.Remove(filepath.Join(GetCwd(), ".env"))
			if err := EnvInitCmd(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("EnvInitCmd() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				errorOnInit = true

				return
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if (errorOnInit == true) == tt.wantErr {
				return
			}
			var got bool
			var err error

			// content, _ := AFS.ReadFile(path.Join(GetCwd(), ".env"))
			// log.Println(string(content))

			if len(tt.args) > 1 {
				got, err = CheckRegexInFile(tt.args[1], filepath.Join(GetCwd(), ".env"))
			} else {
				got, err = CheckRegexInFile("magento2", filepath.Join(GetCwd(), ".env"))
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckRegexInFile() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			log.Println(err)
			if got != tt.want {
				t.Errorf("CheckRegexInFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

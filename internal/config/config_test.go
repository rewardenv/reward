package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/rewardenv/reward/pkg/util"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) SetupTest() {
	FS = &afero.Afero{Fs: afero.NewMemMapFs()}
	util.FS = FS
	f, _ := FS.Create(".env")

	defer f.Close()
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) TestConfigMagentoVersion() {
	config := New("reward", "0.0.1").Init()

	type fields struct {
		composerJSON string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "empty composer.json",
			fields:  fields{},
			want:    "2.4.8-p1",
			wantErr: false,
		},
		{
			name: "composer.json with invalid json",
			fields: fields{
				composerJSON: `{]`,
			},
			want:    "2.4.8-p1",
			wantErr: false,
		},
		{
			name: "composer.json with valid version (old format)",
			fields: fields{
				composerJSON: `
{
  "name": "magento/magento2ce",
  "version": "2.3.8-p9"
}`,
			},
			want:    "2.3.8-p9",
			wantErr: false,
		},
		{
			name: "composer.json with valid version constraint (old format)",
			fields: fields{
				composerJSON: `
{
  "name": "magento/magento2ce",
  "version": ">=2.4.6 <2.4.7"
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
		{
			name: "composer.json with valid version constraint (old format)",
			fields: fields{
				composerJSON: `
{
  "name": "magento/magento2ce",
  "version": "dev-master"
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
		{
			name: "composer.json with invalid version (old format)",
			fields: fields{
				composerJSON: `
{
  "name": "magento/magento2ce",
  "version": "invalid version"
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
		{
			name: "composer.json with valid version (new format)",
			fields: fields{
				composerJSON: `
{
  "name": "magento/project-community-edition",
  "version": "2.4.5-p1",
  "require": {
    "magento/product-community-edition": "2.4.4-p1"
  }
}`,
			},
			want:    "2.4.4-p1",
			wantErr: false,
		},
		{
			name: "composer.json with valid constraint",
			fields: fields{
				composerJSON: `
{
  "name": "magento/project-community-edition",
  "require": {
    "magento/product-community-edition": ">=2.4.6 <2.4.7"
  }
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
		{
			name: "composer.json with invalid version",
			fields: fields{
				composerJSON: `
{
  "name": "magento/project-community-edition",
  "require": {
    "magento/product-community-edition": "invalid version"
  }
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
		{
			name: "composer.json with valid version for cloud metapackage",
			fields: fields{
				composerJSON: `
{
  "name": "magento/project-enterprise-edition",
  "require": {
    "magento/magento-cloud-metapackage": "2.4.5-p8"
  }
}`,
			},
			want:    "2.4.5-p8",
			wantErr: false,
		},
		{
			name: "composer.json with invalid version for cloud metapackage",
			fields: fields{
				composerJSON: `
{
  "name": "magento/project-enterprise-edition",
  "require": {
    "magento/magento-cloud-metapackage": "invalid version"
  }
}`,
			},
			want:    "0.0.0+undefined",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			if tt.fields.composerJSON != "" {
				_ = FS.WriteFile("composer.json", []byte(tt.fields.composerJSON), os.FileMode(0o644))
			}

			v, err := config.MagentoVersion()
			if tt.wantErr {
				assert.Error(suite.T(), err)

				return
			}

			got := v.String()
			assert.Equal(suite.T(), tt.want, got)
		})
	}
}

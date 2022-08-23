package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_defaultShellContainer(t *testing.T) {
	t.Parallel()

	type args struct {
		envType string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "magento2",
			args: args{
				envType: "magento2",
			},
			want: "php-fpm",
		},
		{
			name: "pwa-studio",
			args: args{
				envType: "pwa-studio",
			},
			want: "node",
		},
		{
			name: "default",
			args: args{
				envType: "default",
			},
			want: "php-fpm",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				assert.Equalf(
					t, tt.want, defaultShellContainer(tt.args.envType), "defaultShellContainer(%v)",
					tt.args.envType,
				)
			},
		)
	}
}

func Test_defaultShellCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		containerName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "php-fpm",
			args: args{
				containerName: "php-fpm",
			},
			want: "bash",
		},
		{
			name: "node",
			args: args{
				containerName: "node",
			},
			want: "sh",
		},
		{
			name: "default",
			args: args{
				containerName: "default",
			},
			want: "sh",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, defaultShellCommand(tt.args.containerName),
					"defaultShellCommand(%v)", tt.args.containerName,
				)
			},
		)
	}
}

func Test_defaultShellUser(t *testing.T) {
	t.Parallel()

	type args struct {
		containerName string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "php-fpm",
			args: args{
				containerName: "php-fpm",
			},
			want: "www-data",
		},
		{
			name: "node",
			args: args{
				containerName: "node",
			},
			want: "node",
		},
		{
			name: "nginx",
			args: args{
				containerName: "nginx",
			},
			want: "root",
		},
		{
			name: "default",
			args: args{
				containerName: "default",
			},
			want: "root",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, defaultShellUser(tt.args.containerName), "defaultShellUser(%v)",
					tt.args.containerName,
				)
			},
		)
	}
}

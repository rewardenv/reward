package templates

import (
	"bytes"
	"container/list"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	"github.com/spf13/viper"
)

func Test_isEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		given interface{}
		want  bool
	}{
		// nil / invalid
		{name: "nil returns false", given: nil, want: false},

		// bool
		{name: "bool true", given: true, want: true},
		{name: "bool false", given: false, want: false},

		// string
		{name: "string true", given: "true", want: true},
		{name: "string TRUE", given: "TRUE", want: true},
		{name: "string True", given: "True", want: true},
		{name: "string 1", given: "1", want: true},
		{name: "string false", given: "false", want: false},
		{name: "string 0", given: "0", want: false},
		{name: "string empty", given: "", want: false},
		{name: "string arbitrary", given: "yes", want: false},

		// int
		{name: "int 1", given: 1, want: true},
		{name: "int 0", given: 0, want: false},
		{name: "int -1", given: -1, want: false},
		{name: "int 2", given: 2, want: false},
		{name: "int64 1", given: int64(1), want: true},
		{name: "int64 0", given: int64(0), want: false},

		// unsupported types
		{name: "float64", given: 1.0, want: false},
		{name: "slice", given: []string{"true"}, want: false},
		{name: "map", given: map[string]string{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isEnabled(tt.given); got != tt.want {
				t.Errorf("isEnabled(%v) = %v, want %v", tt.given, got, tt.want)
			}
		})
	}
}

func Test_isEnabledOr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		defaultVal bool
		data       map[string]interface{}
		key        string
		want       bool
	}{
		// key exists
		{
			name:       "key exists with bool true",
			defaultVal: false,
			data:       map[string]interface{}{"feature": true},
			key:        "feature",
			want:       true,
		},
		{
			name:       "key exists with bool false",
			defaultVal: true,
			data:       map[string]interface{}{"feature": false},
			key:        "feature",
			want:       false,
		},
		{
			name:       "key exists with string true",
			defaultVal: false,
			data:       map[string]interface{}{"feature": "true"},
			key:        "feature",
			want:       true,
		},
		{
			name:       "key exists with string 1",
			defaultVal: false,
			data:       map[string]interface{}{"feature": "1"},
			key:        "feature",
			want:       true,
		},
		{
			name:       "key exists with string false",
			defaultVal: true,
			data:       map[string]interface{}{"feature": "false"},
			key:        "feature",
			want:       false,
		},
		{
			name:       "key exists with int 1",
			defaultVal: false,
			data:       map[string]interface{}{"feature": 1},
			key:        "feature",
			want:       true,
		},
		{
			name:       "key exists with int 0",
			defaultVal: true,
			data:       map[string]interface{}{"feature": 0},
			key:        "feature",
			want:       false,
		},
		{
			name:       "key exists with nil value",
			defaultVal: true,
			data:       map[string]interface{}{"feature": nil},
			key:        "feature",
			want:       false,
		},

		// key missing - default used
		{
			name:       "key missing default false",
			defaultVal: false,
			data:       map[string]interface{}{"other": true},
			key:        "feature",
			want:       false,
		},
		{
			name:       "key missing default true",
			defaultVal: true,
			data:       map[string]interface{}{"other": true},
			key:        "feature",
			want:       true,
		},
		{
			name:       "empty map default false",
			defaultVal: false,
			data:       map[string]interface{}{},
			key:        "feature",
			want:       false,
		},
		{
			name:       "empty map default true",
			defaultVal: true,
			data:       map[string]interface{}{},
			key:        "feature",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isEnabledOr(tt.defaultVal, tt.data, tt.key); got != tt.want {
				t.Errorf("isEnabledOr(%v, data, %q) = %v, want %v",
					tt.defaultVal, tt.key, got, tt.want)
			}
		})
	}
}

func TestParseKV(t *testing.T) {
	type args struct {
		kvStr string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Test empty string",
			args: args{
				kvStr: "",
			},
			want: map[string]string{},
		},
		{
			name: "Test malformed string",
			args: args{
				kvStr: "keyword1",
			},
			want: map[string]string{},
		},
		{
			name: "Test key with no value",
			args: args{
				kvStr: "keyword1=",
			},
			want: map[string]string{"keyword1": ""},
		},
		{
			name: "TestParseKV 1",
			args: args{
				kvStr: "keyword1=value1,keyword2=value2,keyword3=value3,value4,value5,keyword4=value6",
			},
			want: map[string]string{
				"keyword1": "value1",
				"keyword2": "value2",
				"keyword3": "value3,value4,value5",
				"keyword4": "value6",
			},
		},
		{
			name: "TestParseKV 2",
			args: args{
				kvStr: "keyword1=value1,keyword2=value2,keyword3=value3,value4,value5:value6,value7,value8,keyword4=value9",
			},
			want: map[string]string{
				"keyword1": "value1",
				"keyword2": "value2",
				"keyword3": "value3,value4,value5:value6,value7,value8",
				"keyword4": "value9",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseKV(tt.args.kvStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setupViperDefaults(overrides map[string]interface{}) {
	viper.Reset()
	viper.Set("app_name", "reward")
	viper.SetDefault("reward_ssl_dir", "/tmp/ssl")
	viper.SetDefault("reward_composer_dir", "/tmp/composer")
	viper.SetDefault("reward_ssh_dir", "/tmp/ssh")
	viper.SetDefault("reward_env_name", "testproject")
	viper.SetDefault("reward_env_type", "magento2")
	viper.SetDefault("reward_service_domain", "reward.test")
	viper.SetDefault("traefik_domain", "testproject.test")
	viper.SetDefault("traefik_subdomain", "app")
	viper.SetDefault("traefik_address", "0.0.0.0")
	viper.SetDefault("reward_shared_composer", true)
	viper.SetDefault("reward_docker_image_repo", "docker.io/rewardenv")
	viper.SetDefault("reward_single_web_container", false)
	viper.SetDefault("reward_resolve_domain_to_traefik", true)
	viper.SetDefault("reward_traefik_allow_http", false)
	viper.SetDefault("reward_restart_policy", "")
	viper.SetDefault("php_version", "8.2")
	viper.SetDefault("reward_svc_php_variant", "")
	viper.SetDefault("reward_svc_php_debug_variant", "")
	viper.SetDefault("mariadb_version", "10.6")
	viper.SetDefault("database_executable", "mysqld")
	viper.SetDefault("nginx_version", "1.24")
	viper.SetDefault("redis_version", "7.2")
	viper.SetDefault("valkey_version", "8.0")
	viper.SetDefault("elasticsearch_version", "7.17")
	viper.SetDefault("opensearch_version", "2.12")
	viper.SetDefault("varnish_version", "7.0")
	viper.SetDefault("rabbitmq_version", "3.12")
	viper.SetDefault("node_version", "18")
	viper.SetDefault("magepack_version", "latest")
	viper.SetDefault("reward_spx", false)
	viper.SetDefault("reward_mercure", false)
	viper.SetDefault("mysql_expose", false)
	viper.SetDefault("redis_expose", false)
	viper.SetDefault("valkey_expose", false)
	viper.SetDefault("elasticsearch_expose", false)
	viper.SetDefault("opensearch_expose", false)
	viper.SetDefault("rabbitmq_expose", false)
	viper.SetDefault("reward_traefik_custom_headers", "")
	viper.SetDefault("reward_http_proxy_ports", "")
	viper.SetDefault("reward_https_proxy_ports", "")
	viper.SetDefault("mercure_server_name", ":80")
	viper.SetDefault("mercure_publisher_jwt_alg", "HS256")
	viper.SetDefault("mercure_publisher_jwt_key", "testkey")
	viper.SetDefault("mercure_subscriber_jwt_alg", "HS256")
	viper.SetDefault("mercure_subscriber_jwt_key", "testkey")
	viper.SetDefault("mercure_extra_directives", "")

	for k, v := range overrides {
		viper.Set(k, v)
	}
}

func loadAndExecuteTemplates(t *testing.T, envType string, services []string) {
	t.Helper()
	client := New()
	tpl := template.New("root").Funcs(funcMap())
	templateList := list.New()

	err := client.AppendEnvironmentTemplates(tpl, templateList, "networks", envType)
	if err != nil {
		t.Fatalf("failed to load networks template: %v", err)
	}

	for _, svc := range services {
		if err := client.AppendEnvironmentTemplates(tpl, templateList, svc, envType); err != nil {
			t.Fatalf("failed to load %s template: %v", svc, err)
		}
	}

	if err := client.AppendEnvironmentTemplates(tpl, templateList, envType, envType); err != nil {
		t.Fatalf("failed to load %s env template: %v", envType, err)
	}

	var bs bytes.Buffer

	for e := templateList.Front(); e != nil; e = e.Next() {
		bs.Reset()

		tplName := fmt.Sprint(e.Value)
		if err := client.ExecuteTemplate(tpl.Lookup(tplName), &bs); err != nil {
			t.Fatalf("failed to execute template %s: %v", tplName, err)
		}
	}
}

func Test_TemplateExecution(t *testing.T) {
	envServices := map[string][]string{
		"magento1":    {"php-fpm", "nginx", "db", "redis"},
		"magento2":    {"php-fpm", "nginx", "db", "redis", "elasticsearch"},
		"laravel":     {"php-fpm", "nginx", "db", "redis"},
		"shopware":    {"php-fpm", "nginx", "db", "redis", "elasticsearch"},
		"wordpress":   {"php-fpm", "nginx", "db", "redis"},
		"symfony":     {"php-fpm", "nginx", "db", "redis"},
		"generic-php": {"php-fpm", "nginx", "db", "redis"},
		"pwa-studio":  {"node"},
		"local":       {},
	}

	type featureFlag struct {
		name      string
		overrides map[string]interface{}
	}

	flags := []featureFlag{
		{name: "defaults", overrides: nil},
		{name: "single_web_container", overrides: map[string]interface{}{
			"reward_single_web_container": true,
		}},
		{name: "mysql_expose", overrides: map[string]interface{}{
			"mysql_expose": true,
		}},
		{name: "redis_expose", overrides: map[string]interface{}{
			"redis_expose": true,
		}},
		{name: "mercure", overrides: map[string]interface{}{
			"reward_mercure": true,
		}},
		{name: "spx", overrides: map[string]interface{}{
			"reward_spx": true,
		}},
		{name: "traefik_allow_http", overrides: map[string]interface{}{
			"reward_traefik_allow_http": true,
		}},
		{name: "traefik_custom_headers", overrides: map[string]interface{}{
			"reward_traefik_custom_headers": "X-Test=value1",
		}},
		{name: "http_proxy_ports", overrides: map[string]interface{}{
			"reward_http_proxy_ports": "8080",
		}},
		{name: "https_proxy_ports", overrides: map[string]interface{}{
			"reward_https_proxy_ports": "8443",
		}},
		{name: "valkey_expose", overrides: map[string]interface{}{
			"valkey_expose": true,
		}},
		{name: "elasticsearch_expose", overrides: map[string]interface{}{
			"elasticsearch_expose": true,
		}},
		{name: "opensearch_expose", overrides: map[string]interface{}{
			"opensearch_expose": true,
		}},
		{name: "rabbitmq_expose", overrides: map[string]interface{}{
			"rabbitmq_expose": true,
		}},
		{name: "single_web_container_with_http_proxy", overrides: map[string]interface{}{
			"reward_single_web_container": true,
			"reward_http_proxy_ports":     "8080",
			"reward_https_proxy_ports":    "8443",
		}},
		{name: "single_web_container_with_custom_headers", overrides: map[string]interface{}{
			"reward_single_web_container":   true,
			"reward_traefik_custom_headers": "X-Test=value1,X-Other=value2",
			"reward_traefik_allow_http":     true,
		}},
		{name: "all_expose_ports", overrides: map[string]interface{}{
			"mysql_expose":         true,
			"redis_expose":         true,
			"elasticsearch_expose": true,
			"rabbitmq_expose":      true,
		}},
	}

	for envType, services := range envServices {
		for _, flag := range flags {
			testName := fmt.Sprintf("%s/%s", envType, flag.name)
			t.Run(testName, func(t *testing.T) {
				setupViperDefaults(flag.overrides)
				viper.Set("reward_env_type", envType)
				svcList := make([]string, 0, len(services)+1)
				svcList = append(svcList, services...)

				if flag.name == "mercure" {
					svcList = append(svcList, "mercure")
				}

				loadAndExecuteTemplates(t, envType, svcList)
			})
		}
	}

	type extraEnv struct {
		envType   string
		testLabel string
		services  []string
	}
	envServicesExtra := []extraEnv{
		{"magento2", "magento2-opensearch", []string{"php-fpm", "nginx", "db", "opensearch"}},
		{"shopware", "shopware-valkey", []string{"php-fpm", "nginx", "db", "valkey"}},
		{"magento2", "magento2-varnish", []string{"php-fpm", "nginx", "db", "redis", "elasticsearch", "varnish"}},
		{"magento2", "magento2-rabbitmq", []string{"php-fpm", "nginx", "db", "redis", "elasticsearch", "rabbitmq"}},
	}

	for _, extra := range envServicesExtra {
		envType := extra.envType
		for _, flag := range flags {
			testName := fmt.Sprintf("%s/%s", extra.testLabel, flag.name)
			t.Run(testName, func(t *testing.T) {
				setupViperDefaults(flag.overrides)
				viper.Set("reward_env_type", envType)

				svcList := make([]string, 0, len(extra.services)+1)
				svcList = append(svcList, extra.services...)

				if flag.name == "mercure" {
					svcList = append(svcList, "mercure")
				}

				loadAndExecuteTemplates(t, envType, svcList)
			})
		}
	}
}

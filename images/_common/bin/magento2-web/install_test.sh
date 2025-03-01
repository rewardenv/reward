#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/install.sh"
}

function test_magento_args_install_only() {
  # Default
  local ARGS=("")
  setup

  magento_args_install_only

  assert_array_contains "--base-url=http://magento.test" "${ARGS[@]}"
  assert_array_contains "--base-url-secure=https://magento.test" "${ARGS[@]}"
  assert_array_contains "--use-secure=1" "${ARGS[@]}"
  assert_array_contains "--use-secure-admin=1" "${ARGS[@]}"
  assert_array_contains "--use-rewrites=1" "${ARGS[@]}"
}

function test_magento_args_install_only_disabled_https() {
  # Disable HTTPS and rewrites
  local ARGS=("")
  local MAGENTO_ENABLE_HTTPS="false"
  local MAGENTO_ENABLE_ADMIN_HTTPS="false"
  local MAGENTO_USE_REWRITES="false"
  setup

  magento_args_install_only
  assert_array_contains "--base-url=http://magento.test" "${ARGS[@]}"
  assert_array_contains "--base-url-secure=https://magento.test" "${ARGS[@]}"
  assert_array_not_contains "--use-secure=1" "${ARGS[@]}"
  assert_array_not_contains "--use-secure-admin=1" "${ARGS[@]}"
  assert_array_not_contains "--use-rewrites=1" "${ARGS[@]}"
}

function test_magento_args_defaults() {
  # Default
  local ARGS=("")
  setup

  magento_args_defaults
  assert_array_contains "--key=12345678901234567890123456789012" "${ARGS[@]}"
  assert_array_contains "--backend-frontname=admin" "${ARGS[@]}"
}

function test_magento_args_custom_values() {
  # Custom values
  local ARGS=("")
  local MAGENTO_KEY="1234"
  local MAGENTO_ADMIN_URL_PREFIX="admin-1234"
  setup

  magento_args_defaults
  assert_array_contains "--key=1234" "${ARGS[@]}"
  assert_array_contains "--backend-frontname=admin-1234" "${ARGS[@]}"
}

function test_magento_args_db_defaults() {
  # Default
  local ARGS=("")
  setup

  magento_args_db

  assert_array_contains "--db-host=db" "${ARGS[@]}"
  assert_array_contains "--db-name=magento" "${ARGS[@]}"
  assert_array_contains "--db-user=magento" "${ARGS[@]}"
  assert_array_contains "--db-password=magento" "${ARGS[@]}"
}

function test_magento_args_db_custom() {
  # Custom values
  local ARGS=("")
  local MAGENTO_DATABASE_HOST="localhost"
  local MAGENTO_DATABASE_NAME="magento2"
  local MAGENTO_DATABASE_USER="root"
  local MAGENTO_DATABASE_PASSWORD="root"
  setup

  magento_args_db

  assert_array_contains "--db-host=localhost" "${ARGS[@]}"
  assert_array_contains "--db-name=magento2" "${ARGS[@]}"
  assert_array_contains "--db-user=root" "${ARGS[@]}"
  assert_array_contains "--db-password=root" "${ARGS[@]}"
}

function test_magento_args_redis_default() {
  # Default
  local ARGS=("")
  setup

  magento_args_redis
  assert_array_contains "--session-save=files" "${ARGS[@]}"
}

function test_magento_args_redis_enabled() {
  # Redis is enabled
  local ARGS=("")
  local MAGENTO_REDIS_ENABLED="true"
  setup

  magento_args_redis

  assert_array_contains "--session-save=redis" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-host=redis" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-port=6379" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-db=2" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-max-concurrency=20" "${ARGS[@]}"
  assert_array_contains "--cache-backend=redis" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-server=redis" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-port=6379" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-db=0" "${ARGS[@]}"
  assert_array_contains "--page-cache=redis" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-server=redis" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-port=6379" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-db=1" "${ARGS[@]}"
}

function test_magento_args_redis_password() {
  # Redis is enabled
  local ARGS=("")
  local MAGENTO_REDIS_ENABLED="true"
  local MAGENTO_REDIS_PASSWORD="password"
  setup

  magento_args_redis

  assert_array_contains "--session-save-redis-password=password" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-password=password" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-password=password" "${ARGS[@]}"
}

function test_magento_args_redis_custom() {
  # Custom values
  local ARGS=("")
  local MAGENTO_REDIS_ENABLED="true"
  local MAGENTO_SESSION_SAVE_REDIS_HOST="localhost1"
  local MAGENTO_SESSION_SAVE_REDIS_PORT="6380"
  local MAGENTO_SESSION_SAVE_REDIS_DB="1"
  local MAGENTO_SESSION_SAVE_REDIS_PASSWORD="password1"
  local MAGENTO_CACHE_BACKEND_REDIS_SERVER="localhost2"
  local MAGENTO_CACHE_BACKEND_REDIS_PORT="6381"
  local MAGENTO_CACHE_BACKEND_REDIS_DB="2"
  local MAGENTO_CACHE_BACKEND_REDIS_PASSWORD="password2"
  local MAGENTO_PAGE_CACHE_REDIS_SERVER="localhost3"
  local MAGENTO_PAGE_CACHE_REDIS_PORT="6382"
  local MAGENTO_PAGE_CACHE_REDIS_DB="3"
  local MAGENTO_PAGE_CACHE_REDIS_PASSWORD="password3"
  setup

  magento_args_redis

  assert_array_contains "--session-save-redis-host=localhost1" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-port=6380" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-db=1" "${ARGS[@]}"
  assert_array_contains "--session-save-redis-password=password1" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-server=localhost2" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-port=6381" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-db=2" "${ARGS[@]}"
  assert_array_contains "--cache-backend-redis-password=password2" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-server=localhost3" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-port=6382" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-db=3" "${ARGS[@]}"
  assert_array_contains "--page-cache-redis-password=password3" "${ARGS[@]}"
}

function test_magento_args_varnish_default() {
  # Default
  local ARGS=("")
  setup

  magento_args_varnish
  assert_equals "" "${ARGS[@]}"
}

function test_magento_args_varnish_enabled() {
  # Varnish is disabled
  local ARGS=("")
  local MAGENTO_VARNISH_ENABLED="true"
  setup

  magento_args_varnish
  assert_array_contains "--http-cache-hosts=varnish:80" "${ARGS[@]}"
}

function test_magento_args_varnish_custom() {
  # Varnish is disabled
  # Custom values
  local ARGS=("")
  local MAGENTO_VARNISH_ENABLED="true"
  local MAGENTO_VARNISH_HOST="localhost"
  local MAGENTO_VARNISH_PORT="8080"
  setup

  magento_args_varnish
  assert_array_contains "--http-cache-hosts=localhost:8080" "${ARGS[@]}"
}

function test_magento_args_search_default() {
  # Default behaviour
  # Elasticsearch enabled
  local ARGS=("")
  setup

  spy search_configured
  magento_args_search
  assert_array_contains "--search-engine=elasticsearch7" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-host=elasticsearch" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-port=9200" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-index-prefix=magento2" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-enable-auth=0" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-timeout=15" "${ARGS[@]}"
  assert_have_been_called search_configured
}

function test_magento_args_search_default_magento_23() {
  local ARGS=("")
  local MAGENTO_VERSION="2.3"
  setup

  spy search_configured
  magento_args_search
  assert_array_not_contains "--search-engine" "${ARGS[@]}"
}

function test_magento_args_search_elasticsearch_enabled_magento_23() {
  # Elasticsearch is enabled and magento version is 2.3
  local ARGS=("")
  local MAGENTO_ELASTICSEARCH_ENABLED="true"
  local MAGENTO_VERSION="2.3"
  setup

  spy search_configured
  magento_args_search
  assert_array_contains "--search-engine=elasticsearch7" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-host=elasticsearch" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-port=9200" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-index-prefix=magento2" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-enable-auth=0" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-timeout=15" "${ARGS[@]}"
  assert_have_been_called search_configured
}

function test_magento_args_search_elasticsearch_enabled_default() {
  # Elasticsearch enabled
  local ARGS=("")
  local MAGENTO_ELASTICSEARCH_ENABLED="true"
  setup

  spy search_configured
  magento_args_search
  assert_array_contains "--search-engine=elasticsearch7" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-host=elasticsearch" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-port=9200" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-index-prefix=magento2" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-enable-auth=0" "${ARGS[@]}"
  assert_array_contains "--elasticsearch-timeout=15" "${ARGS[@]}"
  assert_have_been_called search_configured
}

function test_magento_args_search_opensearch() {
  # Opensearch enabled
  local ARGS=("")
  local MAGENTO_OPENSEARCH_ENABLED="true"
  setup

  spy search_configured
  magento_args_search
  assert_array_contains "--search-engine=opensearch" "${ARGS[@]}"
  assert_array_contains "--opensearch-host=opensearch" "${ARGS[@]}"
  assert_array_contains "--opensearch-port=9200" "${ARGS[@]}"
  assert_array_contains "--opensearch-index-prefix=magento2" "${ARGS[@]}"
  assert_array_contains "--opensearch-enable-auth=0" "${ARGS[@]}"
  assert_array_contains "--opensearch-timeout=15" "${ARGS[@]}"
  assert_have_been_called search_configured
}

function test_magento_args_rabbitmq_default() {
  # Default
  local ARGS=("")
  setup

  magento_args_rabbitmq
  assert_equals "" "${ARGS[@]}"
}

function test_magento_args_rabbitmq_enabled() {
  # RabbitMQ is enabled
  local ARGS=("")
  local MAGENTO_RABBITMQ_ENABLED="true"
  setup

  magento_args_rabbitmq
  assert_array_contains "--amqp-host=rabbitmq" "${ARGS[@]}"
  assert_array_contains "--amqp-port=5672" "${ARGS[@]}"
  assert_array_contains "--amqp-user=guest" "${ARGS[@]}"
  assert_array_contains "--amqp-password=guest" "${ARGS[@]}"
  assert_array_contains "--amqp-virtualhost=/" "${ARGS[@]}"
  assert_array_contains "--consumers-wait-for-messages=0" "${ARGS[@]}"
}

function test_magento_args_rabbitmq_enabled_magento23() {
  # RabbitMQ is enabled
  local ARGS=("")
  local MAGENTO_RABBITMQ_ENABLED="true"
  local MAGENTO_VERSION="2.3"
  setup

  magento_args_rabbitmq
  assert_array_not_contains "--consumers-wait-for-messages" "${ARGS[@]}"
}

function test_magento_args_sample_data_default() {
  # Default
  local ARGS=("")
  setup

  magento_args_sample_data
  assert_equals "" "${ARGS[@]}"
}

function test_magento_args_sample_data_enabled() {
  # Sample data is enabled
  local ARGS=("")
  local MAGENTO_DEPLOY_SAMPLE_DATA="true"
  setup

  magento_args_sample_data
  assert_array_contains "--use-sample-data" "${ARGS[@]}"
}

function test_magento_args_extra_default() {
  # Default
  local ARGS=("")
  setup

  magento_args_extra
  assert_equals "" "${ARGS[@]}"
}

function test_magento_args_extra_custom() {
  # Custom values
  local ARGS=("")
  local MAGENTO_EXTRA_INSTALL_ARGS="--magento-mode=developer"
  setup

  magento_args_extra
  assert_array_contains "--magento-mode=developer" "${ARGS[@]}"
}

function test_magento_setup_install_default() {
  setup

  spy magento_args_install_only
  spy magento_args_defaults
  spy magento_args_db
  spy magento_args_redis
  spy magento_args_varnish
  spy magento_args_search
  spy magento_args_rabbitmq
  spy magento_args_sample_data
  spy magento_args_extra
  spy magento
  spy magento_configure_search

  magento_setup_install

  assert_have_been_called magento_args_install_only
  assert_have_been_called magento_args_defaults
  assert_have_been_called magento_args_db
  assert_have_been_called magento_args_redis
  assert_have_been_called magento_args_varnish
  assert_have_been_called magento_args_search
  assert_have_been_called magento_args_rabbitmq
  assert_have_been_called magento_args_sample_data
  assert_have_been_called magento_args_extra
  assert_have_been_called magento
  assert_have_been_called magento_configure_search
}

function test_magento_setup_install_skip() {
  local MAGENTO_SKIP_INSTALL="true"
  setup

  spy log
  magento_setup_install
  assert_have_been_called_times 0 log
}

function test_magento_configure() {
  setup

  spy magento
  spy magerun
  spy magento_args_defaults
  spy magento_args_db
  spy magento_args_search
  spy magento_args_redis
  spy magento_args_varnish
  spy magento_args_rabbitmq
  spy magento_configure_search

  magento_configure
  assert_have_been_called magento_args_defaults
  assert_have_been_called magento_args_db
  assert_have_been_called_times 0 magento_args_search
  assert_have_been_called magento_args_redis
  assert_have_been_called magento_args_varnish
  assert_have_been_called magento_args_rabbitmq
  assert_have_been_called magento_configure_search

  # Test if search is configured separately
  mock magento_search_configurable true
  spy search_configured
  magento_configure
  assert_have_been_called_times 1 magento_args_search
}

function test_magento_configure_search_separately() {
  setup

  mock magento_search_configurable false
  mock magento echo
  mock magerun echo
  spy magento_args_search
  magento_configure
  assert_have_been_called_times 0 magento_args_search
}

function test_magento_app_config_import() {
  setup

  spy magento
  magento_app_config_import
  assert_have_been_called_with "app:config:import" magento
}

function test_magento_search_configurable_false() {
  setup

  mock magento <<EOF
hello
EOF
  assert_exit_code "1" "$(magento_search_configurable)"
}

function test_magento_search_configurable_true() {
  setup

  mock magento <<EOF
  --search-engine=test
EOF
  assert_exit_code "0" "$(magento_search_configurable)"
}

function test_magento_setup_di_compile_default() {
  setup

  spy magento

  # By default it should not run
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_di_compile_enabled() {
  # If MAGENTO_DI_COMPILE is true, it should run
  local MAGENTO_DI_COMPILE="true"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 1 magento
}

function test_magento_setup_di_compile_on_demand() {
  # If MAGENTO_DI_COMPILE_ON_DEMAND is true, it should run
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 1 magento
}

function test_magento_setup_di_compile_both_true() {
  # If both is true, it should run
  local MAGENTO_DI_COMPILE="true"
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 1 magento
}

function test_magento_setup_static_content_deploy() {
  # By default it should not run
  setup

  mock nproc echo 4
  spy magento

  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy_enabled() {
  # If MAGENTO_STATIC_CONTENT_DEPLOY is true, it should run
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 1 magento
  unset MAGENTO_STATIC_CONTENT_DEPLOY
}


function test_magento_setup_static_content_deploy_scd_undefined_but_scd_on_demand_true() {
  # MAGENTO_STATIC_CONTENT_DEPLOY is undefined, but MAGENTO_SCD_ON_DEMAND is true, it should run
  local MAGENTO_SCD_ON_DEMAND="true"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv" magento
}

function test_magento_setup_static_content_deploy_scd_true_and_scd_on_demand_true() {
  # MAGENTO_SCD_ON_DEMAND is true
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_SCD_ON_DEMAND="true"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv" magento
}

function test_magento_setup_static_content_deploy_magento_24() {
  # SCD args is -fv if Magento version is 2.4
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_VERSION="2.4"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv" magento
}

function test_magento_setup_static_content_deploy_magento_23() {
  # SCD args is -v if Magento version is 2.3
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_VERSION="2.3"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -v" magento
}

function test_magento_setup_static_content_deploy_languages() {
  # MAGENTO_LANGUAGES appended to the command
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_LANGUAGES="en_US de_DE"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv en_US de_DE" magento
}

function test_magento_setup_static_content_deploy_themes() {
  # MAGENTO_THEMES appended to the command
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_THEMES="Magento/blank Magento/luma"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv --theme=Magento/blank --theme=Magento/luma" magento
}

function test_magento_cache_enable() {
  setup

  spy magento
  magento_cache_enable
  assert_have_been_called_with "cache:enable" magento
}

function test_magento_reindex_default() {
  setup

  spy magento
  magento_reindex
  assert_have_been_called_times 0 magento
}

function test_magento_reindex_skip_false() {
  MAGENTO_SKIP_REINDEX=false
  setup

  spy magento
  magento_reindex
  assert_have_been_called_with "indexer:reindex" magento
}


function test_magento_setup_upgrade() {
  setup

  mock magento echo
  mock magento_upgrade_required true
  spy magento
  spy magento_maintenance_enable
  spy magento_maintenance_disable
  magento_setup_upgrade
  assert_have_been_called magento_maintenance_enable
  assert_have_been_called_with "setup:upgrade --keep-generated" magento
  assert_have_been_called magento_maintenance_disable
}

function test_magento_setup_upgrade_not_required() {
  setup

  # If upgrade is not required, it should not run
  spy magento
  spy magento_maintenance_enable
  spy magento_maintenance_disable
  mock magento_upgrade_required false
  magento_setup_upgrade
  assert_have_been_called_times 0 magento_maintenance_enable
  assert_have_been_called_times 0 magento
  assert_have_been_called_times 0 magento_maintenance_disable
}

function test_magento_setup_upgrade_skip() {
  # If MAGENTO_SKIP_UPGRADE is true, it should not run
  local MAGENTO_SKIP_UPGRADE=true
  setup

  mock magento_upgrade_required true
  spy magento
  spy magento_maintenance_enable
  spy magento_maintenance_disable
  magento_setup_upgrade
  assert_have_been_called_times 0 magento_maintenance_enable
  assert_have_been_called_times 0 magento
  assert_have_been_called_times 0 magento_maintenance_disable
}

function test_magento_deploy_mode_set_default() {
  # If MAGENTO_MODE is not set, it should not run
  setup

  spy magento
  magento_deploy_mode_set
  assert_have_been_called_times 0 magento

  # If MAGENTO_MODE is production, it should configure properly
  local MAGENTO_MODE="production"
  spy magento
  magento_deploy_mode_set
  assert_have_been_called_with "deploy:mode:set production" magento
}

function test_magento_deploy_mode_set_developer() {
  # If MAGENTO_MODE is developer, it should configure properly
  local MAGENTO_MODE="developer"
  setup

  spy magento
  magento_deploy_mode_set
  assert_have_been_called_with "deploy:mode:set developer" magento
}

function test_magento_deploy_mode_set_production() {
  # If MAGENTO_MODE is production, it should configure properly
  local MAGENTO_MODE="production"
  setup

  spy magento
  magento_deploy_mode_set
  assert_have_been_called_with "deploy:mode:set production" magento
}

function test_magento_admin_user_exists_default() {
  setup

  # Valid admin
  mock magerun <<EOF
iid,username,email,status
1,admin,admin@example.com,active
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 0 "$(magento_admin_user_exists)"

  # Invalid admin
  mock magerun <<EOF
iid,username,email,status
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 1 "$(magento_admin_user_exists)"

  # Malformed output
  mock magerun <<EOF
iid,username,email,status
admin,active
admintest,active
EOF
  assert_exit_code 1 "$(magento_admin_user_exists)"
}

function test_magento_admin_user_exists_invalid() {
  setup

  # Invalid admin
  mock magerun <<EOF
iid,username,email,status
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 1 "$(magento_admin_user_exists)"
}

function test_magento_admin_user_exists_malformed_output() {
  setup

  # Malformed output
  mock magerun <<EOF
iid,username,email,status
admin,active
admintest,active
EOF
  assert_exit_code 1 "$(magento_admin_user_exists)"
}

function test_magento_admin_user_exists_custom() {
  # Custom admin username
  local MAGENTO_USERNAME="johndoe"
  setup

  mock magerun <<EOF
iid,username,email,status
5,johndoe,johndoe@example.com,inactive
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 0 "$(magento_admin_user_exists)"
  unset MAGENTO_USERNAME
}

function test_magento_admin_user_inactive() {
  setup

  # Admin inactive
  mock magerun <<EOF
iid,username,email,status
1,admin,admin@example.com,inactive
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 0 "$(magento_admin_user_inactive)"
}

function test_magento_admin_user_inactive_malformed() {
  setup

  # Malformed output inactive
  mock magerun <<EOF
iid,username,email,status
admin,inactive
admintest,active
EOF
  assert_exit_code 1 "$(magento_admin_user_inactive)"
}

function test_magento_admin_user_inactive_is_active() {
  setup

  # Admin user active
  mock magerun <<EOF
  iid,username,email,status
  1,admin,admin@example.com,active
  21,admintest,admintest@example.com,active
  39,otheradmin,otheradmin@example.com,active
  39,user,user@example.com,active
EOF
  assert_exit_code 1 "$(magento_admin_user_inactive)"
}

function test_magento_admin_user_inactive_is_active_malformed() {
  setup

  # Malformed output active
  mock magerun <<EOF
iid,username,email,status
admin,active
admintest,active
EOF
  assert_exit_code 1 "$(magento_admin_user_inactive)"
}

function test_magento_admin_user_inactive_custom() {
  # Admin inactive custom admin username
  local MAGENTO_USERNAME="johndoe"
  setup

  mock magerun <<EOF
iid,username,email,status
5,johndoe,johndoe@example.com,inactive
21,admintest,admintest@example.com,active
39,otheradmin,otheradmin@example.com,active
39,user,user@example.com,active
EOF
  assert_exit_code 0 "$(magento_admin_user_inactive)"
}

function test_magento_admin_user_exist_active() {
  setup

  # Admin user exist and active
  mock magento_admin_user_exists true
  spy magerun
  magento_admin_user
  assert_have_been_called_times 1 magerun
}

function test_magento_admin_user_exist_inactive() {
  setup

  # Admin user exist and inactive
  mock magento_admin_user_exists true
  mock magento_admin_user_inactive true
  spy magerun
  magento_admin_user
  assert_have_been_called_times 2 magerun
}

function test_magento_admin_user_doesnt_exist() {
  setup

  # Admin user doesn't exist
  mock magento_admin_user_exists false
  spy magerun
  magento_admin_user
  assert_have_been_called_with "admin:user:create --admin-firstname=admin --admin-lastname=admin --admin-email=admin@example.com --admin-user=admin --admin-password=ASDqwe123" magerun
}

function test_magento_admin_user_custom() {
  # Custom admin user values
  local MAGENTO_FIRST_NAME="John"
  local MAGENTO_LAST_NAME="Doe"
  local MAGENTO_EMAIL="johndoe@example.com"
  local MAGENTO_USERNAME="johndoe"
  local MAGENTO_PASSWORD="johndoepw"
  setup

  mock magento_admin_user_exists false
  spy magerun
  magento_admin_user
  assert_have_been_called_with "admin:user:create --admin-firstname=John --admin-lastname=Doe --admin-email=johndoe@example.com --admin-user=johndoe --admin-password=johndoepw" magerun
}

function test_magento_disable_deploy_sample_data() {
  local MAGENTO_DEPLOY_SAMPLE_DATA="true"
  setup

  magento_disable_deploy_sample_data
  assert_equals "false" "$MAGENTO_DEPLOY_SAMPLE_DATA"
}

function test_magento_deploy_sample_data_default() {
  setup
  spy magento
  spy magento_setup_static_content_deploy

  magento_deploy_sample_data
  assert_have_been_called_times 0 magento
  assert_have_been_called_times 0 magento_setup_static_content_deploy
}

function test_magento_deploy_sample_data_enabled() {
  # If MAGENTO_DEPLOY_SAMPLE_DATA is true, it should run
  local MAGENTO_DEPLOY_SAMPLE_DATA="true"
  setup

  spy magento
  spy magento_setup_static_content_deploy
  magento_deploy_sample_data
  assert_have_been_called_times 2 magento
  assert_have_been_called magento_setup_static_content_deploy
}

function test_magento_deploy_sample_data_force() {
  # If MAGENTO_DEPLOY_SAMPLE_DATA is false, but force is true, it should run
  local MAGENTO_DEPLOY_SAMPLE_DATA="false"
  local MAGENTO_FORCE_DEPLOY_SAMPLE_DATA="true"
  setup

  spy magento
  spy magento_setup_static_content_deploy
  magento_deploy_sample_data
  assert_have_been_called_times 2 magento
  assert_have_been_called magento_setup_static_content_deploy
}

function test_magento_publish_shared_files_default() {
  setup

  # Test with a valid SHARED_CONFIG_PATH
  local SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}/app/etc"
  touch "${APP_PATH}/app/etc/env.php"
  magento_publish_shared_files
  assert_file_exists "test-data/config/app/etc/env.php"
  rm -fr "./test-data"
}

function test_magento_publish_shared_files_not_writable() {
  setup

  # Test if SHARED_CONFIG_PATH is not writable (/config by default)
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}/app/etc"
  touch "${APP_PATH}/app/etc/env.php"
  magento_publish_shared_files
  assert_file_exists "/tmp/app/etc/env.php"
  rm -fr "/tmp/app"
  rm -fr "./test-data"
}

function test_command_before_install_default() {
  # Default
  setup

  spy eval
  assert_exit_code 0 "$(command_before_install)"
}

function test_command_before_install_custom() {
  # Custom command
  local COMMAND_BEFORE_INSTALL="echo 'test'"
  setup

  spy eval
  command_before_install
  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_install_default() {
  # Default
  setup
  spy eval

  assert_exit_code 0 "$(command_before_install)"
}

function test_command_after_install_custom() {
  # Custom command
  local COMMAND_AFTER_INSTALL="echo 'test'"
  setup

  spy eval
  command_after_install
  assert_have_been_called_with "echo 'test'" eval
}

function test_bootstrap_check_default() {
  setup

  # If both are false it should not call the command_after_install
  assert_empty "$(bootstrap_check)"
}

function test_bootstrap_check_skip_bootstrap_but_command_after_install_set() {
  # If MAGENTO_SKIP_BOOTSTRAP is true, it should just run the COMMAND_AFTER_INSTALL and exit
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local MAGENTO_SKIP_BOOTSTRAP="true"
  setup

  assert_contains "test-123" "$(bootstrap_check)"

  # If SKIP_BOOTSTRAP is true, it should run the COMMAND_AFTER_INSTALL and exit
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
}

function test_bootstrap_check_both_enabled() {
  # If both are true it should run the COMMAND_AFTER_INSTALL
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local MAGENTO_SKIP_BOOTSTRAP="true"
  local SKIP_BOOTSTRAP="true"
  setup

  assert_contains "test-123" "$(bootstrap_check)"
}

function test_composer_configure_default() {
  # Default
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer
}

function test_composer_configure_magento() {
  # Test if only MAGENTO_PUBLIC_KEY is set
  local MAGENTO_PUBLIC_KEY="public"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  # Test if only MAGENTO_PRIVATE_KEY is set
  local MAGENTO_PUBLIC_KEY=""
  local MAGENTO_PRIVATE_KEY="private"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  local MAGENTO_PUBLIC_KEY="public"
  local MAGENTO_PRIVATE_KEY="private"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_github() {
  local GITHUB_USER="user"
  local GITHUB_TOKEN="token"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_bitbucket() {
  local BITBUCKET_PUBLIC_KEY="public"
  local BITBUCKET_PRIVATE_KEY="private"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_gitlab() {
  local GITLAB_TOKEN="token"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_auth() {
  local COMPOSER_AUTH="test"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer
}

function test_composer_configure_home_for_magento() {
  local APP_PATH="./test-data/app"
  setup

  mock composer echo
  composer_configure_home_for_magento
  assert_directory_exists "${APP_PATH}/var/composer_home"

  rm -fr "./test-data"
}

function test_composer_configure_plugins() {
  setup

  spy composer
  composer_configure_plugins
  assert_have_been_called_times 4 composer
}

function test_magento_is_installed_app_path() {
  # Test with app path
  local APP_PATH="./test-data/app"
  setup

  assert_exit_code 1 "$(magento_is_installed)"

  mkdir -p "${APP_PATH}/app/etc"
  touch "${APP_PATH}/app/etc/env.php"
  assert_exit_code 0 "$(magento_is_installed)"

  rm -fr "./test-data"
}

function test_magento_is_installed_shared_config_path() {
  # Test with shared config path
  local SHARED_CONFIG_PATH="./test-data/config"
  setup

  assert_exit_code 1 "$(magento_is_installed)"

  mkdir -p "${SHARED_CONFIG_PATH}/app/etc"
  touch "${SHARED_CONFIG_PATH}/app/etc/env.php"
  assert_exit_code 0 "$(magento_is_installed)"

  rm -fr "./test-data"
}

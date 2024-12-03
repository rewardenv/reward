#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/install.sh"
}

function test_command_before_install() {
  # Default
  assert_exit_code 0 "$(command_before_install)"

  # Custom command
  local COMMAND_BEFORE_INSTALL="echo 'test'"
  spy eval

  command_before_install

  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_install() {
  # Default
  assert_exit_code 0 "$(command_before_install)"

  # Custom command
  local COMMAND_AFTER_INSTALL="echo 'test'"
  spy eval

  command_after_install

  assert_have_been_called_with "echo 'test'" eval
}

function test_bootstrap_check() {
  local COMMAND_AFTER_INSTALL="echo 'test-123'"

  # If SHOPWARE_SKIP_BOOTSTRAP is true, it should just run the COMMAND_AFTER_INSTALL and exit
  local SHOPWARE_SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset SHOPWARE_SKIP_BOOTSTRAP

  # If SKIP_BOOTSTRAP is true, it should run the COMMAND_AFTER_INSTALL and exit
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset SKIP_BOOTSTRAP

  # If both are true it should run the COMMAND_AFTER_INSTALL
  local SHOPWARE_SKIP_BOOTSTRAP="true"
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset SKIP_BOOTSTRAP
  unset SHOPWARE_SKIP_BOOTSTRAP

  # If both are false it should not call the command_after_install
  assert_empty "$(bootstrap_check)"
}

function test_composer_configure() {
  # Default
  mock composer echo
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  local GITHUB_USER="user"
  local GITHUB_TOKEN="token"
  spy composer
  composer_configure
  assert_have_been_called_times 1 composer

  local BITBUCKET_PUBLIC_KEY="public"
  local BITBUCKET_PRIVATE_KEY="private"
  spy composer
  composer_configure
  assert_have_been_called_times 2 composer

  local GITLAB_TOKEN="token"
  spy composer
  composer_configure
  assert_have_been_called_times 3 composer

  local COMPOSER_AUTH="test"
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer
}

function test_shopware_version() {
  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  echo '{"packages":[{"name":"shopware/core","version":"v6.4.18.0","extra":{"branch-alias":{"dev-master":"6.4.x-dev","dev-trunk":"6.4.x-dev"}}}]}' >"${APP_PATH}/composer.lock"
  assert_equals "v6.4.18.0" "$(shopware_version)"

  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  echo '{"packages":[{"name":"shopware/core","version":"6.4.18.0","extra":{"branch-alias":{"dev-master":"6.4.x-dev","dev-trunk":"6.4.x-dev"}}}]}' >"${APP_PATH}/composer.lock"
  assert_equals "v6.4.x-dev" "$(shopware_version)"

  rm -fr './test-data'
}

function test_shopware_deployment_helper() {
  local APP_PATH="./test-data/app"
  # Deployment helper exists
  mkdir -p "${APP_PATH}/vendor/bin"
  printf '#!/bin/bash\necho from-helper:$1' >"${APP_PATH}/vendor/bin/shopware-deployment-helper"
  chmod +x "${APP_PATH}/vendor/bin/shopware-deployment-helper"
  assert_matches ".*from-helper:run$" "$(shopware_deployment_helper)"

  rm -fr './test-data'

  mock exec echo "run"
  spy exec
  assert_have_been_called_times 0 exec
}

function test_shopware_is_installed() {
  mock shopware_version echo "v6.5.18.0"
  mock console true
  assert_exit_code 0 "$(shopware_is_installed)"

  mock console false
  assert_exit_code 1 "$(shopware_is_installed)"

  mock shopware_version echo "6.6.0.0"
  mock console true
  assert_exit_code 0 "$(shopware_is_installed)"

  mock console false
  assert_exit_code 1 "$(shopware_is_installed)"
}

function test_shopware_args_defaults() {
  # Default
  local ARGS=("")
  shopware_args_defaults
  assert_array_contains "--app-env=prod" "${ARGS[@]}"
  assert_array_contains "--app-url=https://shopware.test" "${ARGS[@]}"
  assert_array_contains "--database-url=mysql://app:app@db:3306/shopware" "${ARGS[@]}"
  assert_array_contains "--cdn-strategy=physical_filename" "${ARGS[@]}"
  assert_array_contains "--mailer-url=native://default" "${ARGS[@]}"

  # Custom values
  local ARGS=("")
  local SHOPWARE_APP_ENV="dev"
  local SHOPWARE_APP_URL="http://shopware.test"
  local SHOPWARE_DATABASE_HOST="localhost"
  local SHOPWARE_DATABASE_NAME="shopware2"
  local SHOPWARE_DATABASE_USER="root"
  local SHOPWARE_DATABASE_PASSWORD="root"
  local SHOPWARE_CDN_STRATEGY="symlink"
  local SHOPWARE_MAILER_URL="smtp://localhost"
  shopware_args_defaults
  assert_array_contains "--app-env=dev" "${ARGS[@]}"
  assert_array_contains "--app-url=http://shopware.test" "${ARGS[@]}"
  assert_array_contains "--database-url=mysql://root:root@localhost:3306/shopware2" "${ARGS[@]}"
  assert_array_contains "--cdn-strategy=symlink" "${ARGS[@]}"
  assert_array_contains "--mailer-url=smtp://localhost" "${ARGS[@]}"
  unset SHOPWARE_APP_ENV
  unset SHOPWARE_APP_URL
  unset SHOPWARE_DATABASE_HOST
  unset SHOPWARE_DATABASE_NAME
  unset SHOPWARE_DATABASE_USER
  unset SHOPWARE_DATABASE_PASSWORD
  unset SHOPWARE_CDN_STRATEGY
  unset SHOPWARE_MAILER_URL

  # Custom values for SHOPWARE_HOST and SHOPWARE_SCHEME
  local ARGS=("")
  local SHOPWARE_HOST="localhost"
  local SHOPWARE_SCHEME="http"
  shopware_args_defaults
  assert_array_contains "--app-url=http://localhost" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch() {
  # Default
  local ARGS=("")
  shopware_args_elasticsearch
  assert_array_not_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_not_contains "--es-hosts" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"

  # Custom values - elasticsearch enabled
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=elasticsearch:9200" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_ELASTICSEARCH_ENABLED

  # Custom values - elasticsearch indexing disabled
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_INDEXING_ENABLED="false"
  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=elasticsearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_ELASTICSEARCH_INDEXING_ENABLED
  unset SHOPWARE_ELASTICSEARCH_ENABLED

  # Custom values - custom elasticsearch host
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_HOST="localhost"
  local SHOPWARE_ELASTICSEARCH_PORT="9201"
  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost:9201" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_ELASTICSEARCH_HOST
  unset SHOPWARE_ELASTICSEARCH_PORT
  unset SHOPWARE_ELASTICSEARCH_ENABLED

  # Custom values - custom elasticsearch hosts
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_HOSTS="localhost1:9201,localhost2:9202"
  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost1:9201,localhost2:9202" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_ELASTICSEARCH_HOSTS
}

function test_shopware_args_opensearch() {
  # Default
  local ARGS=("")
  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"

  # Custom values - opensearch disabled
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="false"
  shopware_args_opensearch
  assert_array_not_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_not_contains "--es-hosts=opensearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_OPENSEARCH_ENABLED

  # Custom values - opensearch indexing disabled
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_INDEXING_ENABLED="false"
  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=opensearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_OPENSEARCH_INDEXING_ENABLED
  unset SHOPWARE_OPENSEARCH_ENABLED

  # Custom values - custom opensearch host
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_HOST="localhost"
  local SHOPWARE_OPENSEARCH_PORT="9201"
  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost:9201" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_OPENSEARCH_HOST
  unset SHOPWARE_OPENSEARCH_PORT
  unset SHOPWARE_OPENSEARCH_ENABLED

  # Custom values - custom opensearch hosts
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_HOSTS="localhost1:9201,localhost2:9202"
  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost1:9201,localhost2:9202" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
  unset SHOPWARE_OPENSEARCH_HOSTS
}

function test_shopware_args_extra() {
  # Default
  local ARGS=("")
  shopware_args_extra
  assert_equals "" "${ARGS[@]}"

  # Custom values
  local ARGS=("")
  local SHOPWARE_EXTRA_INSTALL_ARGS="--extra-arg=1"
  shopware_args_extra
  assert_array_contains "--extra-arg=1" "${ARGS[@]}"
}

function test_shopware_env_puppeteer_skip_chromium_download() {
  local PUPPETEER_SKIP_CHROMIUM_DOWNLOAD="0"
  shopware_env_puppeteer_skip_chromium_download
  assert_equals "1" "${PUPPETEER_SKIP_CHROMIUM_DOWNLOAD}"

  unset PUPPETEER_SKIP_CHROMIUM_DOWNLOAD
  local SHOPWARE_PUPPETEER_SKIP_CHROMIUM_DOWNLOAD="false"
  shopware_env_puppeteer_skip_chromium_download
  assert_empty "${PUPPETEER_SKIP_CHROMIUM_DOWNLOAD:-}"
}

function test_shopware_env_ci() {
  # Default
  shopware_env_ci
  assert_equals "1" "${CI}"

  # SHOPWARE_CI=true
  unset CI
  local SHOPWARE_CI="true"
  shopware_env_ci
  assert_equals "1" "${CI}"

  # SHOPWARE_CI=false
  unset CI
  local SHOPWARE_CI="false"
  shopware_env_ci
  assert_equals "0" "${CI}"
}

function test_shopware_env_skip_bundle_dump() {
  # Default
  shopware_env_skip_bundle_dump
  assert_equals "0" "${SHOPWARE_SKIP_BUNDLE_DUMP}"

  # SHOPWARE_CI=true
  unset SHOPWARE_SKIP_BUNDLE_DUMP
  local SHOPWARE_SKIP_BUNDLE_DUMP="true"
  shopware_env_skip_bundle_dump
  assert_equals "1" "${SHOPWARE_SKIP_BUNDLE_DUMP}"

  # SHOPWARE_CI=false
  unset SHOPWARE_SKIP_BUNDLE_DUMP
  local SHOPWARE_SKIP_BUNDLE_DUMP="false"
  shopware_env_skip_bundle_dump
  assert_equals "0" "${SHOPWARE_SKIP_BUNDLE_DUMP}"
}

function test_shopware_env_disable_admin_compilation_typecheck() {
  # Default
  shopware_env_disable_admin_compilation_typecheck
  assert_equals "1" "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"

  # SHOPWARE_CI=true
  unset DISABLE_ADMIN_COMPILATION_TYPECHECK
  local SHOPWARE_DISABLE_ADMIN_COMPILATION_TYPECHECK="true"
  shopware_env_disable_admin_compilation_typecheck
  assert_equals "1" "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"

  # SHOPWARE_CI=false
  unset DISABLE_ADMIN_COMPILATION_TYPECHECK
  local SHOPWARE_DISABLE_ADMIN_COMPILATION_TYPECHECK="false"
  shopware_env_disable_admin_compilation_typecheck
  assert_empty "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"
}

function test_shopware_configure_lock_dsn() {
  # File exists
  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=flock://var/lock"
  rm -fr './test-data'

  # File exists and contains a LOCK_DSN
  mkdir -p "${APP_PATH}"
  echo "LOCK_DSN=redis://redis:6379" >"${APP_PATH}/.env"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=redis://redis:6379"
  rm -fr './test-data'

  # File doesn't exist
  mkdir -p "${APP_PATH}"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=flock://var/lock"
  rm -fr './test-data'
}

function test_shopware_maintenance_enable() {
  spy console
  shopware_maintenance_enable
  assert_have_been_called_with "sales-channel:maintenance:enable --all" console
}

function test_shopware_maintenance_disable() {
  spy console
  shopware_maintenance_disable
  assert_have_been_called_with "sales-channel:maintenance:disable --all" console
}

function test_shopware_bundle_dump() {
  spy console
  shopware_bundle_dump
  assert_have_been_called_times 2 console
}

function test_shopware_skip_asset_build_flag() {
  # Default
  assert_empty "$(shopware_skip_asset_build_flag)"

  # SHOPWARE_SKIP_ASSET_COPY=true
  local SHOPWARE_SKIP_ASSET_COPY="true"
  assert_equals "--skip-asset-build" "$(shopware_skip_asset_build_flag)"
}

function test_shopware_install_all_plugins() {
  mock shopware_list_plugins_not_installed echo 'test1 test2'
  spy console
  shopware_install_all_plugins
  assert_have_been_called_times 2 console
}

function test_shopware_update_all_plugins() {
  mock shopware_list_plugins_not_installed echo 'test1 test2'
  spy console
  shopware_install_all_plugins
  assert_have_been_called_times 2 console

  # Shopware version is 6.6+
  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_update_all_plugins
  assert_have_been_called_with "plugin:update:all" console
  assert_have_been_called_times 1 console
}

function test_shopware_configure() {
  spy shopware_args_defaults
  spy shopware_args_elasticsearch
  spy shopware_args_opensearch
  spy shopware_args_extra
  spy shopware_env_puppeteer_skip_chromium_download
  spy shopware_env_ci
  spy shopware_env_disable_admin_compilation_typecheck
  spy shopware_env_skip_bundle_dump
  spy console
  spy shopware_configure_lock_dsn

  shopware_configure

  assert_have_been_called shopware_args_defaults
  assert_have_been_called shopware_args_elasticsearch
  assert_have_been_called shopware_args_opensearch
  assert_have_been_called shopware_args_extra
  assert_have_been_called shopware_env_puppeteer_skip_chromium_download
  assert_have_been_called shopware_env_ci
  assert_have_been_called shopware_env_disable_admin_compilation_typecheck
  assert_have_been_called shopware_env_skip_bundle_dump
  assert_have_been_called_with "system:setup --force" console
  assert_have_been_called shopware_configure_lock_dsn

  # Test if SHOPWARE_SKIP_INSTALL=true
  local SHOPWARE_SKIP_INSTALL="true"
  spy shopware_args_defaults
  spy shopware_args_elasticsearch
  spy shopware_args_opensearch
  spy shopware_args_extra
  spy shopware_env_puppeteer_skip_chromium_download
  spy shopware_env_ci
  spy shopware_env_disable_admin_compilation_typecheck
  spy shopware_env_skip_bundle_dump
  spy console
  spy shopware_configure_lock_dsn
  shopware_configure
  assert_have_been_called_times 0 shopware_args_defaults
  assert_have_been_called_times 0 shopware_args_elasticsearch
  assert_have_been_called_times 0 shopware_args_opensearch
  assert_have_been_called_times 0 shopware_args_extra
  assert_have_been_called_times 0 shopware_env_puppeteer_skip_chromium_download
  assert_have_been_called_times 0 shopware_env_ci
  assert_have_been_called_times 0 shopware_env_disable_admin_compilation_typecheck
  assert_have_been_called_times 0 shopware_env_skip_bundle_dump
  assert_have_been_called_times 0 console
  assert_have_been_called_times 0 shopware_configure_lock_dsn
}

function test_shopware_install() {
  spy console
  shopware_install
  assert_have_been_called_with "system:install --force --create-database --basic-setup --shop-locale=en-GB --shop-currency=EUR" console

  # Test with custom values
  local SHOPWARE_LOCALE="en-US"
  local SHOPWARE_CURRENCY="USD"
  spy console
  shopware_install
  assert_have_been_called_with "system:install --force --create-database --basic-setup --shop-locale=en-US --shop-currency=USD" console
}

function test_shopware_theme_change() {
  spy console
  shopware_theme_change
  assert_have_been_called_with "theme:change --all Storefront" console
}

function test_shopware_system_update_finish() {
  spy console
  shopware_system_update_finish
  assert_have_been_called_with "system:update:finish" console

  # with enabled SHOPWARE_SKIP_ASSET_COPY
  local SHOPWARE_SKIP_ASSET_COPY="true"
  spy console
  shopware_system_update_finish
  assert_have_been_called_with "system:update:finish --skip-asset-build" console
}

function test_shopware_plugin_refresh() {
  mock shopware_version echo "v6.5.0.0"
  spy console
  shopware_plugin_refresh
  assert_have_been_called_with "plugin:refresh" console

  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_plugin_refresh
  assert_have_been_called_times 0 console
}

function test_shopware_scheduled_task_register() {
  spy console
  shopware_scheduled_task_register
  assert_have_been_called_with "scheduled-task:register" console
}

function test_shopware_theme_refresh() {
  spy console
  shopware_theme_refresh
  assert_have_been_called_with "theme:refresh" console
}

function test_shopware_system_config_set() {
  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_system_config_set
  assert_have_been_called_with "system:config:set core.frw.completedAt 2019-10-07T10:46:23+00:00" console

  mock shopware_version echo "v6.5.0.0"
  spy console
  shopware_system_config_set
  assert_have_been_called_times 2 console
}

function test_shopware_admin_user_exists() {
  mock shopware_version echo "v6.5.0.0"
  assert_exit_code 0 "$(shopware_admin_user_exists)"

  mock shopware_version echo "v6.6.0.0"
  mock console <<EOF
[
{}
]
EOF
  assert_exit_code 1 "$(shopware_admin_user_exists)"

  mock console <<EOF
[
{\"username\":\"admin\"}
]
EOF
  assert_exit_code 0 "$(shopware_admin_user_exists)"

  mock console <<EOF
[
{\"username\":\"adm\"}
]
EOF
  assert_exit_code 1 "$(shopware_admin_user_exists)"

  mock console false
  assert_exit_code 0 "$(shopware_admin_user_exists)"
}

function test_shopware_admin_user() {
  mock shopware_admin_user_exists false
  spy console
  shopware_admin_user
  assert_have_been_called_with "user:create admin --admin --firstName=admin --lastName=admin --email=admin@example.com --password=ASDqwe123" console

  # Test with custom values
  local SHOPWARE_USERNAME="johndoe"
  local SHOPWARE_FIRST_NAME="John"
  local SHOPWARE_LAST_NAME="Doe"
  local SHOPWARE_EMAIL="johndoe@example.com"
  local SHOPWARE_PASSWORD="johndoepw"
  spy console
  shopware_admin_user
  assert_have_been_called_with "user:create johndoe --admin --firstName=John --lastName=Doe --email=johndoe@example.com --password=johndoepw" console
  unset SHOPWARE_USERNAME
  unset SHOPWARE_FIRSTNAME
  unset SHOPWARE_LASTNAME
  unset SHOPWARE_EMAIL
  unset SHOPWARE_PASSWORD

  # Below Shopware 6.6 it should not fail even if the console command exits with a non-zero exit code
  mock shopware_admin_user_exists true
  mock shopware_version echo "v6.5.0.0"
  mock console false
  spy true
  shopware_admin_user
  assert_exit_code 0 "$(shopware_admin_user)"
  assert_have_been_called true

  # Above Shopware 6.6 it should just change the password
  mock shopware_admin_user_exists true
  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_admin_user
  assert_have_been_called_with "user:change-password admin --password=ASDqwe123" console
  assert_exit_code 0 "$(shopware_admin_user)"

  # Custom username and password
  local SHOPWARE_USERNAME="johndoe"
  local SHOPWARE_PASSWORD="johndoepw"
  spy console
  shopware_admin_user
  assert_have_been_called_with "user:change-password johndoe --password=johndoepw" console
}

function test_shopware_disable_deploy_sample_data() {
  local SHOPWARE_DEPLOY_SAMPLE_DATA="true"
  shopware_disable_deploy_sample_data
  assert_equals "false" "${SHOPWARE_DEPLOY_SAMPLE_DATA}"
}

function test_shopware_deploy_sample_data() {
  spy console
  shopware_deploy_sample_data
  assert_have_been_called_times 0 console

  local APP_PATH="./test-data/app"
  local SHOPWARE_DEPLOY_SAMPLE_DATA="true"
  spy console
  shopware_deploy_sample_data
  assert_have_been_called_times 5 console

  rm -fr './test-data'
}

function test_shopware_cache_clear() {
  spy console
  shopware_cache_clear
  assert_have_been_called_with "cache:clear --no-warmup" console
}

function test_shopware_cache_warmup() {
  spy console
  shopware_cache_warmup
  assert_have_been_called_times 2 console
}

function test_shopware_publish_config() {
  # Test with a valid SHARED_CONFIG_PATH
  local SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_publish_config
  assert_file_exists "test-data/config/.env"
  rm -fr "./test-data"
  unset SHARED_CONFIG_PATH

  # Test if SHARED_CONFIG_PATH is not writable (/config by default)
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_publish_config
  assert_file_exists "/tmp/.env"
  rm -fr "/tmp/.env"
  rm -fr "./test-data"
}

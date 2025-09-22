#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/install.sh"
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
  assert_have_been_called_with eval "echo 'test'"
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
  assert_have_been_called_with eval "echo 'test'"
}

function test_bootstrap_check_default() {
  setup

  # If both are false it should not call the command_after_install
  assert_empty "$(bootstrap_check)"
}

function test_bootstrap_check_skip_bootstrap_but_command_after_install_set() {
  # If MAGENTO_SKIP_BOOTSTRAP is true, it should just run the COMMAND_AFTER_INSTALL and exit
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local SHOPWARE_SKIP_BOOTSTRAP="true"
  setup

  assert_contains "test-123" "$(bootstrap_check)"

  # If SKIP_BOOTSTRAP is true, it should run the COMMAND_AFTER_INSTALL and exit
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
}

function test_bootstrap_check_both_enabled() {
  # If both are true it should run the COMMAND_AFTER_INSTALL
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local SHOPWARE_SKIP_BOOTSTRAP="true"
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

function test_shopware_version() {
  setup
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
  setup
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

function test_shopware_is_installed_shopware_65() {
  setup

  mock shopware_version echo "6.6.0.0"
  mock console true
  assert_exit_code 0 "$(shopware_is_installed)"

  mock console false
  assert_exit_code 1 "$(shopware_is_installed)"
}

function test_shopware_is_installed_shopware_66() {
  setup

  mock shopware_version echo "6.6.0.0"
  mock console true
  assert_exit_code 0 "$(shopware_is_installed)"

  mock console false
  assert_exit_code 1 "$(shopware_is_installed)"
}

function test_shopware_args_defaults() {
  # Default
  local ARGS=("")
  setup

  shopware_args_defaults
  assert_array_contains "--app-env=prod" "${ARGS[@]}"
  assert_array_contains "--app-url=https://shopware.test" "${ARGS[@]}"
  assert_array_contains "--database-url=mysql://app:app@db:3306/shopware" "${ARGS[@]}"
  assert_array_contains "--cdn-strategy=physical_filename" "${ARGS[@]}"
  assert_array_contains "--mailer-url=native://default" "${ARGS[@]}"
}

function test_shopware_args_defaults_custom() {
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
  setup

  shopware_args_defaults
  assert_array_contains "--app-env=dev" "${ARGS[@]}"
  assert_array_contains "--app-url=http://shopware.test" "${ARGS[@]}"
  assert_array_contains "--database-url=mysql://root:root@localhost:3306/shopware2" "${ARGS[@]}"
  assert_array_contains "--cdn-strategy=symlink" "${ARGS[@]}"
  assert_array_contains "--mailer-url=smtp://localhost" "${ARGS[@]}"
}

function test_shopware_args_defaults_custom_host() {
  # Custom values for SHOPWARE_HOST and SHOPWARE_SCHEME
  local ARGS=("")
  local SHOPWARE_HOST="localhost"
  local SHOPWARE_SCHEME="http"
  setup

  shopware_args_defaults
  assert_array_contains "--app-url=http://localhost" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch_default() {
  # Default
  local ARGS=("")
  setup

  shopware_args_elasticsearch
  assert_array_not_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_not_contains "--es-hosts" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"

  # Custom values - custom elasticsearch hosts
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_HOSTS="localhost1:9201,localhost2:9202"
  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost1:9201,localhost2:9202" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch_enabled() {
  # Custom values - elasticsearch enabled
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  setup

  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=elasticsearch:9200" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch_enabled_indexing_disabled() {
  # Custom values - elasticsearch indexing disabled
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_INDEXING_ENABLED="false"
  setup

  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=elasticsearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch_default_custom_host() {
  # Custom values - custom elasticsearch host
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_HOST="localhost"
  local SHOPWARE_ELASTICSEARCH_PORT="9201"
  setup

  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost:9201" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_elasticsearch_default_custom_hosts() {
  # Custom values - custom elasticsearch hosts
  local ARGS=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_ELASTICSEARCH_HOSTS="localhost1:9201,localhost2:9202"
  setup

  shopware_args_elasticsearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost1:9201,localhost2:9202" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_opensearch_default() {
  # Default
  local ARGS=("")
  setup

  shopware_args_opensearch
  assert_array_not_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_not_contains "--es-hosts=opensearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_opensearch_enabled() {
  # Custom values - opensearch enabled
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  setup

  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_opensearch_enabled_indexing_disabled() {
  # Custom values - opensearch indexing disabled
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_INDEXING_ENABLED="false"
  setup

  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=opensearch:9200" "${ARGS[@]}"
  assert_array_not_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_opensearch_custom_host() {
  # Custom values - custom opensearch host
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_HOST="localhost"
  local SHOPWARE_OPENSEARCH_PORT="9201"
  setup

  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost:9201" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_opensearch_custom_hosts() {
  # Custom values - custom opensearch hosts
  local ARGS=("")
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_HOSTS="localhost1:9201,localhost2:9202"
  setup

  shopware_args_opensearch
  assert_array_contains "--es-enabled=1" "${ARGS[@]}"
  assert_array_contains "--es-hosts=localhost1:9201,localhost2:9202" "${ARGS[@]}"
  assert_array_contains "--es-indexing-enabled=1" "${ARGS[@]}"
}

function test_shopware_args_extra_default() {
  # Default
  local ARGS=("")
  setup

  shopware_args_extra
  assert_equals "" "${ARGS[@]}"
}

function test_shopware_args_extra_custom() {
  # Custom values
  local ARGS=("")
  local SHOPWARE_EXTRA_INSTALL_ARGS="--extra-arg=1"
  setup

  shopware_args_extra
  assert_array_contains "--extra-arg=1" "${ARGS[@]}"
}

function test_shopware_env_puppeteer_skip_chromium_download_env() {
  local PUPPETEER_SKIP_CHROMIUM_DOWNLOAD="0"
  setup

  shopware_env_puppeteer_skip_chromium_download
  assert_equals "1" "${PUPPETEER_SKIP_CHROMIUM_DOWNLOAD}"
}

function test_shopware_env_puppeteer_skip_chromium_download_disabled() {
  local SHOPWARE_PUPPETEER_SKIP_CHROMIUM_DOWNLOAD="false"
  setup

  shopware_env_puppeteer_skip_chromium_download
  assert_empty "${PUPPETEER_SKIP_CHROMIUM_DOWNLOAD:-}"
}

function test_shopware_env_ci_default() {
  # Default
  setup

  shopware_env_ci
  assert_equals "1" "${CI}"
}
function test_shopware_env_ci_true() {
  # SHOPWARE_CI=true
  local SHOPWARE_CI="true"
  setup

  shopware_env_ci
  assert_equals "1" "${CI}"
}

function test_shopware_env_ci_false() {
  # SHOPWARE_CI=false
  local SHOPWARE_CI="false"
  setup

  shopware_env_ci
  assert_equals "0" "${CI}"
}

function test_shopware_env_skip_bundle_dump() {
  # Default
  setup

  shopware_env_skip_bundle_dump
  assert_equals "0" "${SHOPWARE_SKIP_BUNDLE_DUMP}"
}

function test_shopware_env_skip_bundle_dump_ci_true() {
  # SHOPWARE_CI=true
  local SHOPWARE_SKIP_BUNDLE_DUMP="true"
  setup

  shopware_env_skip_bundle_dump
  assert_equals "1" "${SHOPWARE_SKIP_BUNDLE_DUMP}"
}

function test_shopware_env_skip_bundle_dump_ci_false() {
  # SHOPWARE_CI=false
  local SHOPWARE_SKIP_BUNDLE_DUMP="false"
  setup

  shopware_env_skip_bundle_dump
  assert_equals "0" "${SHOPWARE_SKIP_BUNDLE_DUMP}"
}

function test_shopware_env_disable_admin_compilation_typecheck_default() {
  # Default
  setup

  shopware_env_disable_admin_compilation_typecheck
  assert_equals "1" "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"
}

function test_shopware_env_disable_admin_compilation_typecheck_true() {
  # SHOPWARE_CI=true
  unset DISABLE_ADMIN_COMPILATION_TYPECHECK
  local SHOPWARE_DISABLE_ADMIN_COMPILATION_TYPECHECK="true"
  setup

  shopware_env_disable_admin_compilation_typecheck
  assert_equals "1" "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"
}

function test_shopware_env_disable_admin_compilation_typecheck_false() {
  # SHOPWARE_CI=false
  local SHOPWARE_DISABLE_ADMIN_COMPILATION_TYPECHECK="false"
  setup

  shopware_env_disable_admin_compilation_typecheck
  assert_empty "${DISABLE_ADMIN_COMPILATION_TYPECHECK:-}"
}

function test_shopware_configure_lock_dsn_exists_not_configured() {
  # File exists
  setup

  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=flock://var/lock"
  rm -fr './test-data'
}

function test_shopware_configure_lock_dsn_configured() {
  # File exists and contains a LOCK_DSN
  setup

  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  echo "LOCK_DSN=redis://redis:6379" >"${APP_PATH}/.env"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=redis://redis:6379"
  rm -fr './test-data'
}

function test_shopware_configure_lock_dsn_not_exists() {
  # File doesn't exist
  setup

  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}"
  shopware_configure_lock_dsn
  assert_file_contains "${APP_PATH}/.env" "LOCK_DSN=flock://var/lock"
  rm -fr './test-data'
}

function test_shopware_maintenance_enable() {
  setup

  spy console
  shopware_maintenance_enable
  assert_have_been_called_with console "sales-channel:maintenance:enable --all"
}

function test_shopware_maintenance_disable() {
  setup

  spy console
  shopware_maintenance_disable
  assert_have_been_called_with console "sales-channel:maintenance:disable --all"
}

function test_shopware_skip_asset_build_flag_default() {
  # Default
  setup
  assert_empty "$(shopware_skip_asset_build_flag)"
}

function test_shopware_skip_asset_build_flag_skip() {
  # SHOPWARE_SKIP_ASSET_COPY=true
  local SHOPWARE_SKIP_ASSET_COPY="true"
  setup

  assert_equals "--skip-asset-build" "$(shopware_skip_asset_build_flag)"
}

function test_shopware_install_all_plugins() {
  setup

  mock shopware_list_plugins_not_installed echo 'test1 test2'
  spy console
  shopware_install_all_plugins
  assert_have_been_called_times 2 console
}

function test_shopware_update_all_plugins_default() {
  setup

  mock shopware_list_plugins_not_installed echo 'test1 test2'
  spy console
  shopware_install_all_plugins
  assert_have_been_called_times 2 console
}

function test_shopware_update_all_plugins_shopware_66() {
  setup

  # Shopware version is 6.6+
  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_update_all_plugins
  assert_have_been_called_with console "plugin:update:all"
  assert_have_been_called_times 1 console
}

function test_shopware_configure_default() {
  setup

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
  assert_have_been_called_with console "system:setup --force"
  assert_have_been_called shopware_configure_lock_dsn
}

function test_shopware_configure_skip_install() {
  # Test if SHOPWARE_SKIP_INSTALL=true
  local SHOPWARE_SKIP_INSTALL="true"
  setup

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

function test_shopware_install_default() {
  setup

  local APP_PATH="./test-data/app"

  spy console
  shopware_install
  assert_have_been_called_with console "system:install --force --create-database --basic-setup --shop-locale=en-GB --shop-currency=EUR"

  rm -fr './test-data'
}

function test_shopware_install_custom() {
  # Test with custom values
  local APP_PATH="./test-data/app"
  local SHOPWARE_LOCALE="en-US"
  local SHOPWARE_CURRENCY="USD"
  setup

  spy console
  shopware_install
  assert_have_been_called_with console "system:install --force --create-database --basic-setup --shop-locale=en-US --shop-currency=USD"

  rm -fr './test-data'
}

function test_shopware_theme_change() {
  setup

  spy console
  shopware_theme_change
  assert_have_been_called_with console "theme:change --all Storefront"
}

function test_shopware_system_update_finish_default() {
  setup

  spy console
  shopware_system_update_finish
  assert_have_been_called_with console "system:update:finish"
}

function test_shopware_system_update_finish_skip() {
  # with enabled SHOPWARE_SKIP_ASSET_COPY
  local SHOPWARE_SKIP_ASSET_COPY="true"
  setup

  spy console
  shopware_system_update_finish
  assert_have_been_called_with console "system:update:finish --skip-asset-build"
}

function test_shopware_plugin_refresh_shopware_65() {
  setup
  mock shopware_version echo "v6.5.0.0"
  spy console
  shopware_plugin_refresh
  assert_have_been_called_with console "plugin:refresh"
}

function test_shopware_plugin_refresh_shopware_66() {
  setup

  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_plugin_refresh
  assert_have_been_called_times 0 console
}

function test_shopware_scheduled_task_register() {
  setup

  spy console
  shopware_scheduled_task_register
  assert_have_been_called_with console "scheduled-task:register"
}

function test_shopware_theme_refresh() {
  setup

  spy console
  shopware_theme_refresh
  assert_have_been_called_with console "theme:refresh"
}

function test_shopware_system_config_set_shopware_66() {
  setup

  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_system_config_set
  assert_have_been_called_with console "system:config:set core.frw.completedAt 2019-10-07T10:46:23+00:00"
}

function test_shopware_system_config_set_shopware_65() {
  setup

  mock shopware_version echo "v6.5.0.0"
  spy console
  shopware_system_config_set
  assert_have_been_called_times 2 console
}

function test_shopware_admin_user_exists_65_ignored() {
  setup

  mock shopware_version echo "v6.5.0.0"
  assert_exit_code 0 "$(shopware_admin_user_exists)"
}

function test_shopware_admin_user_exists_66_not_exist() {
  setup

  mock shopware_version echo "v6.6.0.0"
  mock console <<EOF
[
{}
]
EOF
  assert_exit_code 1 "$(shopware_admin_user_exists)"
}

function test_shopware_admin_user_exists_66_exist() {
  setup

  mock shopware_version echo "v6.6.0.0"
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

function test_shopware_admin_user_exists_66_not_exist_similar_name() {
  setup

  mock shopware_version echo "v6.6.0.0"
  mock console <<EOF
[
{\"username\":\"adm\"}
]
EOF
  assert_exit_code 1 "$(shopware_admin_user_exists)"
}

function test_shopware_admin_user_exists_false() {
  setup

  mock console false
  assert_exit_code 0 "$(shopware_admin_user_exists)"
}

function test_shopware_admin_user_default() {
  setup

  mock shopware_admin_user_exists false
  spy console
  shopware_admin_user
  assert_have_been_called_with console "user:create admin --admin --firstName=admin --lastName=admin --email=admin@example.com --password=ASDqwe123"
}

function test_shopware_admin_user_custom() {
  # Test with custom values
  local SHOPWARE_USERNAME="johndoe"
  local SHOPWARE_FIRST_NAME="John"
  local SHOPWARE_LAST_NAME="Doe"
  local SHOPWARE_EMAIL="johndoe@example.com"
  local SHOPWARE_PASSWORD="johndoepw"
  setup

  mock shopware_admin_user_exists false
  spy console
  shopware_admin_user
  assert_have_been_called_with console "user:create johndoe --admin --firstName=John --lastName=Doe --email=johndoe@example.com --password=johndoepw"
}

function test_shopware_admin_user_exists_65() {
  # Below Shopware 6.6 it should not fail even if the console command exits with a non-zero exit code
  setup

  mock shopware_admin_user_exists true
  mock shopware_version echo "v6.5.0.0"
  mock console false
  spy true
  shopware_admin_user
  assert_exit_code 0 "$(shopware_admin_user)"
  assert_have_been_called true
}

function test_shopware_admin_user_exists_66() {
  # Above Shopware 6.6 it should just change the password
  setup

  mock shopware_admin_user_exists true
  mock shopware_version echo "v6.6.0.0"
  spy console
  shopware_admin_user
  assert_have_been_called_with console "user:change-password admin --password=ASDqwe123"
  assert_exit_code 0 "$(shopware_admin_user)"
}

function test_shopware_admin_user_custom_username_password() {
  # Custom username and password
  local SHOPWARE_USERNAME="johndoe"
  local SHOPWARE_PASSWORD="johndoepw"
  setup

  mock shopware_admin_user_exists true
  spy console
  shopware_admin_user
  assert_have_been_called_with console "user:change-password johndoe --password=johndoepw"
}

function test_shopware_disable_deploy_sample_data() {
  local SHOPWARE_DEPLOY_SAMPLE_DATA="true"
  setup

  shopware_disable_deploy_sample_data
  assert_equals "false" "${SHOPWARE_DEPLOY_SAMPLE_DATA}"
}

function test_shopware_deploy_sample_data_default() {
  setup

  spy console
  shopware_deploy_sample_data
  assert_have_been_called_times 0 console
}

function test_shopware_deploy_sample_data_enabled() {
  local APP_PATH="./test-data/app"
  local SHOPWARE_DEPLOY_SAMPLE_DATA="true"
  setup

  spy console
  shopware_deploy_sample_data
  assert_have_been_called_times 4 console
  rm -fr './test-data'
}

function test_shopware_deploy_sample_data_force() {
  local APP_PATH="./test-data/app"
  local SHOPWARE_DEPLOY_SAMPLE_DATA="false"
  local SHOPWARE_FORCE_DEPLOY_SAMPLE_DATA="true"
  setup

  spy console
  shopware_deploy_sample_data
  assert_have_been_called_times 4 console

  rm -fr './test-data'
}

function test_shopware_cache_clear() {
  setup

  spy console
  shopware_cache_clear
  assert_have_been_called_with console "cache:clear --no-warmup"
}

function test_shopware_cache_warmup() {
  setup

  spy console
  shopware_cache_warmup
  assert_have_been_called_with console "cache:warmup"
}

function test_shopware_reindex_default() {
  setup

  spy console
  shopware_reindex
  assert_have_been_called_times 0 console
}

function test_shopware_reindex_opensearch() {
  # Opensearch enabled but by default it doesn't run reindex (if shopware_dont_skip_reindex is not called)
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  setup

  spy console
  shopware_reindex
  assert_have_been_called_times 0 console

  spy console
  shopware_dont_skip_reindex
  shopware_reindex
  assert_have_been_called_times 3 console
}

function test_shopware_reindex_elasticsearch() {
  # Elasticsearch enabled
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  setup

  spy console
  shopware_dont_skip_reindex
  shopware_reindex
  assert_have_been_called_times 3 console
}

function test_shopware_configure_redis() {
  local APP_PATH="./test-data/app"
  setup

  # Default
  shopware_configure_redis
  assert_is_file_empty "${APP_PATH}/config/packages/zz-redis.yml"
  rm -fr './test-data'
}

function test_shopware_configure_redis_custom() {
  # Custom values
  local APP_PATH="./test-data/app"
  local SHOPWARE_REDIS_ENABLED=true
  setup

  shopware_configure_redis
  assert_file_contains "${APP_PATH}/config/packages/zz-redis.yml" "app: cache.adapter.redis"
  rm -fr './test-data'

  mock shopware_version echo "v6.4.0.0"
  shopware_configure_redis
  assert_file_not_contains "${APP_PATH}/config/packages/zz-redis.yml" "connection: \"redis_cart\""
  assert_file_contains "${APP_PATH}/config/packages/zz-redis.yml" "redis_url: \"%env(string:REDIS_URL)%/4?persistent=1\""
  rm -fr './test-data'

  mock shopware_version echo "v6.6.7.0"
  shopware_configure_redis
  assert_file_not_contains "${APP_PATH}/config/packages/zz-redis.yml" "connection: \"redis_cart\""
  assert_file_contains "${APP_PATH}/config/packages/zz-redis.yml" "redis_url: \"%env(string:REDIS_URL)%/4?persistent=1\""
  rm -fr './test-data'

  mock shopware_version echo "v6.6.8.0"
  shopware_configure_redis
  assert_file_not_contains "${APP_PATH}/config/packages/zz-redis.yml" "redis_url: \"%env(string:REDIS_URL)%/4?persistent=1\""
  assert_file_contains "${APP_PATH}/config/packages/zz-redis.yml" "connection: \"redis_cart\""
  rm -fr './test-data'

  mock shopware_version echo "v6.7.0.0"
  shopware_configure_redis
  assert_file_not_contains "${APP_PATH}/config/packages/zz-redis.yml" "redis_url: \"%env(string:REDIS_URL)%/4?persistent=1\""
  assert_file_contains "${APP_PATH}/config/packages/zz-redis.yml" "connection: \"redis_cart\""
  rm -fr './test-data'
}

function test_shopware_publish_shared_files_default() {
  setup

  # Test with a valid SHARED_CONFIG_PATH
  local SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_publish_shared_files
  assert_file_exists "test-data/config/.env"
  rm -fr "./test-data"
}

function test_shopware_publish_shared_files_not_writable() {
  setup

  # Test if SHARED_CONFIG_PATH is not writable (/config by default)
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/.env"
  shopware_publish_shared_files
  assert_file_exists "/tmp/.env"
  rm -fr "/tmp/.env"
  rm -fr "./test-data"
}

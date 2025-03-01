#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/build.sh"
}

function test_command_before_build_default() {
  # Default
  setup

  assert_exit_code 0 "$(command_before_build)"
}

function test_command_before_build_custom() {
  # Custom command
  local COMMAND_BEFORE_BUILD="echo 'test'"
  setup

  spy eval
  command_before_build
  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_build_default() {
  # Default
  setup

  assert_exit_code 0 "$(command_before_build)"
}

function test_command_after_build_custom() {
  # Custom command
  local COMMAND_AFTER_BUILD="echo 'test'"
  setup

  spy eval
  command_after_build
  assert_have_been_called_with "echo 'test'" eval
}

function test_n_install_default() {
  setup

  spy n
  n_install
  assert_have_been_called_times 0 n
}

function test_n_install_custom() {
  local NODE_VERSION="16"
  setup

  spy n
  n_install
  assert_have_been_called_with "install 16" n
}

function test_composer_self_update_default() {
  setup

  spy composer
  composer_self_update
  assert_have_been_called_times 0 composer
}

function test_composer_self_update_1() {
  local COMPOSER_VERSION="1"
  setup

  spy composer
  composer_self_update
  assert_have_been_called_with "self-update --1" composer
}

function test_composer_self_update_2() {
  local COMPOSER_VERSION="2"
  setup

  spy composer
  composer_self_update
  assert_have_been_called_with "self-update --2" composer
}

function test_composer_self_update_major_minor_version() {
  local COMPOSER_VERSION="2.2"
  setup

  spy composer
  composer_self_update
  assert_have_been_called_with "self-update --2.2" composer
}

function test_composer_self_update_semantic_version() {
  local COMPOSER_VERSION="2.5.12"
  setup

  spy composer
  composer_self_update
  assert_have_been_called_with "self-update 2.5.12" composer
}

function test_composer_configure_default() {
  setup
  # Default
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
  local GITHUB_USER="user"
  local GITHUB_TOKEN="token"
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

function test_composer_install() {
  setup

  spy composer
  composer_install
  assert_have_been_called_times 0 composer

  touch composer.json
  spy composer
  composer_install
  assert_have_been_called_with "install --no-progress" composer
  rm -f composer.json
}

function test_composer_clear_cache() {
  setup

  spy composer

  composer_clear_cache
  assert_have_been_called_with "clear-cache" composer
}

function test_composer_dump_autoload() {
  setup

  spy composer

  composer_dump_autoload
  assert_have_been_called_with "dump-autoload --optimize" composer
}

function test_magento_setup_di_compile_default() {
  # By default it should run
  setup

  mock magento echo
  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 1 magento
}

function test_magento_setup_di_compile_disabled() {
  # If MAGENTO_DI_COMPILE is false, it should not run
  local MAGENTO_DI_COMPILE="false"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_di_compile_on_demand_enabled() {
  # if magento_di_compile_on_demand is true, it should not run
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_di_compile_enabled_on_demand_enabled() {
  # If both is true, it should not run
  local MAGENTO_DI_COMPILE="true"
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_di_compile_disabled_on_demand_disabled() {
  # If both is false, it should not run
  local MAGENTO_DI_COMPILE="false"
  local MAGENTO_DI_COMPILE_ON_DEMAND="false"
  setup

  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy_default() {
  setup

  mock nproc echo 4
  spy magento

  # by default it should not run
  magento_setup_static_content_deploy
  assert_have_been_called_times 1 magento
}

function test_magento_setup_static_content_deploy_enabled() {
  # if MAGENTO_SKIP_STATIC_CONTENT_DEPLOY is true, it should run
  local MAGENTO_SKIP_STATIC_CONTENT_DEPLOY="true"
  setup

  spy magento
  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy_scd_on_demand() {
  # If MAGENTO_SCD_ON_DEMAND is true, it should run
  local MAGENTO_SCD_ON_DEMAND="true"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy_scd_enabled_on_demand_enabled() {
  # If both is true, it should not run
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_SCD_ON_DEMAND="true"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy_args_magento_24_plus() {
  # SCD args is -fv if Magento version is 2.4+
  local MAGENTO_VERSION="2.4"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv" magento
}

function test_magento_setup_static_content_deploy_args_magento_24_minus() {
  # SCD args is -v if Magento version is 2.3
  local MAGENTO_VERSION="2.3.3"
  setup

  mock nproc echo 4
  spy magento
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -v" magento
}

function test_magento_setup_static_content_deploy_args_magento_languages() {
  # MAGENTO_LANGUAGES appended to the command
  local MAGENTO_LANGUAGES="en_US de_DE"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv en_US de_DE" magento
}

function test_magento_setup_static_content_deploy_args_magento_themes() {
  # MAGENTO_THEMES appended to the command
  local MAGENTO_THEMES="Magento/blank Magento/luma"
  setup

  mock nproc echo 4
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv --theme=Magento/blank --theme=Magento/luma" magento
}

function test_magento_create_pub_static_dir() {
  setup

  local APP_PATH="./test-data/app"

  magento_create_pub_static_dir
  assert_directory_exists "${APP_PATH}/pub/static"

  rm -fr "./test-data"
}

function test_dump_build_version() {
  setup

  local APP_PATH="./test-data/app"

  dump_build_version
  assert_file_contains "${APP_PATH}/pub/version.php" "php-version: "
  assert_file_contains "${APP_PATH}/pub/version.php" "build-date: "

  rm -fr "./test-data"
}

#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/build.sh"
}

function test_command_before_build() {
  # Default
  assert_exit_code 0 "$(command_before_build)"

  # Custom command
  local COMMAND_BEFORE_BUILD="echo 'test'"
  spy eval

  command_before_build

  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_build() {
  # Default
  assert_exit_code 0 "$(command_before_build)"

  # Custom command
  local COMMAND_AFTER_BUILD="echo 'test'"
  spy eval

  command_after_build

  assert_have_been_called_with "echo 'test'" eval
}

function test_n_install() {
  spy n
  n_install
  assert_have_been_called_times 0 n

  local NODE_VERSION="16"
  spy n
  n_install
  assert_have_been_called_with "install 16" n
  unset NODE_VERSION
}

function test_composer_self_update() {
  spy composer
  composer_self_update
  assert_have_been_called_times 0 composer

  local COMPOSER_VERSION="2"
  spy composer
  composer_self_update
  assert_have_been_called_with "self-update 2" composer
  unset COMPOSER_VERSION
}

function test_composer_install() {
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
  spy composer
  composer_clear_cache
  assert_have_been_called_with "clear-cache" composer
}

function test_composer_dump_autoload() {
  spy composer
  composer_dump_autoload
  assert_have_been_called_with "dump-autoload --optimize" composer
}

function test_magento_setup_di_compile() {
  mock magento echo
  spy magento

  # By default it should run
  magento_setup_di_compile
  assert_have_been_called_times 1 magento

  # If MAGENTO_DI_COMPILE is false, it should not run
  local MAGENTO_DI_COMPILE="false"
  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
  unset MAGENTO_DI_COMPILE

  # If MAGENTO_DI_COMPILE_ON_DEMAND is true, it should not run
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
  unset MAGENTO_DI_COMPILE_ON_DEMAND

  # If both is true, it should not run
  local MAGENTO_DI_COMPILE="true"
  local MAGENTO_DI_COMPILE_ON_DEMAND="true"
  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento

  # If both is false, it should not run
  local MAGENTO_DI_COMPILE="false"
  local MAGENTO_DI_COMPILE_ON_DEMAND="false"
  spy magento
  magento_setup_di_compile
  assert_have_been_called_times 0 magento
}

function test_magento_setup_static_content_deploy() {
  mock magento echo
  mock nproc echo 4
  spy magento

  # By default it should not run
  magento_setup_static_content_deploy
  assert_have_been_called_times 1 magento

  # If MAGENTO_SKIP_STATIC_CONTENT_DEPLOY is true, it should run
  local MAGENTO_SKIP_STATIC_CONTENT_DEPLOY="true"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
  unset MAGENTO_SKIP_STATIC_CONTENT_DEPLOY

  # If MAGENTO_SCD_ON_DEMAND is true, it should run
  local MAGENTO_SCD_ON_DEMAND="true"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
  unset MAGENTO_SCD_ON_DEMAND

  # If both is true, it should not run
  local MAGENTO_STATIC_CONTENT_DEPLOY="true"
  local MAGENTO_SCD_ON_DEMAND="true"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_times 0 magento
  unset MAGENTO_STATIC_CONTENT_DEPLOY
  unset MAGENTO_SCD_ON_DEMAND

  # SCD args is -fv if Magento version is 2.4+
  local MAGENTO_VERSION="2.4"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv" magento
  unset MAGENTO_VERSION

  # SCD args is -v if Magento version is 2.3
  local MAGENTO_VERSION="2.3.3"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -v" magento
  unset MAGENTO_VERSION

  # MAGENTO_LANGUAGES appended to the command
  local MAGENTO_LANGUAGES="en_US de_DE"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv en_US de_DE" magento
  unset MAGENTO_LANGUAGES

  # MAGENTO_THEMES appended to the command
  local MAGENTO_THEMES="Magento/blank Magento/luma"
  spy magento
  magento_setup_static_content_deploy
  assert_have_been_called_with "setup:static-content:deploy --jobs=4 -fv --theme=Magento/blank --theme=Magento/luma" magento
  unset MAGENTO_THEMES
}

function test_magento_create_pub_static_dir() {
  local APP_PATH="./test-data/app"
  magento_create_pub_static_dir
  assert_directory_exists "${APP_PATH}/pub/static"

  rm -fr "./test-data"
}

function test_dump_build_version() {
  local APP_PATH="./test-data/app"
  dump_build_version
  assert_file_contains "${APP_PATH}/pub/version.php" "php-version: "
  assert_file_contains "${APP_PATH}/pub/version.php" "build-date: "

  rm -fr "./test-data"
}

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

function test_command_after_build() {
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

function test_composer_install() {
  setup

  spy composer
  composer_install
  assert_have_been_called_times 0 composer

  touch composer.json
  spy composer
  composer_install
  assert_have_been_called_with "install --no-progress " composer
  rm -f composer.json

  export COMPOSER_INSTALL_ARGS="--no-dev"
  touch composer.json
  spy composer
  composer_install
  assert_have_been_called_with "install --no-progress --no-dev" composer
  rm -f composer.json
}

function test_composer_clear_cache() {
  setup

  spy composer
  composer_clear_cache
  assert_have_been_called_with "clear-cache" composer
}

function test_dump_build_version() {
  setup

  local APP_PATH="./test-data/app"

  dump_build_version
  assert_file_contains "${APP_PATH}/version.php" "php-version: "
  assert_file_contains "${APP_PATH}/version.php" "build-date: "

  rm -fr "./test-data"
}

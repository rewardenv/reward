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

function test_composer_configure() {
  # Default
  mock composer echo
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  # Test if only GITHUB_USER is set
  local GITHUB_USER="user"
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  # Test if only GITHUB_TOKEN is set
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

function test_dump_build_version() {
  local APP_PATH="./test-data/app"
  dump_build_version
  assert_file_contains "${APP_PATH}/version.php" "php-version: "
  assert_file_contains "${APP_PATH}/version.php" "build-date: "

  rm -fr "./test-data"
}

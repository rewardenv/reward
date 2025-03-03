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

function test_dump_build_version() {
  setup

  local APP_PATH="./test-data/app"

  dump_build_version
  assert_file_contains "${APP_PATH}/public/version.php" "node-version: "
  assert_file_contains "${APP_PATH}/public/version.php" "build-date: "

  rm -fr "./test-data"
}

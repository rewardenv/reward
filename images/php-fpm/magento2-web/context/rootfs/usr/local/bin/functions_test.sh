#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/functions.sh"
}

function test_log() {
  assert_matches "^[0-9]{4}(-[0-9]{2}(-[0-9]{2}(T[0-9]{2}:[0-9]{2}(:[0-9]{2})?(\.[0-9]+)?(([+-][0-9]{2}:[0-9]{2})|Z)?)?)?)? INFO: test$" "$(log 'test')"
}

function test_error() {
  assert_exit_code "1" "$(error 'test')"
}

function test_conditional_sleep() {
  spy sleep

  local SLEEP="true"
  conditional_sleep
  assert_have_been_called_with "infinity" sleep

  local SLEEP="5"
  conditional_sleep
  assert_have_been_called_with "5" sleep
}

function test_lock() {
  # Test lock_acquire without lockfile
  assert_exit_code "1" "$(lock_acquire)"

  # Test lock_acquire with lockfile
  lock_acquire "test-data/lock"
  assert_file_exists "test-data/lock"

  # Test lock_acquire with existing lockfile
  assert_exit_code "1" "$(lock_acquire 'test-data/lock')"

  # Release lock
  lock_release "test-data/lock"
  assert_file_not_exists "test-data/lock"

  rm -fr "./test-data"
}

function test_shared_config_path() {
  local SHARED_CONFIG_PATH="/config"
  assert_equals "/tmp" "$(shared_config_path)"
  unset SHARED_CONFIG_PATH

  local TEST_PATH="test-data/config"

  mkdir -p "${TEST_PATH}"
  chmod 777 "${TEST_PATH}"
  local SHARED_CONFIG_PATH="${TEST_PATH}"
  assert_equals "test-data/config" "$(shared_config_path)"
  rm -fr "./test-data"

  # Skip the test if the caller is the root user as it has write permissions to everything
  if [[ "$(id -u)" == "0" ]]; then
    skip && return
  fi

  mkdir -p "${TEST_PATH}"
  chmod 444 "${TEST_PATH}"
  local SHARED_CONFIG_PATH="${TEST_PATH}"
  assert_equals "/tmp" "$(shared_config_path)"
  rm -fr "./test-data"
}

function test_app_path() {
  assert_equals "/var/www/html" "$(app_path)"

  local APP_PATH="/app"
  assert_equals "/app" "$(app_path)"
}

function version_gt() {
  assert_true "$(version_gt '2.4.4' '2.3.99')"
  assert_false "$(version_gt '2.3.99' '2.4.4')"
  assert_false "$(version_gt '2.4.4' '2.4.4')"
  assert_true "$(version_gt '2.4' '2.3.99')"
  assert_false "$(version_gt '2.3.99' '2.4')"
  assert_false "$(version_gt '2.4' '2.4')"
}

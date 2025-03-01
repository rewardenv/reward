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

function test_version_gt() {
  assert_exit_code 0 "$(version_gt '2.4.4' '2.3.99')"
  assert_exit_code 1 "$(version_gt '2.3.99' '2.4.4')"
  assert_exit_code 1 "$(version_gt '2.4.4' '2.4.4')"
  assert_exit_code 0 "$(version_gt '2.4' '2.3.99')"
  assert_exit_code 1 "$(version_gt '2.3.99' '2.4')"
  assert_exit_code 1 "$(version_gt '2.4' '2.4')"
  assert_exit_code 0 "$(version_gt 'v2.4.4' '2.3.99')"
  assert_exit_code 1 "$(version_gt '2.3.99' 'v2.4.4')"
  assert_exit_code 1 "$(version_gt 'v2.4.4' 'v2.4.4')"
  assert_exit_code 0 "$(version_gt '2.4' 'v2.3.99')"
  assert_exit_code 1 "$(version_gt 'v2.3.99' '2.4')"
  assert_exit_code 1 "$(version_gt 'v2.4' '2.4')"

  if version_gt '2.3.99' '2.4.0'; then
    fail
  fi
  if ! version_gt '2.4.0' 'v2.3.99.99'; then
    fail
  fi
}

function test_run_hooks() {
  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}/hooks/test.d"
  printf "#!/bin/bash\necho 'test-123'" >"${APP_PATH}/hooks/test.d/01-test.sh"
  assert_contains "test-123" "$(run_hooks 'test')"
  rm -fr "./test-data"
}

function test_check_timeout() {
  export TIMEOUT="0"
  assert_exit_code 1 "$(check_timeout)"

  export TIMEOUT="10"
  assert_exit_code 0 "$(check_timeout)"
}

function test_check_dependency() {
  local TIMEOUT=1
  local checks=("check_database")
  mock check_database true
  assert_exit_code 0 "$(check_dependency 'check_database')"

  local checks=("check_database")
  mock check_database false
  assert_exit_code 1 "$(check_dependency 'check_database')"
}

#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/pre-stop.sh"
}

function test_lock_deploy() {
  # Test with the default SHARED_CONFIG_PATH
  lock_deploy
  assert_file_exists '/tmp/.deploy.lock'
  rm -f '/tmp/.deploy.lock'

  # Test with a custom SHARED_CONFIG_PATH
  export SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  lock_deploy
  assert_file_exists './test-data/config/.deploy.lock'
  rm -fr "./test-data"
}

function test_run_hooks() {
  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}/hooks/pre-stop.d"
  printf "#!/bin/bash\necho 'test-123'" >"${APP_PATH}/hooks/pre-stop.d/01-test.sh"
  assert_contains "test-123" "$(main)"
  rm -fr "./test-data"
}

#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/post-start.sh"
}

function test_create_symlink() {
  # Test with a valid SHARED_CONFIG_PATH
  export SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}/app"
  touch "${SHARED_CONFIG_PATH}/wp-config.php"

  export APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"

  create_symlink
  assert_exit_code 0 "$(test -L './test-data/var/www/html/wp-config.php')"

  rm -fr "./test-data"
}

function test_run_hooks() {
  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}/hooks/post-start.d"
  printf "#!/bin/bash\necho 'test-123'" >"${APP_PATH}/hooks/post-start.d/01-test.sh"
  assert_contains "test-123" "$(main)"
  rm -fr "./test-data"
}

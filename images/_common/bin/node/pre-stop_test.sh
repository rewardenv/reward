#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/pre-stop.sh"
}

function test_run_hooks() {
  local APP_PATH="./test-data/app"
  setup

  mkdir -p "${APP_PATH}/hooks/pre-stop.d"
  printf "#!/bin/bash\necho 'test-123'" >"${APP_PATH}/hooks/pre-stop.d/01-test.sh"
  assert_contains "test-123" "$(main)"
  rm -fr "./test-data"
}

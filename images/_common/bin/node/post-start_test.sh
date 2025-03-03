#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/post-start.sh"
}

function test_run_hooks() {
  setup

  local APP_PATH="./test-data/app"
  mkdir -p "${APP_PATH}/hooks/post-start.d"
  printf "#!/bin/bash\necho 'test-123'" >"${APP_PATH}/hooks/post-start.d/01-test.sh"
  assert_contains "test-123" "$(main)"
  rm -fr "./test-data"
}

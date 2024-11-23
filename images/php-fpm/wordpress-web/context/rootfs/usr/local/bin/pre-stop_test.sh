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

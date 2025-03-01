#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/check-dependencies.sh"
}

function test_configure_checks_default() {
  local checks=("")
  setup

  configure_checks
  assert_array_contains "check_database" "${checks[@]}"
}

function test_configure_checks_with_services() {
  local checks=("")
  local WORDPRESS_REDIS_ENABLED="true"
  setup

  configure_checks
  assert_array_contains "check_database" "${checks[@]}"
  assert_array_contains "check_redis" "${checks[@]}"
}

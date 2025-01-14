#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/check-dependencies.sh"
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

function test_configure_checks() {
  local checks=("")
  configure_checks
  assert_array_contains "check_database" "${checks[@]}"

  local checks=("")
  local SHOPWARE_ELASTICSEARCH_ENABLED="true"
  local SHOPWARE_OPENSEARCH_ENABLED="true"
  local SHOPWARE_REDIS_ENABLED="true"
  local SHOPWARE_RABBITMQ_ENABLED="true"
  local SHOPWARE_VARNISH_ENABLED="true"
  configure_checks
  assert_array_contains "check_database" "${checks[@]}"
  assert_array_contains "check_elasticsearch" "${checks[@]}"
  assert_array_contains "check_opensearch" "${checks[@]}"
  assert_array_contains "check_redis" "${checks[@]}"
  assert_array_contains "check_rabbitmq" "${checks[@]}"
  assert_array_contains "check_varnish" "${checks[@]}"
}

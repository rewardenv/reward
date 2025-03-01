#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/check-dependencies.sh"
}

function test_configure_checks_default() {
  local checks=("")
  setup

  configure_checks
  assert_array_contains "check_database" "${checks[@]}"
  assert_array_not_contains "check_elasticsearch" "${checks[@]}"
  assert_array_not_contains "check_opensearch" "${checks[@]}"
  assert_array_not_contains "check_redis" "${checks[@]}"
  assert_array_not_contains "check_rabbitmq" "${checks[@]}"
  assert_array_not_contains "check_varnish" "${checks[@]}"
}

function test_configure_checks_with_services() {
  local checks=("")
  local MAGENTO_ELASTICSEARCH_ENABLED="true"
  local MAGENTO_OPENSEARCH_ENABLED="true"
  local MAGENTO_REDIS_ENABLED="true"
  local MAGENTO_RABBITMQ_ENABLED="true"
  local MAGENTO_VARNISH_ENABLED="true"
  setup

  configure_checks
  assert_array_contains "check_database" "${checks[@]}"
  assert_array_contains "check_elasticsearch" "${checks[@]}"
  assert_array_contains "check_opensearch" "${checks[@]}"
  assert_array_contains "check_redis" "${checks[@]}"
  assert_array_contains "check_rabbitmq" "${checks[@]}"
  assert_array_contains "check_varnish" "${checks[@]}"
}

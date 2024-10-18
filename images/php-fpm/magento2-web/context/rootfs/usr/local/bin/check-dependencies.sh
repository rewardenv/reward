#!/usr/bin/env bash
[ "${DEBUG:-false}" = "true" ] && set -x
set -eE -o pipefail -o errtrace
shopt -s extdebug

if [ "${DEBUG:-false}" != "true" ]; then
  QUIET="--quiet"
fi

log() {
  [ "${SILENT:-false}" != "true" ] && printf "%s INFO: %s\n" "$(date --iso-8601=seconds)" "$*"
}

error() {
  exitcode=$?
  # color red
  printf "\033[1;31m%s ERROR: %s\033[0m\n" "$(date --iso-8601=seconds)" "$*"

  if [ "${exitcode}" -eq 0 ]; then exit 1; fi
  exit "${exitcode}"
}

trapinfo() {
  # shellcheck disable=SC2145
  error "Command failed: $BASH_COMMAND STATUS=$? LINENO=${@:1:$(($# - 1))}"
}

check_database() {
  log "Checking database connection..."
  if mysql -h"${MAGENTO_DATABASE_HOST:-db}" -P"${MAGENTO_DATABASE_PORT:-3306}" -u"${MAGENTO_DATABASE_USER:-magento}" -p"${MAGENTO_DATABASE_PASSWORD:-magento}" -e "CREATE DATABASE IF NOT EXISTS ${MAGENTO_DATABASE_NAME:-magento}; "; then
    log "Database connection ready."
  else
    error "Database connection failed."
  fi
}

check_elasticsearch() {
  if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-false}" = "true" ]; then
    log "Checking Elasticsearch connection..."
    if curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_ELASTICSEARCH_HOST:-opensearch}:${MAGENTO_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
      log "Elasticsearch connection ready."
    else
      error "Elasticsearch connection failed."
    fi
  fi
}

check_opensearch() {
  if [ "${MAGENTO_OPENSEARCH_ENABLED:-false}" = "true" ]; then
    log "Checking Opensearch connection..."
    if curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_OPENSEARCH_HOST:-opensearch}:${MAGENTO_OPENSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
      log "Opensearch connection ready."
    else
      error "Opensearch connection failed."
    fi
  fi
}

check_redis() {
  if [ "${MAGENTO_REDIS_ENABLED:-false}" = "true" ]; then
    log "Checking Redis connection..."
    AUTH_COMMAND=""
    if [ -n "${MAGENTO_REDIS_PASSWORD:-}" ]; then
      AUTH_COMMAND="AUTH ${MAGENTO_REDIS_PASSWORD:-redis}\r\n"
    fi

    if printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${MAGENTO_REDIS_HOST:-redis}" "${MAGENTO_REDIS_PORT:-6379}" | grep "PONG"; then
      log "Redis connection ready."
    else
      error "Redis connection failed."
    fi
  fi
}

check_rabbitmq() {
  if [ "${MAGENTO_RABBITMQ_ENABLED:-false}" = "true" ]; then
    log "Checking RabbitMQ connection..."
    if nc -v -z "${MAGENTO_AMQP_HOST:-rabbitmq}" "${MAGENTO_AMQP_PORT:-5672}"; then
      log "RabbitMQ connection ready."
    else
      error "RabbitMQ connection failed."
    fi
  fi
}

check_varnish() {
  if [ "${MAGENTO_VARNISH_ENABLED:-false}" = "true" ]; then
    log "Checking Varnish connection..."
    if nc -v -z "${MAGENTO_VARNISH_HOST:-varnish}" "${MAGENTO_VARNISH_PORT:-80}"; then
      log "Varnish connection ready."
    else
      error "Varnish connection failed."
    fi
  fi
}

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  check_database
  check_elasticsearch
  check_opensearch
  check_redis
  check_rabbitmq
  check_varnish
  log "All connections are ready."
}

main

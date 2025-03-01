#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eEu -o pipefail -o errtrace
shopt -s extdebug

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
FUNCTIONS_FILE="${SCRIPT_DIR}/../lib/functions.sh"
for lib_path in \
  "${SCRIPT_DIR}/../lib/functions.sh" \
  "${SCRIPT_DIR}/../../lib/functions.sh" \
  "${HOME}/.local/lib/functions.sh" \
  "/usr/local/lib/functions.sh" \
  "$(command -v functions.sh)"; do
  if [[ -f "${lib_path}" ]]; then
    FUNCTIONS_FILE="${lib_path}"
    break
  fi
done

if [[ -f "${FUNCTIONS_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${FUNCTIONS_FILE}"
else
  printf "\033[1;31m%s ERROR: Required file %s not found\033[0m\n" "$(date --iso-8601=seconds)" "${FUNCTIONS_FILE}" >&2
  exit 1
fi

: "${SHOPWARE_DATABASE_HOST:=db}"
: "${SHOPWARE_DATABASE_PORT:=3306}"
: "${SHOPWARE_DATABASE_USER:=shopware}"
: "${SHOPWARE_DATABASE_PASSWORD:=shopware}"
: "${SHOPWARE_DATABASE_NAME:=shopware}"
: "${SHOPWARE_ELASTICSEARCH_ENABLED:=false}"
: "${SHOPWARE_ELASTICSEARCH_HOST:=elasticsearch}"
: "${SHOPWARE_ELASTICSEARCH_PORT:=9200}"
: "${SHOPWARE_OPENSEARCH_ENABLED:=false}"
: "${SHOPWARE_OPENSEARCH_HOST:=opensearch}"
: "${SHOPWARE_OPENSEARCH_PORT:=9200}"
: "${SHOPWARE_REDIS_ENABLED:=false}"
: "${SHOPWARE_REDIS_HOST:=redis}"
: "${SHOPWARE_REDIS_PORT:=6379}"
: "${SHOPWARE_REDIS_PASSWORD:=}"
: "${SHOPWARE_RABBITMQ_ENABLED:=false}"
: "${SHOPWARE_AMQP_HOST:=rabbitmq}"
: "${SHOPWARE_AMQP_PORT:=5672}"
: "${SHOPWARE_VARNISH_ENABLED:=false}"
: "${SHOPWARE_VARNISH_HOST:=varnish}"
: "${SHOPWARE_VARNISH_PORT:=80}"

check_database() {
  check_command mysql

  if ! mysql -h"${SHOPWARE_DATABASE_HOST}" -P"${SHOPWARE_DATABASE_PORT}" -u"${SHOPWARE_DATABASE_USER}" -p"${SHOPWARE_DATABASE_PASSWORD}" -e "CREATE DATABASE IF NOT EXISTS ${SHOPWARE_DATABASE_NAME}; "; then
    return 1
  fi
}

check_elasticsearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${SHOPWARE_ELASTICSEARCH_HOST}:${SHOPWARE_ELASTICSEARCH_PORT}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_opensearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${SHOPWARE_OPENSEARCH_HOST}:${SHOPWARE_OPENSEARCH_PORT}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_redis() {
  check_command nc

  AUTH_COMMAND=""
  if [[ -n "${SHOPWARE_REDIS_PASSWORD}" ]]; then
    AUTH_COMMAND="AUTH ${SHOPWARE_REDIS_PASSWORD}\r\n"
  fi

  if ! printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${SHOPWARE_REDIS_HOST}" "${SHOPWARE_REDIS_PORT}" | grep "PONG"; then
    return 1
  fi
}

check_rabbitmq() {
  check_command nc

  if ! nc -v -z "${SHOPWARE_AMQP_HOST}" "${SHOPWARE_AMQP_PORT}"; then
    return 1
  fi
}

check_varnish() {
  check_command nc

  if ! nc -v -z "${SHOPWARE_VARNISH_HOST}" "${SHOPWARE_VARNISH_PORT}"; then
    return 1
  fi
}

configure_checks() {
  checks+=("check_database")

  if [[ "${SHOPWARE_ELASTICSEARCH_ENABLED}" == "true" ]]; then
    checks+=("check_elasticsearch")
  fi

  if [[ "${SHOPWARE_OPENSEARCH_ENABLED}" == "true" ]]; then
    checks+=("check_opensearch")
  fi

  if [[ "${SHOPWARE_REDIS_ENABLED}" == "true" ]]; then
    checks+=("check_redis")
  fi

  if [[ "${SHOPWARE_RABBITMQ_ENABLED}" == "true" ]]; then
    checks+=("check_rabbitmq")
  fi

  if [[ "${SHOPWARE_VARNISH_ENABLED}" == "true" ]]; then
    checks+=("check_varnish")
  fi
}

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  declare -a checks
  declare -A check_status
  configure_checks

  for check in "${checks[@]}"; do
    if ! check_dependency "$check"; then
      log "Dependency checks failed due to timeout or error"
      exit 1
    fi
  done

  log "All dependency checks passed"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

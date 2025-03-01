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

: "${MAGENTO_DATABASE_HOST:=db}"
: "${MAGENTO_DATABASE_PORT:=3306}"
: "${MAGENTO_DATABASE_USER:=magento}"
: "${MAGENTO_DATABASE_PASSWORD:=magento}"
: "${MAGENTO_DATABASE_NAME:=magento}"
: "${MAGENTO_ELASTICSEARCH_ENABLED:=false}"
: "${MAGENTO_ELASTICSEARCH_HOST:=elasticsearch}"
: "${MAGENTO_ELASTICSEARCH_PORT:=9200}"
: "${MAGENTO_OPENSEARCH_ENABLED:=false}"
: "${MAGENTO_OPENSEARCH_HOST:=opensearch}"
: "${MAGENTO_OPENSEARCH_PORT:=9200}"
: "${MAGENTO_REDIS_ENABLED:=false}"
: "${MAGENTO_REDIS_HOST:=redis}"
: "${MAGENTO_REDIS_PORT:=6379}"
: "${MAGENTO_REDIS_PASSWORD:=}"
: "${MAGENTO_RABBITMQ_ENABLED:=false}"
: "${MAGENTO_AMQP_HOST:=rabbitmq}"
: "${MAGENTO_AMQP_PORT:=5672}"
: "${MAGENTO_VARNISH_ENABLED:=false}"
: "${MAGENTO_VARNISH_HOST:=varnish}"
: "${MAGENTO_VARNISH_PORT:=80}"

check_database() {
  check_command mysql

  if ! mysql -h"${MAGENTO_DATABASE_HOST}" -P"${MAGENTO_DATABASE_PORT}" -u"${MAGENTO_DATABASE_USER}" -p"${MAGENTO_DATABASE_PASSWORD}" -e "CREATE DATABASE IF NOT EXISTS ${MAGENTO_DATABASE_NAME}; "; then
    return 1
  fi
}

check_elasticsearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_ELASTICSEARCH_HOST}:${MAGENTO_ELASTICSEARCH_PORT}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_opensearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_OPENSEARCH_HOST}:${MAGENTO_OPENSEARCH_PORT}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_redis() {
  check_command nc

  AUTH_COMMAND=""
  if [[ -n "${MAGENTO_REDIS_PASSWORD}" ]]; then
    AUTH_COMMAND="AUTH ${MAGENTO_REDIS_PASSWORD}\r\n"
  fi

  if ! printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${MAGENTO_REDIS_HOST}" "${MAGENTO_REDIS_PORT}" | grep "PONG"; then
    return 1
  fi
}

check_rabbitmq() {
  check_command nc

  if ! nc -v -z "${MAGENTO_AMQP_HOST}" "${MAGENTO_AMQP_PORT}"; then
    return 1
  fi
}

check_varnish() {
  check_command nc

  if ! nc -v -z "${MAGENTO_VARNISH_HOST}" "${MAGENTO_VARNISH_PORT}"; then
    return 1
  fi
}

configure_checks() {
  checks+=("check_database")

  if [[ "${MAGENTO_ELASTICSEARCH_ENABLED}" == "true" ]]; then
    checks+=("check_elasticsearch")
  fi

  if [[ "${MAGENTO_OPENSEARCH_ENABLED}" == "true" ]]; then
    checks+=("check_opensearch")
  fi

  if [[ "${MAGENTO_REDIS_ENABLED}" == "true" ]]; then
    checks+=("check_redis")
  fi

  if [[ "${MAGENTO_RABBITMQ_ENABLED}" == "true" ]]; then
    checks+=("check_rabbitmq")
  fi

  if [[ "${MAGENTO_VARNISH_ENABLED}" == "true" ]]; then
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

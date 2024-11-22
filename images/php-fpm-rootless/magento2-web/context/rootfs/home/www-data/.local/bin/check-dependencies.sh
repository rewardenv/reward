#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eE -o pipefail
shopt -s extdebug

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
FUNCTIONS_FILE="${SCRIPT_DIR}/functions.sh"
if [[ ! -f "${FUNCTIONS_FILE}" ]]; then
  FUNCTIONS_FILE="$(command -v functions.sh)"
fi
readonly FUNCTIONS_FILE

if [[ -f "${FUNCTIONS_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${FUNCTIONS_FILE}"
else
  printf "\033[1;31m%s ERROR: Required file %s not found\033[0m\n" "$(date --iso-8601=seconds)" "${FUNCTIONS_FILE}" >&2
  exit 1
fi

START_TIME=$(date +%s)
readonly START_TIME
RETRY_INTERVAL=${RETRY_INTERVAL:-1}
TIMEOUT=${TIMEOUT:-600}

# Check if a command exists
check_timeout() {
  local current_time
  current_time=$(date +%s)
  local elapsed=$((current_time - START_TIME))

  if [[ "$elapsed" -ge "$TIMEOUT" ]]; then
    error "Global timeout of ${TIMEOUT}s reached"
  fi
}

# Main check function with retry logic
check_dependency() {
  local func_name="$1"
  local attempt=1

  # Skip if already successful
  if [[ ${check_status[$func_name]:-false} == true ]]; then
    log "Skipping $func_name - already successful"
    return 0
  fi

  while check_timeout; do
    log "Checking $func_name (attempt $attempt)"

    if ${func_name}; then
      log "Check succeeded: ${func_name}"
      check_status[${func_name}]=true
      return 0
    fi

    attempt=$((attempt + 1))
    log "Check failed: ${func_name}, retrying in ${RETRY_INTERVAL}s..."
    sleep "${RETRY_INTERVAL}"
  done

  # If we get here, we've timed out
  log "Global Timeout reached ${func_name}"
  return 1
}

check_database() {
  check_command mysql

  if ! mysql -h"${MAGENTO_DATABASE_HOST:-db}" -P"${MAGENTO_DATABASE_PORT:-3306}" -u"${MAGENTO_DATABASE_USER:-magento}" -p"${MAGENTO_DATABASE_PASSWORD:-magento}" -e "CREATE DATABASE IF NOT EXISTS ${MAGENTO_DATABASE_NAME:-magento}; "; then
    return 1
  fi
}

check_elasticsearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_ELASTICSEARCH_HOST:-elasticsearch}:${MAGENTO_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_opensearch() {
  check_command curl

  if ! curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_OPENSEARCH_HOST:-opensearch}:${MAGENTO_OPENSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_redis() {
  check_command nc

  AUTH_COMMAND=""
  if [[ -n "${MAGENTO_REDIS_PASSWORD:-}" ]]; then
    AUTH_COMMAND="AUTH ${MAGENTO_REDIS_PASSWORD:-redis}\r\n"
  fi

  if ! printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${MAGENTO_REDIS_HOST:-redis}" "${MAGENTO_REDIS_PORT:-6379}" | grep "PONG"; then
    return 1
  fi
}

check_rabbitmq() {
  check_command nc

  if ! nc -v -z "${MAGENTO_AMQP_HOST:-rabbitmq}" "${MAGENTO_AMQP_PORT:-5672}"; then
    return 1
  fi
}

check_varnish() {
  check_command nc

  if ! nc -v -z "${MAGENTO_VARNISH_HOST:-varnish}" "${MAGENTO_VARNISH_PORT:-80}"; then
    return 1
  fi
}

configure_checks() {
  checks+=("check_database")

  if [[ "${MAGENTO_ELASTICSEARCH_ENABLED:-false}" == "true" ]]; then
    checks+=("check_elasticsearch")
  fi

  if [[ "${MAGENTO_OPENSEARCH_ENABLED:-false}" == "true" ]]; then
    checks+=("check_opensearch")
  fi

  if [[ "${MAGENTO_REDIS_ENABLED:-false}" == "true" ]]; then
    checks+=("check_redis")
  fi

  if [[ "${MAGENTO_RABBITMQ_ENABLED:-false}" == "true" ]]; then
    checks+=("check_rabbitmq")
  fi

  if [[ "${MAGENTO_VARNISH_ENABLED:-false}" == "true" ]]; then
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

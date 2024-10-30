#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eE -o pipefail -o errtrace
shopt -s extdebug

FUNCTIONS_FILE="$(dirname "$(realpath "$0")")/functions.sh"
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
readonly RETRY_INTERVAL=${RETRY_INTERVAL:-1}
readonly TIMEOUT=${TIMEOUT:-600}

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
  if ! mysql -h"${SHOPWARE_DATABASE_HOST:-db}" -P"${SHOPWARE_DATABASE_PORT:-3306}" -u"${SHOPWARE_DATABASE_USER:-shopware}" -p"${SHOPWARE_DATABASE_PASSWORD:-shopware}" -e "CREATE DATABASE IF NOT EXISTS ${SHOPWARE_DATABASE_NAME:-shopware}; "; then
    return 1
  fi
}

check_elasticsearch() {
  if ! curl --connect-timeout 10 -fsSL -X GET "http://${SHOPWARE_ELASTICSEARCH_HOST:-elasticsearch}:${SHOPWARE_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_opensearch() {
  if ! curl --connect-timeout 10 -fsSL -X GET "http://${SHOPWARE_OPENSEARCH_HOST:-opensearch}:${SHOPWARE_OPENSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
    return 1
  fi
}

check_redis() {
  AUTH_COMMAND=""
  if [[ -n "${SHOPWARE_REDIS_PASSWORD:-}" ]]; then
    AUTH_COMMAND="AUTH ${SHOPWARE_REDIS_PASSWORD:-redis}\r\n"
  fi

  if ! printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${SHOPWARE_REDIS_HOST:-redis}" "${SHOPWARE_REDIS_PORT:-6379}" | grep "PONG"; then
    return 1
  fi
}

check_rabbitmq() {
  if ! nc -v -z "${SHOPWARE_AMQP_HOST:-rabbitmq}" "${SHOPWARE_AMQP_PORT:-5672}"; then
    return 1
  fi
}

check_varnish() {
  if ! nc -v -z "${SHOPWARE_VARNISH_HOST:-varnish}" "${SHOPWARE_VARNISH_PORT:-80}"; then
    return 1
  fi
}

configure_checks() {
  checks+=("check_database")

  if [[ "${SHOPWARE_ELASTICSEARCH_ENABLED:-false}" == "true" ]]; then
    checks+=("check_elasticsearch")
  fi

  if [[ "${SHOPWARE_OPENSEARCH_ENABLED:-false}" == "true" ]]; then
    checks+=("check_opensearch")
  fi

  if [[ "${SHOPWARE_REDIS_ENABLED:-false}" == "true" ]]; then
    checks+=("check_redis")
  fi

  if [[ "${SHOPWARE_RABBITMQ_ENABLED:-false}" == "true" ]]; then
    checks+=("check_rabbitmq")
  fi

  if [[ "${SHOPWARE_VARNISH_ENABLED:-false}" == "true" ]]; then
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

main

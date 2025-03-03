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

check_redis() {
  check_command nc

  AUTH_COMMAND=""
  if [[ -n "${NODE_REDIS_PASSWORD:-}" ]]; then
    AUTH_COMMAND="AUTH ${NODE_REDIS_PASSWORD:-redis}\r\n"
  fi

  if ! printf "%bPING\r\n" "${AUTH_COMMAND}" | nc -N -v "${NODE_REDIS_HOST:-redis}" "${NODE_REDIS_PORT:-6379}" | grep "PONG"; then
    return 1
  fi
}

configure_checks() {
  if [[ "${NODE_REDIS_ENABLED:-false}" == "true" ]]; then
    checks+=("check_redis")
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

#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eEu -o pipefail -o errtrace
shopt -s extdebug

FUNCTIONS_FILE="$(dirname "$(realpath "${BASH_SOURCE[0]}")")/functions.sh"
readonly FUNCTIONS_FILE
if [[ -f "${FUNCTIONS_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${FUNCTIONS_FILE}"
else
  printf "\033[1;31m%s ERROR: Required file %s not found\033[0m\n" "$(date --iso-8601=seconds)" "${FUNCTIONS_FILE}" >&2
  exit 1
fi

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  touch "$(shared_config_path)/.deploy.lock"
}

main
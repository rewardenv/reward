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

: "${COMMAND_BEFORE_BUILD:=}"
: "${COMMAND_AFTER_BUILD:=}"

: "${NODE_PACKAGE_MANAGER_COMMAND:=npm}"
: "${NODE_INSTALL_ARGS:=install --prefer-offline --no-audit --progress=false}"
: "${NODE_BUILD_ARGS:=run build}"
: "${COMMAND_BEFORE_NODE_INSTALL:=}"
: "${COMMAND_AFTER_NODE_INSTALL:=}"
: "${NODE_COMMAND_BEFORE_BUILD:=}"
: "${NODE_COMMAND_AFTER_BUILD:=}"

command_before_build() {
  if [[ -z "${COMMAND_BEFORE_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command before the whole build process"
  eval "${COMMAND_BEFORE_BUILD}"
}

command_after_build() {
  if [[ -z "${COMMAND_AFTER_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command after the whole build process"
  eval "${COMMAND_AFTER_BUILD}"
}

command_before_node_build() {
  if [[ -z "${COMMAND_BEFORE_NODE_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command before node build"
  eval "${COMMAND_BEFORE_NODE_BUILD}"
}

command_after_node_build() {
  if [[ -z "${COMMAND_AFTER_NODE_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command after node build"
  eval "${COMMAND_AFTER_NODE_BUILD}"
}

command_before_node_install() {
  if [[ -z "${COMMAND_BEFORE_NODE_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command before node build"
  eval "${COMMAND_BEFORE_NODE_INSTALL}"
}

command_after_node_install() {
  if [[ -z "${COMMAND_AFTER_NODE_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command after node build"
  eval "${COMMAND_AFTER_NODE_INSTALL}"
}

node_build() {
  command_before_node_build

  log "Building node dependencies"
  eval "${NODE_PACKAGE_MANAGER_COMMAND} ${NODE_BUILD_ARGS}"

  command_after_node_build
}

node_install() {
  if [[ -z "${NODE_INSTALL_ARGS}" ]]; then
    return 0
  fi

  command_before_node_install

  log "Installing node dependencies"
  eval "${NODE_PACKAGE_MANAGER_COMMAND} ${NODE_INSTALL_ARGS}"

  command_after_node_install
}

dump_build_version() {
  log "Creating build version file"
  mkdir -p "$(app_path)/public"
  printf "<html>node-version: $(node --version) </br>\nbuild-date: $(date '+%Y/%m/%d %H:%M:%S')\n</html>" >"$(app_path)/public/version.html"
}

main() {
  run_hooks "pre-build"

  command_before_build

  node_install

  node_build

  dump_build_version

  command_after_build

  run_hooks "post-build"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eEu -o pipefail -o errtrace
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

PHP_ARGS="-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} --memory_limit=${PHP_MEMORY_LIMIT:-2G}"

_console_command="bin/ci"
CONSOLE_COMMAND="${CONSOLE_COMMAND:-php ${PHP_ARGS} ${_console_command} --no-ansi --no-interaction}"
readonly CONSOLE_COMMAND
unset _console_command

_composer_command="composer"
if command -v composer 2>/dev/null; then
  _composer_command="$(command -v composer 2>/dev/null)"
fi
COMPOSER_COMMAND="${COMPOSER_COMMAND:-php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
readonly COMPOSER_COMMAND
unset _composer_command

_n_command="n"
if command -v n 2>/dev/null; then
  _n_command="$(command -v n 2>/dev/null)"
fi
N_COMMAND="${N_COMMAND:-${_n_command}}"
unset _n_command

console() {
  ${CONSOLE_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

n() {
  ${N_COMMAND} "$@"
}

check_requirements() {
  check_command "composer"
  check_command "n"
}

command_before_build() {
  if [[ -z "${COMMAND_BEFORE_BUILD:-}" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "${COMMAND_BEFORE_BUILD:-}"
}

command_after_build() {
  if [[ -z "${COMMAND_AFTER_BUILD:-}" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "${COMMAND_AFTER_BUILD:-}"
}

n_install() {
  if [[ -z "${NODE_VERSION:-}" ]]; then
    return 0
  fi

  log "Installing Node.js version ${NODE_VERSION:-}"
  n install "${NODE_VERSION:-}"
}

composer_self_update() {
  if [[ -z "${COMPOSER_VERSION:-}" ]]; then
    return 0
  fi

  log "Self-updating Composer to version ${COMPOSER_VERSION:-}"
  composer self-update "${COMPOSER_VERSION:-}"
}

composer_configure() {
  log "Configuring Composer"

  if [[ -n "${GITHUB_USER:-}" ]] && [[ -n "${GITHUB_TOKEN:-}" ]]; then
    composer global config http-basic.github.com "${GITHUB_USER:-}" "${GITHUB_TOKEN:-}"
  fi

  if [[ -n "${BITBUCKET_PUBLIC_KEY:-}" ]] && [[ -n "${BITBUCKET_PRIVATE_KEY:-}" ]]; then
    composer global config bitbucket-oauth.bitbucket.org "${BITBUCKET_PUBLIC_KEY:-}" "${BITBUCKET_PRIVATE_KEY:-}"
  fi

  if [[ -n "${GITLAB_TOKEN:-}" ]]; then
    composer global config gitlab-token.gitlab.com "${GITLAB_TOKEN:-}"
  fi
}

composer_install() {
  if [[ ! -f "composer.json" ]]; then
    return 0
  fi

  log "Installing Composer dependencies"
  composer install --no-progress
}

composer_clear_cache() {
  log "Clearing Composer cache"
  composer clear-cache
}

shopware_remove_env_file() {
  log "Removing .env file"
  rm -f "$(app_path)/.env"
}

shopware_bundle_dump() {
  log "Dumping Shopware bundles"
  console bundle:dump
}

shopware_build() {
  if [[ "${SHOPWARE_BUILD_STOREFRONT:-true}" == "true" ]] && [[ -f "$(app_path)/bin/build-storefront.sh" ]]; then
    export CI=1
    export SHOPWARE_SKIP_THEME_COMPILE=true
    export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
    "$(app_path)/bin/build-storefront.sh"
  fi
}

dump_build_version() {
  log "Creating build version file"
  mkdir -p "$(app_path)/public"
  printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" >"$(app_path)/public/version.php"
}

main() {
  run_hooks "pre-build"

  check_requirements

  command_before_build

  n_install
  composer_self_update

  composer_configure
  composer_install
  composer_clear_cache
  shopware_remove_env_file
  shopware_bundle_dump
  shopware_build
  dump_build_version

  command_after_build

  run_hooks "post-build"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

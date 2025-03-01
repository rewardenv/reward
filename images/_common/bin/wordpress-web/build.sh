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


_wordpress_command=wp
if command -v wp &>/dev/null; then
  _wordpress_command="$(command -v wp 2>/dev/null)"
fi

_composer_command="composer"
if command -v composer &>/dev/null; then
  _composer_command="$(command -v composer 2>/dev/null)"
fi

_n_command="n"
if command -v n &>/dev/null; then
  _n_command="$(command -v n 2>/dev/null)"
fi

: "${PHP_ARGS:=-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} -dmemory_limit=${PHP_MEMORY_LIMIT:-2G}}"
: "${WORDPRESS_COMMAND:=php ${PHP_ARGS} ${_wordpress_command} --no-color}"
: "${COMPOSER_COMMAND:=php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
: "${N_COMMAND:=${_n_command}}"

unset PHP_ARGS _wordpress_command _composer_command _n_command

: "${COMMAND_BEFORE_BUILD:=}"
: "${COMMAND_AFTER_BUILD:=}"
: "${NODE_VERSION:=}"
: "${COMPOSER_VERSION:=}"
: "${COMPOSER_AUTH:=}"
: "${GITHUB_USER:=}"
: "${GITHUB_TOKEN:=}"
: "${BITBUCKET_PUBLIC_KEY:=}"
: "${BITBUCKET_PRIVATE_KEY:=}"
: "${GITLAB_TOKEN:=}"

wp() {
  ${WORDPRESS_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

n() {
  ${N_COMMAND} "$@"
}

check_requirements() {
  check_command "wp"
  check_command "composer"
  check_command "n"
}

command_before_build() {
  if [[ -z "${COMMAND_BEFORE_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "${COMMAND_BEFORE_BUILD}"
}

command_after_build() {
  if [[ -z "${COMMAND_AFTER_BUILD}" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "${COMMAND_AFTER_BUILD}"
}

n_install() {
  if [[ -z "${NODE_VERSION}" ]]; then
    return 0
  fi

  log "Installing Node.js version ${NODE_VERSION}"
  n install "${NODE_VERSION}"
}

composer_self_update() {
  if [[ -z "${COMPOSER_VERSION}" ]]; then
    return 0
  fi

  log "Self-updating Composer to version ${COMPOSER_VERSION}"

  case "${COMPOSER_VERSION}" in
  "1")
    composer self-update --1
    ;;
  "2")
    composer self-update --2
    ;;
  "2.2")
    composer self-update --2.2
    ;;
  "stable")
    composer self-update --stable
    ;;
  *)
    composer self-update "${COMPOSER_VERSION}"
    ;;
  esac
}

composer_configure() {
  if [[ -n "${COMPOSER_AUTH}" ]]; then
    # HACK: workaround for
    # https://github.com/composer/composer/issues/12084
    # shellcheck disable=SC2016
    log '$COMPOSER_AUTH is set, skipping Composer configuration'

    return 0
  fi

  log "Configuring Composer"

  if [[ -n "${GITHUB_USER}" ]] && [[ -n "${GITHUB_TOKEN}" ]]; then
    composer global config http-basic.github.com "${GITHUB_USER}" "${GITHUB_TOKEN}"
  fi

  if [[ -n "${BITBUCKET_PUBLIC_KEY}" ]] && [[ -n "${BITBUCKET_PRIVATE_KEY}" ]]; then
    composer global config bitbucket-oauth.bitbucket.org "${BITBUCKET_PUBLIC_KEY}" "${BITBUCKET_PRIVATE_KEY}"
  fi

  if [[ -n "${GITLAB_TOKEN}" ]]; then
    composer global config gitlab-token.gitlab.com "${GITLAB_TOKEN}"
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

dump_build_version() {
  log "Creating build version file"
  mkdir -p "$(app_path)"
  printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" >"$(app_path)/version.php"
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
  dump_build_version

  command_after_build

  run_hooks "post-build"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

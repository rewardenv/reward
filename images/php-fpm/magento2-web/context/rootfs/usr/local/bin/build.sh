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

readonly MAGENTO_COMMAND="${MAGENTO_COMMAND:-php -derror_reporting=E_ALL bin/magento --no-ansi --no-interaction}"
readonly MAGERUN_COMMAND="${MAGERUN_COMMAND:-php -derror_reporting=E_ALL $(command -v mr) --no-ansi --no-interaction}"
readonly COMPOSER_COMMAND="${COMPOSER_COMMAND:-php -derror_reporting=E_ALL $(command -v composer) --no-ansi --no-interaction}"
readonly N_COMMAND="${N_COMMAND:-$(command -v n)}"

magento() {
  ${MAGENTO_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

n() {
  ${N_COMMAND} "$@"
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

  if [[ -n "${MAGENTO_PUBLIC_KEY:-}" ]] && [[ -n "${MAGENTO_PRIVATE_KEY:-}" ]]; then
    composer global config http-basic.repo.magento.com "${MAGENTO_PUBLIC_KEY:-}" "${MAGENTO_PRIVATE_KEY:-}"
  fi

  if [[ -n "${GITHUB_USER:-}" ]] && [[ -n "${GITHUB_TOKEN:-}" ]]; then
    composer global config http-basic.github.com "${GITHUB_USER:-}" "${GITHUB_TOKEN:-}"
  fi

  if [[ -n "${BITBUCKET_PUBLIC_KEY:-}" ]] && [[ -n "${BITBUCKET_PRIVATE_KEY:-}" ]]; then
    composer global config bitbucket-oauth.bitbucket.org "${BITBUCKET_PUBLIC_KEY:-}" "${BITBUCKET_PRIVATE_KEY:-}"
  fi

  if [[ -n "${GITLAB_TOKEN:-}" ]]; then
    composer global config gitlab-token.gitlab.com "${GITLAB_TOKEN:-}"
  fi

  mkdir -p "$(app_path)/var/composer_home"

  local composer_home
  composer_home="$(composer config --global home)"
  if [[ -f "${composer_home}/auth.json" ]]; then
    cp -a "${composer_home}/auth.json" "$(app_path)/"
  fi

  if [[ -f "${composer_home}/composer.json" ]]; then
    cp -a "${composer_home}/composer.json" "$(app_path)/var/composer_home/"
  fi

  composer config --no-plugins allow-plugins.magento/* true || true
  composer config --no-plugins allow-plugins.laminas/laminas-dependency-plugin true || true
  composer config --no-plugins allow-plugins.dealerdirect/phpcodesniffer-composer-installer true || true
  composer config --no-plugins allow-plugins.cweagans/composer-patches true || true
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

magento_remove_env_file() {
  log "Removing env.php file"
  rm -f "$(app_path)/app/etc/env.php"
}

composer_dump_autoload() {
  log "Dumping Composer autoload"
  composer dump-autoload -o
}

magento_setup_di_compile() {
  if [[ "${MAGENTO_DI_COMPILE:-true}" != "true" ]] || [[ "${MAGENTO_DI_COMPILE_ON_DEMAND:-false}" == "true" ]]; then
    return 0
  fi

  log "Compiling Magento dependencies"
  magento setup:di:compile
}

magento_setup_static_content_deploy() {
  if [[ "${MAGENTO_SKIP_STATIC_CONTENT_DEPLOY:-false}" == "true" ]]; then
    return 0
  fi

  log "Deploying static content"
  magento setup:static-content:deploy -f
}

magento_create_pub_static_dir() {
  log "Creating static directory"
  mkdir -p "$(app_path)/pub/static"
}

dump_build_version() {
  log "Creating build version file"
  printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" >pub/version.php
}

main() {
  command_before_build

  n_install
  composer_self_update

  composer_configure
  composer_install
  composer_clear_cache
  magento_remove_env_file
  magento_setup_di_compile
  composer_dump_autoload
  # https://github.com/magento/magento2/issues/33802
  magento_remove_env_file
  magento_create_pub_static_dir
  dump_build_version

  command_after_build
}

main

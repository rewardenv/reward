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

_magento_command="bin/magento"

_magerun_command="n98-magerun2"
if command -v mr &>/dev/null; then
  _magerun_command="$(command -v mr 2>/dev/null)"
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
: "${MAGENTO_COMMAND:=php ${PHP_ARGS} ${_magento_command} --no-ansi --no-interaction}"
: "${MAGERUN_COMMAND:=php ${PHP_ARGS} ${_magerun_command} --no-ansi --no-interaction}"
: "${COMPOSER_COMMAND:=php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
: "${N_COMMAND:=${_n_command}}"

unset PHP_ARGS _magento_command _magerun_command _composer_command _n_command

: "${COMMAND_BEFORE_BUILD:=}"
: "${COMMAND_AFTER_BUILD:=}"
: "${NODE_VERSION:=}"
: "${COMPOSER_VERSION:=}"
: "${COMPOSER_AUTH:=}"
: "${MAGENTO_PUBLIC_KEY:=}"
: "${MAGENTO_PRIVATE_KEY:=}"
: "${GITHUB_USER:=}"
: "${GITHUB_TOKEN:=}"
: "${BITBUCKET_PUBLIC_KEY:=}"
: "${BITBUCKET_PRIVATE_KEY:=}"
: "${GITLAB_TOKEN:=}"

: "${MAGENTO_VERSION:=2.4.4}"
: "${MAGENTO_DI_COMPILE:=true}"
: "${MAGENTO_DI_COMPILE_ON_DEMAND:=false}"
: "${MAGENTO_SKIP_STATIC_CONTENT_DEPLOY:=false}"
: "${MAGENTO_SCD_ON_DEMAND:=false}"
: "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE:=true}"
: "${MAGENTO_THEMES:=}"
: "${MAGENTO_LANGUAGES:=}"
: "${COMMAND_BEFORE_MAGENTO_DI_COMPILE:=}"
: "${COMMAND_AFTER_MAGENTO_DI_COMPILE:=}"

: "${COMPOSER_INSTALL:=true}"
: "${COMPOSER_INSTALL_ARGS:=}"
: "${COMPOSER_DUMP_AUTOLOAD:=true}"
: "${COMMAND_BEFORE_COMPOSER_INSTALL:=}"
: "${COMMAND_AFTER_COMPOSER_INSTALL:=}"

magento() {
  ${MAGENTO_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

n() {
  ${N_COMMAND} "$@"
}

check_requirements() {
  check_command "mr"
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

  if [[ -n "${MAGENTO_PUBLIC_KEY}" ]] && [[ -n "${MAGENTO_PRIVATE_KEY}" ]]; then
    composer global config http-basic.repo.magento.com "${MAGENTO_PUBLIC_KEY}" "${MAGENTO_PRIVATE_KEY}"
  fi

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

composer_configure_home_for_magento() {
  mkdir -p "$(app_path)/var/composer_home"

  local composer_home
  composer_home="$(composer config --global home)"
  if [[ -n "${composer_home}" ]]; then
    if [[ -f "${composer_home}/auth.json" ]]; then
      cp -a "${composer_home}/auth.json" "$(app_path)/"
    fi

    if [[ -f "${composer_home}/composer.json" ]]; then
      cp -a "${composer_home}/composer.json" "$(app_path)/var/composer_home/"
    fi
  fi
}

composer_configure_plugins() {
  composer config --no-plugins allow-plugins.magento/* true || true
  composer config --no-plugins allow-plugins.laminas/laminas-dependency-plugin true || true
  composer config --no-plugins allow-plugins.dealerdirect/phpcodesniffer-composer-installer true || true
  composer config --no-plugins allow-plugins.cweagans/composer-patches true || true
}

command_before_composer_install() {
  if [[ -z "${COMMAND_BEFORE_COMPOSER_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command before composer_install"
  eval "${COMMAND_BEFORE_COMPOSER_INSTALL}"
}

command_after_composer_install() {
  if [[ -z "${COMMAND_AFTER_COMPOSER_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command after composer install"
  eval "${COMMAND_AFTER_MAGENTO_DI_COMPILE}"
}

composer_install() {
  if [[ ! -f "composer.json" ]]; then
    return 0
  fi

  command_before_composer_install

  log "Installing Composer dependencies"
  # shellcheck disable=SC2086
  composer install --no-progress ${COMPOSER_INSTALL_ARGS}

  command_after_composer_install
}

composer_clear_cache() {
  log "Clearing Composer cache"
  composer clear-cache
}

composer_dump_autoload() {
  if [[ "${COMPOSER_DUMP_AUTOLOAD}" != "true" ]]; then
    return 0
  fi

  log "Dumping Composer autoload"

  composer dump-autoload --optimize
}

magento_remove_env_file() {
  log "Removing env.php file"
  rm -f "$(app_path)/app/etc/env.php"
}

command_before_magento_di_compile() {
  if [[ -z "${COMMAND_BEFORE_MAGENTO_DI_COMPILE}" ]]; then
    return 0
  fi

  log "Executing custom command before magento setup:di:compile"
  eval "${COMMAND_BEFORE_MAGENTO_DI_COMPILE}"
}

command_after_magento_di_compile() {
  if [[ -z "${COMMAND_AFTER_MAGENTO_DI_COMPILE}" ]]; then
    return 0
  fi

  log "Executing custom command after magento setup:di:compile"
  eval "${COMMAND_AFTER_MAGENTO_DI_COMPILE}"
}

magento_setup_di_compile() {
  if [[ "${MAGENTO_DI_COMPILE}" != "true" ]] || [[ "${MAGENTO_DI_COMPILE_ON_DEMAND}" == "true" ]]; then
    return 0
  fi

  command_before_magento_di_compile

  log "Compiling Magento dependencies"
  magento setup:di:compile

  command_after_magento_di_compile
}

magento_setup_static_content_deploy() {
  if [[ "${MAGENTO_SKIP_STATIC_CONTENT_DEPLOY}" == "true" ]] || [[ "${MAGENTO_SCD_ON_DEMAND}" == "true" ]]; then
    return 0
  fi
  local ARGS=("--jobs=$(nproc)")

  local SCD_ARGS="-v"
  if version_gt "${MAGENTO_VERSION}" "2.3.99" && [[ "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE}" == "true" ]]; then
    SCD_ARGS="-fv"
  fi
  ARGS+=("${SCD_ARGS}")

  if [[ -n ${MAGENTO_THEMES} ]]; then
    read -r -a themes <<<"$MAGENTO_THEMES"
    local THEME_ARGS=$(printf -- '--theme=%s ' "${themes[@]}")
    # Remove trailing space
    THEME_ARGS=${THEME_ARGS% }
    ARGS+=("${THEME_ARGS}")
  fi

  if [[ -n "${MAGENTO_LANGUAGES}" ]]; then
    ARGS+=("${MAGENTO_LANGUAGES}")
  fi

  log "Deploying static content"
  magento setup:static-content:deploy "${ARGS[@]}"
}

magento_create_pub_static_dir() {
  log "Creating static directory"
  mkdir -p "$(app_path)/pub/static"
}

dump_build_version() {
  log "Creating build version file"
  mkdir -p "$(app_path)/pub"
  printf "<?php\nprintf(\"php-version: %%g </br>\", phpversion());\nprintf(\"build-date: $(date '+%Y/%m/%d %H:%M:%S')\");\n?>\n" >"$(app_path)/pub/version.php"
}

main() {
  run_hooks "pre-build"

  check_requirements

  command_before_build

  n_install
  composer_self_update

  composer_configure
  composer_configure_home_for_magento
  composer_configure_plugins
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

  run_hooks "post-build"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

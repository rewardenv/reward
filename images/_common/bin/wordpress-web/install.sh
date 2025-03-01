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

: "${PHP_ARGS:=-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} -dmemory_limit=${PHP_MEMORY_LIMIT:-2G}}"
: "${WORDPRESS_COMMAND:=php ${PHP_ARGS} ${_wordpress_command} --no-color}"
: "${COMPOSER_COMMAND:=php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"

unset PHP_ARGS _wordpress_command _composer_command

: "${COMPOSER_AUTH:=}"
: "${GITHUB_USER:=}"
: "${GITHUB_TOKEN:=}"
: "${BITBUCKET_PUBLIC_KEY:=}"
: "${BITBUCKET_PRIVATE_KEY:=}"
: "${GITLAB_TOKEN:=}"

: "${COMMAND_BEFORE_INSTALL:=}"
: "${COMMAND_AFTER_INSTALL:=}"
: "${SKIP_BOOTSTRAP:=false}"

: "${WORDPRESS_SKIP_BOOTSTRAP:=false}"
: "${WORDPRESS_SKIP_INSTALL:=false}"
: "${WORDPRESS_CONFIG:=true}"
: "${WORDPRESS_DATABASE_HOST:=db}"
: "${WORDPRESS_DATABASE_NAME:=wordpress}"
: "${WORDPRESS_DATABASE_USER:=wordpress}"
: "${WORDPRESS_DATABASE_PASSWORD:=wordpress}"
: "${WORDPRESS_DATABASE_PREFIX:=wp_}"
: "${WORDPRESS_DATABASE_CHARSET:=utf8}"
: "${WORDPRESS_DATABASE_COLLATE:=}"
: "${WORDPRESS_LOCALE:=}"
: "${WORDPRESS_EXTRA_PHP:=}"
: "${WORDPRESS_SCHEME:=https}"
: "${WORDPRESS_HOST:=wp.test}"
: "${WORDPRESS_URL:=${WORDPRESS_SCHEME}://${WORDPRESS_HOST}}"
: "${WORDPRESS_BLOG_NAME:=wordpress}"
: "${WORDPRESS_USER:=admin}"
: "${WORDPRESS_PASSWORD:=ASDqwe12345}"
: "${WORDPRESS_EMAIL:=admin@example.com}"
: "${WORDPRESS_DEPLOY_SAMPLE_DATA:=false}"
: "${WORDPRESS_FORCE_DEPLOY_SAMPLE_DATA:=false}"
: "${WORDPRESS_SHARED_FILES:=wp-config.php}"

wp() {
  ${WORDPRESS_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

check_requirements() {
  check_command "composer"
}

command_before_install() {
  if [[ -z "${COMMAND_BEFORE_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "${COMMAND_BEFORE_INSTALL}"
}

command_after_install() {
  if [[ -z "${COMMAND_AFTER_INSTALL}" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "${COMMAND_AFTER_INSTALL}"
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

bootstrap_check() {
  if [[ "${WORDPRESS_SKIP_BOOTSTRAP}" != "true" ]] && [[ "${SKIP_BOOTSTRAP}" != "true" ]]; then
    return 0
  fi

  log "Skipping Wordpress bootstrap"
  command_after_install
  exit
}

wordpress_is_installed() {
  if ! wp core is-installed; then
    false
  fi
}

wordpress_configure() {
  if [[ "${WORDPRESS_CONFIG}" != "true" ]]; then
    return 0
  fi
  log "Configuring Wordpress"

  declare -a ARGS
  ARGS+=(
    "--force"
    "--dbhost=${WORDPRESS_DATABASE_HOST}"
    "--dbname=${WORDPRESS_DATABASE_NAME}"
    "--dbuser=${WORDPRESS_DATABASE_USER}"
    "--dbpass=${WORDPRESS_DATABASE_PASSWORD}"
    "--dbprefix=${WORDPRESS_DATABASE_PREFIX}"
    "--dbcharset=${WORDPRESS_DATABASE_CHARSET}"
  )

  if [[ -n "${WORDPRESS_DATABASE_COLLATE}" ]]; then
    ARGS+=("--dbcollate=${WORDPRESS_DATABASE_COLLATE}")
  fi

  if [[ -n "${WORDPRESS_LOCALE}" ]]; then
    ARGS+=("--locale=${WORDPRESS_LOCALE}")
  fi

  if [[ -n "${WORDPRESS_EXTRA_PHP}" ]]; then
    wp core config "${ARGS[@]}" --extra-php <<PHP
${WORDPRESS_EXTRA_PHP}
PHP
    return $?
  fi

  wp core config "${ARGS[@]}"
}

wordpress_install() {
  if [[ "${WORDPRESS_SKIP_INSTALL}" == "true" ]]; then
    return 0
  fi

  log "Installing Wordpress"

  local ARGS=()
  ARGS+=(
    "--url=${WORDPRESS_URL}"
    "--title=${WORDPRESS_BLOG_NAME}"
    "--admin_user=${WORDPRESS_USER}"
    "--admin_password=${WORDPRESS_PASSWORD}"
    "--admin_email=${WORDPRESS_EMAIL}"
  )

  wp core install "${ARGS[@]}"
}

wordpress_deploy_sample_data() {
  if [[ "${WORDPRESS_DEPLOY_SAMPLE_DATA}" != "true" ]] && [[ "${WORDPRESS_FORCE_DEPLOY_SAMPLE_DATA}" != "true" ]]; then
    return 0
  fi

  log "Deploying sample data"
  wp plugin install --activate wordpress-importer

  curl -O https://raw.githubusercontent.com/manovotny/wptest/master/wptest.xml
  wp import wptest.xml --authors=create
  rm wptest.xml

  wp theme install twentytwentytwo --activate
}

wordpress_publish_shared_files() {
  log "Publishing config"

  local _shared_files="${WORDPRESS_SHARED_FILES}"

  publish_shared_files
}

main() {
  conditional_sleep

  LOCKFILE="$(shared_config_path)/.deploy.lock"
  readonly LOCKFILE

  trap 'lock_cleanup ${LOCKFILE}' EXIT
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  lock_acquire "${LOCKFILE}"

  run_hooks "pre-install"

  check_requirements

  command_before_install
  bootstrap_check
  composer_configure

  if wordpress_is_installed; then
    wordpress_configure
  else
    wordpress_configure
    wordpress_install
    wordpress_deploy_sample_data
  fi

  wordpress_publish_shared_files

  command_after_install

  run_hooks "post-install"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

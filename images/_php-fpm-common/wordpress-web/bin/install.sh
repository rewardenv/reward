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

PHP_ARGS="-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} -dmemory_limit=${PHP_MEMORY_LIMIT:-2G}"

_wordpress_command=wp
if command -v wp &>/dev/null; then
  _wordpress_command="$(command -v wp 2>/dev/null)"
fi
WORDPRESS_COMMAND="${WORDPRESS_COMMAND:-php ${PHP_ARGS} ${_wordpress_command} --no-color}"
readonly WORDPRESS_COMMAND
unset _wordpress_command

_composer_command="composer"
if command -v composer &>/dev/null; then
  _composer_command="$(command -v composer 2>/dev/null)"
fi
COMPOSER_COMMAND="${COMPOSER_COMMAND:-php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
readonly COMPOSER_COMMAND
unset _composer_command

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
  if [[ -z "${COMMAND_BEFORE_INSTALL:-}" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "${COMMAND_BEFORE_INSTALL:-}"
}

command_after_install() {
  if [[ -z "${COMMAND_AFTER_INSTALL:-}" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "${COMMAND_AFTER_INSTALL:-}"
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

bootstrap_check() {
  if [[ "${WORDPRESS_SKIP_BOOTSTRAP:-false}" != "true" ]] && [[ "${SKIP_BOOTSTRAP:-false}" != "true" ]]; then
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
  if [[ "${WORDPRESS_CONFIG:-true}" != "true" ]]; then
    return 0
  fi
  log "Configuring Wordpress"

  declare -a ARGS
  ARGS+=(
    "--force"
    "--dbhost=${WORDPRESS_DATABASE_HOST:-db}"
    "--dbname=${WORDPRESS_DATABASE_NAME:-wordpress}"
    "--dbuser=${WORDPRESS_DATABASE_USER:-wordpress}"
    "--dbpass=${WORDPRESS_DATABASE_PASSWORD:-wordpress}"
    "--dbprefix=${WORDPRESS_DATABASE_PREFIX:-wp_}"
    "--dbcharset=${WORDPRESS_DATABASE_CHARSET:-utf8}"
  )

  if [[ -n "${WORDPRESS_DATABASE_COLLATE:-}" ]]; then
    ARGS+=("--dbcollate=${WORDPRESS_DATABASE_COLLATE:-}")
  fi

  if [[ -n "${WORDPRESS_LOCALE:-}" ]]; then
    ARGS+=("--locale=${WORDPRESS_LOCALE:-}")
  fi

  if [[ -n "${WORDPRESS_EXTRA_PHP:-}" ]]; then
    wp core config "${ARGS[@]}" --extra-php <<PHP
${WORDPRESS_EXTRA_PHP:-}
PHP
    return $?
  fi

  wp core config "${ARGS[@]}"
}

wordpress_install() {
  if [[ "${WORDPRESS_SKIP_INSTALL:-false}" == "true" ]]; then
    return 0
  fi

  log "Installing Wordpress"

  local WORDPRESS_SCHEME="${WORDPRESS_SCHEME:-https}"
  local WORDPRESS_HOST="${WORDPRESS_HOST:-wp.test}"
  local WORDPRESS_URL="${WORDPRESS_URL:-"$WORDPRESS_SCHEME://$WORDPRESS_HOST"}"

  local ARGS=()
  ARGS+=(
    "--url=${WORDPRESS_URL}"
    "--title=${WORDPRESS_BLOG_NAME:-wordpress}"
    "--admin_user=${WORDPRESS_USER:-admin}"
    "--admin_password=${WORDPRESS_PASSWORD:-ASDqwe12345}"
    "--admin_email=${WORDPRESS_EMAIL:-admin@example.com}"
  )

  wp core install "${ARGS[@]}"
}

wordpress_deploy_sample_data() {
  if [[ "${WORDPRESS_DEPLOY_SAMPLE_DATA:-false}" != "true" ]]; then
    return 0
  fi

  log "Deploying sample data"
  wp plugin install --activate wordpress-importer

  curl -O https://raw.githubusercontent.com/manovotny/wptest/master/wptest.xml
  wp import wptest.xml --authors=create
  rm wptest.xml

  wp theme install twentytwentytwo --activate
}

wordpress_publish_config() {
  log "Publishing Wordpress configuration"
  cp -a "$(app_path)/wp-config.php" "$(shared_config_path)/wp-config.php"
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

  wordpress_publish_config

  command_after_install

  run_hooks "post-install"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

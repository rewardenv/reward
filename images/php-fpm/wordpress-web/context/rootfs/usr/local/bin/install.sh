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

readonly WORDPRESS_COMMAND="${WORDPRESS_COMMAND:-$(command -v wp) --no-color}"
readonly COMPOSER_COMMAND="${COMPOSER_COMMAND:-php -derror_reporting=E_ALL $(command -v composer) --no-ansi --no-interaction}"

wp() {
  ${WORDPRESS_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
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
  if [[ "${WORDPRESS_SKIP_BOOTSTRAP:-false}" != "true" ]] || [[ "${SKIP_BOOTSTRAP:-false}" == "true" ]]; then
    return 0
  fi

  log "Skipping Wordpress bootstrap"
  command_after_install
  exit
}

wordpress_is_installed() {
  if ! ${WORDPRESS_COMMAND} core is-installed; then
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
    "--dbcollate=${WORDPRESS_DATABASE_COLLATE:-}"
    "--locale=${WORDPRESS_LOCALE:-}"
  )

  wp core config "${ARGS[@]}" --extra-php <<PHP
${WORDPRESS_EXTRA_PHP:-}
PHP
}

wordpress_install() {
  if [[ "${WORDPRESS_SKIP_INSTALL:-false}" == "true" ]]; then
    return 0
  fi

  log "Installing Wordpress"

  WORDPRESS_SCHEME="${WORDPRESS_SCHEME:-https}"
  WORDPRESS_HOST="${WORDPRESS_HOST:-wp.test}"
  WORDPRESS_URL="${WORDPRESS_URL:-"$WORDPRESS_SCHEME://$WORDPRESS_HOST"}"

  declare -a ARGS
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
  LOCKFILE="$(shared_config_path)/.deploy.lock"
  readonly LOCKFILE

  trap 'lock_cleanup ${LOCKFILE}' EXIT
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  lock_acquire "${LOCKFILE}"

  conditional_sleep
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
}

main

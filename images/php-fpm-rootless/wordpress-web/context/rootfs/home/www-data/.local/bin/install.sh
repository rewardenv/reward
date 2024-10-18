#!/usr/bin/env bash
[ "${DEBUG:-false}" = "true" ] && set -x
set -eE -o pipefail -o errtrace
shopt -s extdebug

log() {
  [ "${SILENT:-false}" != "true" ] && printf "%s INFO: %s\n" "$(date --iso-8601=seconds)" "$*"
}

error() {
  exitcode=$?
  # color red
  printf "\033[1;31m%s ERROR: %s\033[0m\n" "$(date --iso-8601=seconds)" "$*"

  if [ "${exitcode}" -eq 0 ]; then exit 1; fi
  exit "${exitcode}"
}

trapinfo() {
  # shellcheck disable=SC2145
  error "Command failed: $BASH_COMMAND STATUS=$? LINENO=${@:1:$(($# - 1))}"
}

version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

command_before_install() {
  if [ -n "${COMMAND_BEFORE_INSTALL-}" ]; then
    log "Executing custom command before installation"
    eval "${COMMAND_BEFORE_INSTALL-}"
  fi
}

command_after_install() {
  if [ -n "${COMMAND_AFTER_INSTALL-}" ]; then
    log "Executing custom command after installation"
    eval "${COMMAND_AFTER_INSTALL-}"
  fi
}

WORDPRESS_COMMAND="${WORDPRESS_COMMAND:-php -derror_reporting=E_ALL $(command -v wp)} --no-color"

wordpress_bootstrap_check() {
  if [ "${WORDPRESS_SKIP_BOOTSTRAP:-false}" = "true" ]; then
    log "Skipping Wordpress bootstrap"
    command_after_install
    exit
  fi
}

wordpress_is_installed() {
  if ! ${WORDPRESS_COMMAND} core is-installed; then
    false
  fi
}

wordpress_configure() {
  if [ "${WORDPRESS_CONFIG:-true}" = "true" ]; then
    log "Configuring WordPress"

    ARGS=()
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

    ${WORDPRESS_COMMAND} core config "${ARGS[@]}" --extra-php <<PHP
${WORDPRESS_EXTRA_PHP}
PHP
  fi
}

wordpress_install() {
  if [ "${WORDPRESS_SKIP_INSTALL:-false}" != "true" ]; then
    log "Installing WordPress"

    WORDPRESS_SCHEME="${WORDPRESS_SCHEME:-https}"
    WORDPRESS_HOST="${WORDPRESS_HOST:-wp.test}"
    WORDPRESS_URL="${WORDPRESS_URL:-"$WORDPRESS_SCHEME://$WORDPRESS_HOST"}"

    ARGS=()
    ARGS+=(
      "--url=${WORDPRESS_URL}"
      "--title=${WORDPRESS_BLOG_NAME:-wordpress}"
      "--admin_user=${WORDPRESS_USER:-admin}"
      "--admin_password=${WORDPRESS_PASSWORD:-ASDqwe12345}"
      "--admin_email=${WORDPRESS_EMAIL:-admin@example.com}"
    )

    ${WORDPRESS_COMMAND} core install "${ARGS[@]}"
  fi
}

wordpress_deploy_sample_data() {
  if [ "${WORDPRESS_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
    log "deploying sample data"
    ${WORDPRESS_COMMAND} plugin install --activate wordpress-importer

    curl -O https://raw.githubusercontent.com/manovotny/wptest/master/wptest.xml
    ${WORDPRESS_COMMAND} import wptest.xml --authors=create
    rm wptest.xml

    ${WORDPRESS_COMMAND} theme install twentytwentytwo --activate
  fi
}

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  command_before_install

  wordpress_bootstrap_check

  if wordpress_is_installed; then
    wordpress_configure
  else
    wordpress_configure
    wordpress_install
    wordpress_deploy_sample_data
  fi

  command_after_install
}

main

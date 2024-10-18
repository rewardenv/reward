#!/usr/bin/env bash
[ "${DEBUG:-false}" = "true" ] && set -x
set -eE -o pipefail -o errtrace
shopt -s extdebug

if [ "${DEBUG:-false}" != "true" ]; then
  QUIET="--quiet"
fi

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

check_database() {
  log "Checking database connection..."
  if mysql -h"${WORDPRESS_DATABASE_HOST:-db}" -P"${WORDPRESS_DATABASE_PORT:-3306}" -u"${WORDPRESS_DATABASE_NAME:-wordpress}" -p"${WORDPRESS_DATABASE_PASSWORD:-wordpress}" -e "CREATE DATABASE IF NOT EXISTS ${WORDPRESS_DATABASE_NAME:-wordpress}; "; then
    log "Database connection ready."
  else
    error "Database connection failed."
  fi
}

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  check_database
  log "All connections are ready."
}

main

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

MAGENTO_COMMAND="${MAGENTO_COMMAND:-php -derror_reporting=E_ALL bin/magento} --no-ansi --no-interaction"
MAGERUN_COMMAND="${MAGERUN_COMMAND:-php -derror_reporting=E_ALL $(command -v mr)} --no-ansi --no-interaction"

magento_bootstrap_check() {
  if [ "${MAGENTO_SKIP_BOOTSTRAP:-false}" = "true" ]; then
    log "Skipping Magento bootstrap"
    command_after_install
    exit
  fi
}

magento_is_installed() {
  if [ -f app/etc/env.php ]; then
    return
  fi

  false
}

magento_args_install_only() {
  ARGS+=(
    "--base-url=${MAGENTO_BASE_URL}"
    "--base-url-secure=${MAGENTO_BASE_URL_SECURE}"
  )
  if [ "${MAGENTO_ENABLE_HTTPS:-true}" = "true" ]; then
    ARGS+=(
      "--use-secure=1"
    )
  fi
  if [ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" = "true" ]; then
    ARGS+=(
      "--use-secure-admin=1"
    )
  fi
  if [ "${MAGENTO_USE_REWRITES:-true}" = "true" ]; then
    ARGS+=(
      "--use-rewrites=1"
    )
  fi
}

magento_args_defaults() {
  ARGS+=(
    "--key=${MAGENTO_KEY:-12345678901234567890123456789012}"
    "--backend-frontname=${MAGENTO_ADMIN_URL_PREFIX:-admin}"
  )
}

magento_args_db() {
  ARGS+=(
    "--db-host=${MAGENTO_DATABASE_HOST:-db}"
    "--db-name=${MAGENTO_DATABASE_NAME:-magento}"
    "--db-user=${MAGENTO_DATABASE_USER:-magento}"
    "--db-password=${MAGENTO_DATABASE_PASSWORD:-magento}"
  )
}

magento_args_redis() {
  # Configure Redis
  if [ "${MAGENTO_REDIS_ENABLED:-true}" = "true" ]; then
    MAGENTO_REDIS_HOST=${MAGENTO_REDIS_HOST:-redis}
    MAGENTO_REDIS_PORT=${MAGENTO_REDIS_PORT:-6379}

    ARGS+=(
      "--session-save=redis"
      "--session-save-redis-host=${MAGENTO_SESSION_SAVE_REDIS_HOST:-$MAGENTO_REDIS_HOST}"
      "--session-save-redis-port=${MAGENTO_SESSION_SAVE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
      "--session-save-redis-db=${MAGENTO_SESSION_SAVE_REDIS_SESSION_DB:-2}"
      "--session-save-redis-max-concurrency=${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY:-20}"
    )

    if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_SESSION_REDIS_PASSWORD} ]]; then
      ARGS+=(
        "--session-save-redis-password=${MAGENTO_SESSION_SAVE_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
      )
    fi

    if [ "${MAGENTO_CACHE_BACKEND:-redis}" = "redis" ]; then
      ARGS+=(
        "--cache-backend=redis"
        "--cache-backend-redis-server=${MAGENTO_CACHE_BACKEND_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
        "--cache-backend-redis-port=${MAGENTO_CACHE_BACKEND_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
        "--cache-backend-redis-db=${MAGENTO_CACHE_BACKEND_REDIS_DB:-0}"
      )

      if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD} ]]; then
        ARGS+=(
          "--cache-backend-redis-password=${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
        )
      fi
    fi

    if [ "${MAGENTO_PAGE_CACHE:-redis}" = "redis" ]; then
      ARGS+=(
        "--page-cache=redis"
        "--page-cache-redis-server=${MAGENTO_PAGE_CACHE_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
        "--page-cache-redis-port=${MAGENTO_PAGE_CACHE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
        "--page-cache-redis-db=${MAGENTO_PAGE_CACHEC_REDIS_DB:-1}"
      )
      if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_PAGE_CACHE_REDIS_PASSWORD} ]]; then
        ARGS+=(
          "--page-cache-redis-password=${MAGENTO_PAGE_CACHE_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
        )
      fi
    fi
  else
    ARGS+=(
      "--session-save=files"
    )
  fi
}

magento_args_varnish() {
  # Configure Varnish
  if [ "${MAGENTO_VARNISH_ENABLED:-true}" = "true" ]; then
    ARGS+=(
      "--http-cache-hosts=${MAGENTO_VARNISH_HOST:-varnish}:${MAGENTO_VARNISH_PORT:-80}"
    )
  fi
}

magento_args_rabbitmq() {
  # Configure RabbitMQ
  if [ "${MAGENTO_RABBITMQ_ENABLED:-true}" = "true" ]; then
    ARGS+=(
      "--amqp-host=${MAGENTO_AMQP_HOST:-rabbitmq}"
      "--amqp-port=${MAGENTO_AMQP_PORT:-5672}"
      "--amqp-user=${MAGENTO_AMQP_USER:-guest}"
      "--amqp-password=${MAGENTO_AMQP_PASSWORD:-guest}"
      "--amqp-virtualhost=${MAGENTO_AMQP_VIRTUAL_HOST:-/}"
    )

    if version_gt "${MAGENTO_VERSION}" "2.3.99"; then
      ARGS+=(
        "--consumers-wait-for-messages=0"
      )
    fi
  fi
}

magento_args_search() {
  if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" = "true" ]; then
    if version_gt "${MAGENTO_VERSION}" "2.3.99"; then
      if [ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" = "elasticsearch7" ]; then
        ARGS+=(
          "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
          "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST:-elasticsearch}"
          "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT:-9200}"
          "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
          "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
          "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
        )
      elif [ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" = "opensearch" ]; then
        ARGS+=(
          "--search-engine=${MAGENTO_SEARCH_ENGINE:-opensearch}"
          "--opensearch-host=${MAGENTO_OPENSEARCH_HOST:-opensearch}"
          "--opensearch-port=${MAGENTO_OPENSEARCH_PORT:-9200}"
          "--opensearch-index-prefix=${MAGENTO_OPENSEARCH_INDEX_PREFIX:-magento2}"
          "--opensearch-enable-auth=${MAGENTO_OPENSEARCH_ENABLE_AUTH:-0}"
          "--opensearch-timeout=${MAGENTO_OPENSEARCH_TIMEOUT:-15}"
        )
      fi
    fi
  fi
}

magento_args_sample_data() {
  if [ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
    ARGS+=(
      "--use-sample-data"
    )
  fi
}

magento_args_extra() {
  if [ -n "${MAGENTO_EXTRA_INSTALL_ARGS:-}" ]; then
    ARGS+=(
      "${MAGENTO_EXTRA_INSTALL_ARGS}"
    )
  fi
}

magento_setup_install() {
  if [ "${MAGENTO_SKIP_INSTALL:-false}" != "true" ]; then
    log "Installing Magento ${MAGENTO_VERSION:-'2.4.4'}"

    MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.4'}
    MAGENTO_HOST=${MAGENTO_HOST:-'magento.test'}
    MAGENTO_BASE_URL=${MAGENTO_BASE_URL:-"http://$MAGENTO_HOST"}
    MAGENTO_BASE_URL_SECURE=${MAGENTO_BASE_URL_SECURE:-"https://$MAGENTO_HOST"}

    ARGS=()

    magento_args_install_only
    magento_args_defaults
    magento_args_db
    magento_args_redis
    magento_args_varnish
    magento_args_rabbitmq
    magento_args_search
    magento_args_sample_data
    magento_args_extra

    # shellcheck disable=SC2068
    ${MAGENTO_COMMAND} setup:install ${ARGS[@]}

    magento_configure_search
  fi
}

magento_configure() {
  log "Configuring Magento"
  INSTALL_DATE=$(${MAGERUN_COMMAND} config:env:show install.date 2>/dev/null || true)
  CACHE_GRAPHQL_ID_SALT=$(${MAGERUN_COMMAND} config:env:show cache.graphql.id_salt 2>/dev/null || true)

  rm -f app/etc/env.php

  MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.4'}

  ARGS=()
  magento_args_defaults
  magento_args_db
  magento_args_redis
  magento_args_varnish
  magento_args_rabbitmq

  # shellcheck disable=SC2068
  ${MAGENTO_COMMAND} setup:config:set ${ARGS[@]}

  if [ -z "${INSTALL_DATE:+x}" ]; then INSTALL_DATE=$(date -R); fi
  ${MAGERUN_COMMAND} config:env:set install.date "${INSTALL_DATE}"

  if [ -n "${CACHE_GRAPHQL_ID_SALT}" ]; then ${MAGERUN_COMMAND} config:env:set cache.graphql.id_salt "${CACHE_GRAPHQL_ID_SALT}"; fi

  magento_configure_search
}

magento_configure_search() {
  if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" = "true" ]; then
    if version_gt "${MAGENTO_VERSION}" "2.3.99"; then
      if [ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" = "elasticsearch7" ]; then
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-opensearch}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/elasticsearch7_server_hostname" "${MAGENTO_ELASTICSEARCH_HOST:-opensearch}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/elasticsearch7_server_port" "${MAGENTO_ELASTICSEARCH_PORT:-9200}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/elasticsearch7_index_prefix" "${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/elasticsearch7_enable_auth" "${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/elasticsearch7_server_timeout" "${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
      elif [ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" = "opensearch" ]; then
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-opensearch}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/opensearch_server_hostname" "${MAGENTO_OPENSEARCH_HOST:-opensearch}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/opensearch_server_port" "${MAGENTO_OPENSEARCH_PORT:-9200}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/opensearch_index_prefix" "${MAGENTO_OPENSEARCH_INDEX_PREFIX:-magento2}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/opensearch_enable_auth" "${MAGENTO_OPENSEARCH_ENABLE_AUTH:-0}"
        ${MAGENTO_COMMAND} config:set --lock-env "catalog/search/opensearch_server_timeout" "${MAGENTO_OPENSEARCH_TIMEOUT:-15}"
      fi
    fi
  fi
}

magento_setup_di_compile() {
  if [ "${MAGENTO_DI_COMPILE:-false}" = "true" ] || [ "${MAGENTO_DI_COMPILE_ON_DEMAND:-false}" = "true" ]; then
    log "Compiling Magento dependencies"
    ${MAGENTO_COMMAND} setup:di:compile
  fi
}

magento_setup_static_content_deploy() {
  if [ "${MAGENTO_STATIC_CONTENT_DEPLOY:-false}" = "true" ] || [ "${MAGENTO_SCD_ON_DEMAND:-false}" = "true" ]; then
    log "Deploying static content"
    if [ "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE}" = "true" ]; then
      ${MAGENTO_COMMAND} setup:static-content:deploy --jobs=$(nproc) -fv ${MAGENTO_LANGUAGES:-}
    else
      ${MAGENTO_COMMAND} setup:static-content:deploy --jobs=$(nproc) -v ${MAGENTO_LANGUAGES:-}
    fi
  fi
}

magento_cache_enable() {
  log "Enabling cache"
  ${MAGENTO_COMMAND} cache:enable
}

magento_reindex() {
  if [ "${MAGENTO_SKIP_REINDEX:-true}" != "true" ]; then
    ${MAGENTO_COMMAND} indexer:reindex
  fi
}

magento_upgrade_required() {
  log "Checking if Magento upgrade is required"
  if ${MAGENTO_COMMAND} setup:db:status; then
    false
  fi
}

magento_maintenance_disable() {
  log "Disabling maintenance mode"
  ${MAGENTO_COMMAND} maintenance:disable
}

magento_maintenance_enable() {
  log "Enabling maintenance mode"
  ${MAGENTO_COMMAND} maintenance:enable
}

magento_setup_upgrade() {
  if [ "${MAGENTO_SKIP_UPGRADE:-false}" != "true" ]; then
    if magento_upgrade_required; then
      log "Running Magento setup:upgrade"
      magento_maintenance_enable
      ${MAGENTO_COMMAND} setup:upgrade --keep-generated
      magento_maintenance_disable
    fi

    log "Magento upgrade is not required"
  fi
}

magento_deploy_mode_set() {
  if [ "${MAGENTO_MODE:-default}" != "default" ]; then
    log "Setting Magento deploy mode to ${MAGENTO_MODE:-default}"
    ${MAGENTO_COMMAND} deploy:mode:set "${MAGENTO_MODE:-default}"
  fi
}

magento_secure_frontend() {
  if [ "${MAGENTO_ENABLE_HTTPS:-true}" = "true" ]; then
    log "Enabling HTTPS for frontend"
    ${MAGENTO_COMMAND} config:set "web/secure/use_in_frontend" 1
  fi
}

magento_secure_backend() {
  if [ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" = "true" ]; then
    log "Enabling HTTPS for admin"
    ${MAGENTO_COMMAND} config:set "web/secure/use_in_adminhtml" 1
  fi
}

magento_use_rewrites() {
  if [ "${MAGENTO_USE_REWRITES:-true}" = "true" ]; then
    log "Enabling rewrites"
    ${MAGENTO_COMMAND} config:set "web/seo/use_rewrites" 1
  fi
}

magento_admin_user() {
  if ${MAGERUN_COMMAND} admin:user:list --format=csv | tail -n +2 | awk -F',' '{print $2}' | grep "^${MAGENTO_USERNAME:-admin}$" >/dev/null; then
    log "Admin user already exists, updating password"
    ${MAGERUN_COMMAND} admin:user:change-password ${MAGENTO_USERNAME:-admin} ${MAGENTO_PASSWORD:-ASDqwe123}
  else
    log "Creating admin user"
    ARGS=()
    ARGS+=(
      "--admin-firstname=${MAGENTO_FIRST_NAME:-admin}"
      "--admin-lastname=${MAGENTO_LAST_NAME:-admin}"
      "--admin-email=${MAGENTO_EMAIL:-admin@example.com}"
      "--admin-user=${MAGENTO_USERNAME:-admin}"
      "--admin-password=${MAGENTO_PASSWORD:-ASDqwe123}"
    )
    ${MAGERUN_COMMAND} admin:user:delete --force "${MAGENTO_USERNAME:-admin}" || true
    ${MAGERUN_COMMAND} admin:user:delete --force "${MAGENTO_EMAIL:-admin@example.com}" || true
    ${MAGERUN_COMMAND} admin:user:create "${ARGS[@]}"
  fi
}

magento_deploy_sample_data() {
  if [ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
    log "Deploying sample data"
    ${MAGENTO_COMMAND} sampledata:deploy
    ${MAGENTO_COMMAND} setup:upgrade --keep-generated
  fi
}

main() {
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  command_before_install

  magento_bootstrap_check

  if magento_is_installed; then
    magento_configure
    magento_setup_upgrade
  else
    magento_setup_install
  fi

  magento_setup_di_compile
  magento_setup_static_content_deploy
  magento_deploy_mode_set
  magento_secure_frontend
  magento_secure_backend
  magento_use_rewrites
  magento_admin_user
  magento_deploy_sample_data
  magento_setup_static_content_deploy
  magento_cache_enable
  magento_reindex

  command_after_install
}

main

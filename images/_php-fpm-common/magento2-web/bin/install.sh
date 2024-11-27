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

_magento_command="bin/magento"
MAGENTO_COMMAND="${MAGENTO_COMMAND:-php ${PHP_ARGS} ${_magento_command} --no-ansi --no-interaction}"
readonly MAGENTO_COMMAND
unset _magento_command

_magerun_command="n98-magerun2"
if command -v mr 2>/dev/null; then
  _magerun_command="$(command -v mr 2>/dev/null)"
fi
MAGERUN_COMMAND="${MAGERUN_COMMAND:-php ${PHP_ARGS} ${_magerun_command} --no-ansi --no-interaction}"
readonly MAGERUN_COMMAND
unset _magerun_command

_composer_command="composer"
if command -v composer 2>/dev/null; then
  _composer_command="$(command -v composer 2>/dev/null)"
fi
COMPOSER_COMMAND="${COMPOSER_COMMAND:-php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
readonly COMPOSER_COMMAND
unset _composer_command

magento() {
  ${MAGENTO_COMMAND} "$@"
}

magerun() {
  ${MAGERUN_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

check_requirements() {
  check_command "composer"
  check_command "mr"
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
  if [[ -n "${composer_home:-}" ]]; then
    if [[ -f "${composer_home}/auth.json" ]]; then
      cp -a "${composer_home}/auth.json" "$(app_path)/"
    fi

    if [[ -f "${composer_home}/composer.json" ]]; then
      cp -a "${composer_home}/composer.json" "$(app_path)/var/composer_home/"
    fi
  fi
}

bootstrap_check() {
  if [[ "${MAGENTO_SKIP_BOOTSTRAP:-false}" != "true" ]] && [[ "${SKIP_BOOTSTRAP:-false}" != "true" ]]; then
    return 0
  fi

  log "Skipping Magento bootstrap"
  command_after_install
  exit
}

magento_is_installed() {
  if [[ ! -f "$(app_path)/app/etc/env.php" ]] && [[ ! -f "$(shared_config_path)/app/etc/env.php" ]]; then
    false
  fi
}

magento_args_install_only() {
  local MAGENTO_HOST=${MAGENTO_HOST:-'magento.test'}
  local MAGENTO_BASE_URL=${MAGENTO_BASE_URL:-"http://${MAGENTO_HOST}"}
  local MAGENTO_BASE_URL_SECURE=${MAGENTO_BASE_URL_SECURE:-"https://${MAGENTO_HOST}"}

  ARGS+=(
    "--base-url=${MAGENTO_BASE_URL}"
    "--base-url-secure=${MAGENTO_BASE_URL_SECURE}"
  )
  if [[ "${MAGENTO_ENABLE_HTTPS:-true}" == "true" ]]; then
    ARGS+=(
      "--use-secure=1"
    )
  fi
  if [[ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" == "true" ]]; then
    ARGS+=(
      "--use-secure-admin=1"
    )
  fi
  if [[ "${MAGENTO_USE_REWRITES:-true}" == "true" ]]; then
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
  if [[ "${MAGENTO_REDIS_ENABLED:-true}" != "true" ]]; then
    ARGS+=(
      "--session-save=files"
    )
    return 0
  fi

  MAGENTO_REDIS_HOST=${MAGENTO_REDIS_HOST:-redis}
  MAGENTO_REDIS_PORT=${MAGENTO_REDIS_PORT:-6379}

  if [[ "${MAGENTO_SESSION_SAVE:-redis}" == "redis" ]]; then
    ARGS+=(
      "--session-save=redis"
      "--session-save-redis-host=${MAGENTO_SESSION_SAVE_REDIS_HOST:-$MAGENTO_REDIS_HOST}"
      "--session-save-redis-port=${MAGENTO_SESSION_SAVE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
      "--session-save-redis-db=${MAGENTO_SESSION_SAVE_REDIS_DB:-2}"
      "--session-save-redis-max-concurrency=${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY:-20}"
    )

    if [[ -n "${MAGENTO_REDIS_PASSWORD:-}" ]] || [[ -n "${MAGENTO_SESSION_SAVE_REDIS_PASSWORD:-}" ]]; then
      ARGS+=(
        "--session-save-redis-password=${MAGENTO_SESSION_SAVE_REDIS_PASSWORD:-${MAGENTO_REDIS_PASSWORD:-}}"
      )
    fi
  fi

  if [[ "${MAGENTO_CACHE_BACKEND:-redis}" == "redis" ]]; then
    ARGS+=(
      "--cache-backend=redis"
      "--cache-backend-redis-server=${MAGENTO_CACHE_BACKEND_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
      "--cache-backend-redis-port=${MAGENTO_CACHE_BACKEND_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
      "--cache-backend-redis-db=${MAGENTO_CACHE_BACKEND_REDIS_DB:-0}"
    )

    if [[ -n "${MAGENTO_REDIS_PASSWORD:-}" ]] || [[ -n "${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD:-}" ]]; then
      ARGS+=(
        "--cache-backend-redis-password=${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD:-${MAGENTO_REDIS_PASSWORD:-}}"
      )
    fi
  fi

  if [[ "${MAGENTO_PAGE_CACHE:-redis}" == "redis" ]]; then
    ARGS+=(
      "--page-cache=redis"
      "--page-cache-redis-server=${MAGENTO_PAGE_CACHE_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
      "--page-cache-redis-port=${MAGENTO_PAGE_CACHE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
      "--page-cache-redis-db=${MAGENTO_PAGE_CACHE_REDIS_DB:-1}"
    )
    if [[ -n "${MAGENTO_REDIS_PASSWORD:-}" ]] || [[ -n "${MAGENTO_PAGE_CACHE_REDIS_PASSWORD:-}" ]]; then
      ARGS+=(
        "--page-cache-redis-password=${MAGENTO_PAGE_CACHE_REDIS_PASSWORD:-${MAGENTO_REDIS_PASSWORD:-}}"
      )
    fi
  fi
}

magento_args_varnish() {
  # Configure Varnish
  if [[ "${MAGENTO_VARNISH_ENABLED:-true}" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--http-cache-hosts=${MAGENTO_VARNISH_HOST:-varnish}:${MAGENTO_VARNISH_PORT:-80}"
  )
}

magento_args_rabbitmq() {
  # Configure RabbitMQ
  if [[ "${MAGENTO_RABBITMQ_ENABLED:-true}" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--amqp-host=${MAGENTO_AMQP_HOST:-rabbitmq}"
    "--amqp-port=${MAGENTO_AMQP_PORT:-5672}"
    "--amqp-user=${MAGENTO_AMQP_USER:-guest}"
    "--amqp-password=${MAGENTO_AMQP_PASSWORD:-guest}"
    "--amqp-virtualhost=${MAGENTO_AMQP_VIRTUAL_HOST:-/}"
  )

  if version_gt "${MAGENTO_VERSION:-'2.4.4'}" "2.3.99"; then
    ARGS+=(
      "--consumers-wait-for-messages=${MAGENTO_CONSUMERS_WAIT_FOR_MESSAGES:-0}"
    )
  fi
}

magento_args_search() {
  if [[ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" != "true" ]] && [[ "${MAGENTO_OPENSEARCH_ENABLED:-false}" != "true" ]]; then
    return 0
  fi

  if [[ "${MAGENTO_OPENSEARCH_ENABLED:-false}" == "true" ]] || [[ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" == "opensearch" ]]; then
    ARGS+=(
      "--search-engine=${MAGENTO_SEARCH_ENGINE:-opensearch}"
      "--opensearch-host=${MAGENTO_OPENSEARCH_HOST:-opensearch}"
      "--opensearch-port=${MAGENTO_OPENSEARCH_PORT:-9200}"
      "--opensearch-index-prefix=${MAGENTO_OPENSEARCH_INDEX_PREFIX:-magento2}"
      "--opensearch-enable-auth=${MAGENTO_OPENSEARCH_ENABLE_AUTH:-0}"
      "--opensearch-timeout=${MAGENTO_OPENSEARCH_TIMEOUT:-15}"
    )

    search_configured

    return 0
  fi

  # Elasticsearch 7 is required for Magento 2.4.0+ or later (if not using OpenSearch)
  if version_gt "${MAGENTO_VERSION:-'2.4.4'}" "2.3.99" || [[ "${MAGENTO_ELASTICSEARCH_ENABLED:-}" == "true" ]]; then
    ARGS+=(
      "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
      "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST:-elasticsearch}"
      "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT:-9200}"
      "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
      "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
      "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
    )
  fi

  search_configured
}

search_configured() {
  export SEARCH_CONFIGURED=true
}

magento_args_sample_data() {
  if [[ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--use-sample-data"
  )
}

magento_args_extra() {
  if [[ -z "${MAGENTO_EXTRA_INSTALL_ARGS:-}" ]]; then
    return 0
  fi

  ARGS+=(
    "${MAGENTO_EXTRA_INSTALL_ARGS:-}"
  )
}

magento_setup_install() {
  if [[ "${MAGENTO_SKIP_INSTALL:-false}" == "true" ]]; then
    return 0
  fi

  log "Installing Magento ${MAGENTO_VERSION:-'2.4.4'}"

  local MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.4'}

  local ARGS=("")

  magento_args_install_only
  magento_args_defaults
  magento_args_db
  magento_args_search
  magento_args_redis
  magento_args_varnish
  magento_args_rabbitmq
  magento_args_sample_data
  magento_args_extra

  # shellcheck disable=SC2068
  magento setup:install ${ARGS[@]}

  magento_configure_search
}

magento_configure() {
  log "Configuring Magento"

  install_date=$(magerun config:env:show install.date 2>/dev/null || true)
  cache_graphql_id_salt=$(magerun config:env:show cache.graphql.id_salt 2>/dev/null || true)

  rm -f "$(app_path)/app/etc/env.php"

  local MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.4'}
  local ARGS=("")

  magento_args_defaults
  magento_args_db
  if magento_search_configurable; then
    magento_args_search
  fi
  magento_args_redis
  magento_args_varnish
  magento_args_rabbitmq

  # shellcheck disable=SC2068
  magento setup:config:set ${ARGS[@]}

  if [[ -z "${install_date:+x}" ]]; then install_date=$(date -R); fi
  magerun config:env:set install.date "${install_date}"

  if [[ -n "${cache_graphql_id_salt:-}" ]]; then magerun config:env:set cache.graphql.id_salt "${cache_graphql_id_salt:-}"; fi

  magento_configure_search
}

magento_search_configurable() {
  if ! magento setup:config:set --help | grep -q '\-\-search-engine'; then
    false
  fi
}

magento_configure_search() {
  if [[ "${SEARCH_CONFIGURED:-false}" == "true" ]]; then
    return 0
  fi

  if [[ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" != "true" ]] && [[ "${MAGENTO_OPENSEARCH_ENABLED:-false}" != "true" ]]; then
    return 0
  fi

  if [[ "${MAGENTO_OPENSEARCH_ENABLED:-false}" == "true" ]] || [[ "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}" == "opensearch" ]]; then
    magento config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-opensearch}"
    magento config:set --lock-env "catalog/search/opensearch_server_hostname" "${MAGENTO_OPENSEARCH_HOST:-opensearch}"
    magento config:set --lock-env "catalog/search/opensearch_server_port" "${MAGENTO_OPENSEARCH_PORT:-9200}"
    magento config:set --lock-env "catalog/search/opensearch_index_prefix" "${MAGENTO_OPENSEARCH_INDEX_PREFIX:-magento2}"
    magento config:set --lock-env "catalog/search/opensearch_enable_auth" "${MAGENTO_OPENSEARCH_ENABLE_AUTH:-0}"
    magento config:set --lock-env "catalog/search/opensearch_server_timeout" "${MAGENTO_OPENSEARCH_TIMEOUT:-15}"

    return 0
  fi

  # Elasticsearch 7 is required for Magento 2.4.0+ or later (if not using OpenSearch)
  if version_gt "${MAGENTO_VERSION:-'2.4.4'}" "2.3.99"; then
    magento config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_hostname" "${MAGENTO_ELASTICSEARCH_HOST:-opensearch}"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_port" "${MAGENTO_ELASTICSEARCH_PORT:-9200}"
    magento config:set --lock-env "catalog/search/elasticsearch7_index_prefix" "${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
    magento config:set --lock-env "catalog/search/elasticsearch7_enable_auth" "${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_timeout" "${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
  fi
}

magento_setup_di_compile() {
  # Skip DI compile if it's not enabled explicitly (as it should be part of the build by default)
  if [[ "${MAGENTO_DI_COMPILE:-false}" != "true" ]] && [[ "${MAGENTO_DI_COMPILE_ON_DEMAND:-false}" != "true" ]]; then
    return 0
  fi

  log "Compiling Magento dependencies"
  magento setup:di:compile
}

magento_setup_static_content_deploy() {
  if [[ "${MAGENTO_STATIC_CONTENT_DEPLOY:-false}" != "true" ]] && [[ "${MAGENTO_SCD_ON_DEMAND:-false}" != "true" ]]; then
    return 0
  fi

  local ARGS=("--jobs=$(nproc)")

  local SCD_ARGS="-v"
  if version_gt "${MAGENTO_VERSION:-'2.4.4'}" "2.3.99" && [[ "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE:-true}" == "true" ]]; then
    SCD_ARGS="-fv"
  fi
  ARGS+=("${SCD_ARGS}")

  if [[ -n ${MAGENTO_THEMES:-} ]]; then
    read -r -a themes <<<"$MAGENTO_THEMES"
    local THEME_ARGS=$(printf -- '--theme=%s ' "${themes[@]}")
    # Remove trailing space
    THEME_ARGS=${THEME_ARGS% }
    ARGS+=("${THEME_ARGS}")
  fi

  if [[ -n "${MAGENTO_LANGUAGES:-}" ]]; then
    ARGS+=("${MAGENTO_LANGUAGES:-}")
  fi

  log "Deploying static content"
  magento setup:static-content:deploy "${ARGS[@]}"
}

magento_cache_enable() {
  log "Enabling cache"
  magento cache:enable
}

magento_reindex() {
  if [[ "${MAGENTO_SKIP_REINDEX:-true}" == "true" ]]; then
    return 0
  fi

  log "Running indexer:reindex"
  magento indexer:reindex
}

magento_upgrade_required() {
  log "Checking if Magento upgrade is required"
  if magento setup:db:status; then
    false
  fi
}

magento_maintenance_disable() {
  log "Disabling maintenance mode"
  magento maintenance:disable
}

magento_maintenance_enable() {
  log "Enabling maintenance mode"
  magento maintenance:enable
}

magento_setup_upgrade() {
  if [[ "${MAGENTO_SKIP_UPGRADE:-false}" == "true" ]]; then
    return 0
  fi

  if ! magento_upgrade_required; then
    log "Magento upgrade is not required"
    return 0
  fi

  log "Running Magento setup:upgrade"
  magento_maintenance_enable
  magento setup:upgrade --keep-generated
  magento_maintenance_disable
}

magento_deploy_mode_set() {
  if [[ "${MAGENTO_MODE:-default}" == "default" ]]; then
    return 0
  fi

  log "Setting Magento deploy mode to ${MAGENTO_MODE:-default}"
  magento deploy:mode:set "${MAGENTO_MODE:-default}"
}

magento_secure_frontend() {
  if [[ "${MAGENTO_ENABLE_HTTPS:-true}" != "true" ]]; then
    return 0
  fi

  log "Enabling HTTPS for frontend"
  magento config:set "web/secure/use_in_frontend" 1
}

magento_secure_backend() {
  if [[ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" != "true" ]]; then
    return 0
  fi

  log "Enabling HTTPS for admin"
  magento config:set "web/secure/use_in_adminhtml" 1
}

magento_use_rewrites() {
  if [[ "${MAGENTO_USE_REWRITES:-true}" != "true" ]]; then
    return 0
  fi

  log "Enabling rewrites"
  magento config:set "web/seo/use_rewrites" 1
}

magento_admin_user_exists() {
  if ! magerun admin:user:list --format=csv | tail -n +2 | awk -F',' '{print $2}' | grep "^${MAGENTO_USERNAME:-admin}$" >/dev/null; then
    false
  fi
}

magento_admin_user_inactive() {
  if ! magerun admin:user:list --format=csv | tail -n +2 | awk -F',' '{print $2,$4}' | grep "^${MAGENTO_USERNAME:-admin} inactive$" >/dev/null; then
    false
  fi
}

magento_admin_user() {
  if magento_admin_user_exists; then
    log "Admin user already exists, updating password"
    magerun admin:user:change-password "${MAGENTO_USERNAME:-admin}" "${MAGENTO_PASSWORD:-ASDqwe123}"

    if [[ "${MAGENTO_ACTIVATE_INACTIVE_ADMIN_USER:-true}" == "true" ]] && magento_admin_user_inactive; then
      log "Admin user is inactive, activating"
      magerun admin:user:activate "${MAGENTO_USERNAME:-admin}"
    fi

    return 0
  fi

  log "Creating admin user"
  local ARGS=("")
  ARGS=(
    "--admin-firstname=${MAGENTO_FIRST_NAME:-admin}"
    "--admin-lastname=${MAGENTO_LAST_NAME:-admin}"
    "--admin-email=${MAGENTO_EMAIL:-admin@example.com}"
    "--admin-user=${MAGENTO_USERNAME:-admin}"
    "--admin-password=${MAGENTO_PASSWORD:-ASDqwe123}"
  )

  magerun admin:user:delete --force "${MAGENTO_USERNAME:-admin}" || true
  magerun admin:user:delete --force "${MAGENTO_EMAIL:-admin@example.com}" || true
  magerun admin:user:create "${ARGS[@]}"
}

magento_disable_deploy_sample_data() {
  log "Disabling sample data deployment"
  export MAGENTO_DEPLOY_SAMPLE_DATA=false
}

magento_deploy_sample_data() {
  if [[ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" != "true" ]]; then
    return 0
  fi

  log "Deploying sample data"
  magento sampledata:deploy
  magento setup:upgrade --keep-generated
  magento_setup_static_content_deploy
}

magento_publish_config() {
  log "Publishing configuration from $(app_path)/app/etc/env.php to $(shared_config_path)/app/etc/env.php"
  mkdir -p "$(shared_config_path)/app/etc"
  cp -a "$(app_path)/app/etc/env.php" "$(shared_config_path)/app/etc/env.php"
}

main() {
  LOCKFILE="$(shared_config_path)/.deploy.lock"
  readonly LOCKFILE

  trap 'lock_cleanup ${LOCKFILE}' EXIT
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  lock_acquire "${LOCKFILE}"

  run_hooks "pre-install"

  check_requirements

  conditional_sleep
  command_before_install
  bootstrap_check
  composer_configure

  if magento_is_installed; then
    magento_configure
    magento_setup_upgrade
    magento_disable_deploy_sample_data
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
  magento_cache_enable
  magento_reindex
  magento_publish_config

  command_after_install

  run_hooks "post-install"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

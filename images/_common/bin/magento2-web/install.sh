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
  if [[ -f "$lib_path" ]]; then
    FUNCTIONS_FILE="$lib_path"
    break
  fi
done

if [[ -f "$FUNCTIONS_FILE" ]]; then
  # shellcheck source=/dev/null
  source "$FUNCTIONS_FILE"
else
  printf "\033[1;31m%s ERROR: Required file %s not found\033[0m\n" "$(date --iso-8601=seconds)" "$FUNCTIONS_FILE" >&2
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

: "${PHP_ARGS:=-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} -dmemory_limit=${PHP_MEMORY_LIMIT:-2G}}"
: "${MAGENTO_COMMAND:=php ${PHP_ARGS} ${_magento_command} --no-ansi --no-interaction}"
: "${MAGERUN_COMMAND:=php ${PHP_ARGS} ${_magerun_command} --no-ansi --no-interaction}"
: "${COMPOSER_COMMAND:=php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"

unset PHP_ARGS _magento_command _magerun_command _composer_command

: "${COMPOSER_AUTH:=}"
: "${MAGENTO_PUBLIC_KEY:=}"
: "${MAGENTO_PRIVATE_KEY:=}"
: "${GITHUB_USER:=}"
: "${GITHUB_TOKEN:=}"
: "${BITBUCKET_PUBLIC_KEY:=}"
: "${BITBUCKET_PRIVATE_KEY:=}"
: "${GITLAB_TOKEN:=}"

: "${COMMAND_BEFORE_INSTALL:=}"
: "${COMMAND_AFTER_INSTALL:=}"
: "${SKIP_BOOTSTRAP:=false}"

: "${MAGENTO_SKIP_BOOTSTRAP:=false}"
: "${MAGENTO_VERSION:=2.4.4}"
: "${MAGENTO_HOST:=magento.test}"
: "${MAGENTO_BASE_URL:="http://${MAGENTO_HOST}"}"
: "${MAGENTO_BASE_URL_SECURE:="https://${MAGENTO_HOST}"}"
: "${MAGENTO_ENABLE_HTTPS:=true}"
: "${MAGENTO_ENABLE_ADMIN_HTTPS:=true}"
: "${MAGENTO_USE_REWRITES:=true}"
: "${MAGENTO_KEY:=12345678901234567890123456789012}"
: "${MAGENTO_ADMIN_URL_PREFIX:=admin}"

: "${MAGENTO_DATABASE_HOST:=db}"
: "${MAGENTO_DATABASE_PORT:=3306}"
: "${MAGENTO_DATABASE_USER:=magento}"
: "${MAGENTO_DATABASE_PASSWORD:=magento}"
: "${MAGENTO_DATABASE_NAME:=magento}"

: "${MAGENTO_REDIS_ENABLED:=false}"
: "${MAGENTO_REDIS_HOST:=redis}"
: "${MAGENTO_REDIS_PORT:=6379}"
: "${MAGENTO_REDIS_PASSWORD:=redis}"
: "${MAGENTO_SESSION_SAVE:=redis}"
: "${MAGENTO_SESSION_SAVE_REDIS_HOST:=${MAGENTO_REDIS_HOST}}"
: "${MAGENTO_SESSION_SAVE_REDIS_PORT:=${MAGENTO_REDIS_PORT}}"
: "${MAGENTO_SESSION_SAVE_REDIS_DB:=2}"
: "${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY:=20}"
: "${MAGENTO_SESSION_SAVE_REDIS_PASSWORD:=${MAGENTO_REDIS_PASSWORD}}"
: "${MAGENTO_CACHE_BACKEND:=redis}"
: "${MAGENTO_CACHE_BACKEND_REDIS_SERVER:=${MAGENTO_REDIS_HOST}}"
: "${MAGENTO_CACHE_BACKEND_REDIS_PORT:=${MAGENTO_REDIS_PORT}}"
: "${MAGENTO_CACHE_BACKEND_REDIS_DB:=0}"
: "${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD:=${MAGENTO_REDIS_PASSWORD}}"
: "${MAGENTO_PAGE_CACHE:=redis}"
: "${MAGENTO_PAGE_CACHE_REDIS_SERVER:=${MAGENTO_REDIS_HOST}}"
: "${MAGENTO_PAGE_CACHE_REDIS_PORT:=${MAGENTO_REDIS_PORT}}"
: "${MAGENTO_PAGE_CACHE_REDIS_DB:=1}"
: "${MAGENTO_PAGE_CACHE_REDIS_PASSWORD:=${MAGENTO_REDIS_PASSWORD}}"

: "${MAGENTO_VARNISH_ENABLED:=false}"
: "${MAGENTO_VARNISH_HOST:=varnish}"
: "${MAGENTO_VARNISH_PORT:=80}"

: "${MAGENTO_RABBITMQ_ENABLED:=false}"
: "${MAGENTO_AMQP_HOST:=rabbitmq}"
: "${MAGENTO_AMQP_PORT:=5672}"
: "${MAGENTO_AMQP_USER:=guest}"
: "${MAGENTO_AMQP_PASSWORD:=guest}"
: "${MAGENTO_AMQP_VIRTUAL_HOST:=/}"
: "${MAGENTO_CONSUMERS_WAIT_FOR_MESSAGES:=0}"

: "${MAGENTO_ELASTICSEARCH_ENABLED:=false}"
: "${MAGENTO_ELASTICSEARCH_HOST:=elasticsearch}"
: "${MAGENTO_ELASTICSEARCH_PORT:=9200}"
: "${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:=0}"
: "${MAGENTO_ELASTICSEARCH_TIMEOUT:=15}"
: "${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:=magento2}"
: "${MAGENTO_OPENSEARCH_ENABLED:=false}"
: "${MAGENTO_OPENSEARCH_HOST:=opensearch}"
: "${MAGENTO_OPENSEARCH_PORT:=9200}"
: "${MAGENTO_OPENSEARCH_ENABLE_AUTH:=0}"
: "${MAGENTO_OPENSEARCH_TIMEOUT:=15}"
: "${MAGENTO_OPENSEARCH_INDEX_PREFIX:=magento2}"
# Search engine will be configured later
: "${MAGENTO_SEARCH_ENGINE:=}"

: "${MAGENTO_DEPLOY_SAMPLE_DATA:=false}"
: "${MAGENTO_FORCE_DEPLOY_SAMPLE_DATA:=false}"
: "${MAGENTO_EXTRA_INSTALL_ARGS:=}"
: "${MAGENTO_SKIP_INSTALL:=false}"
: "${MAGENTO_DI_COMPILE:=false}"
: "${MAGENTO_DI_COMPILE_ON_DEMAND:=false}"
: "${MAGENTO_STATIC_CONTENT_DEPLOY:=false}"
: "${MAGENTO_SCD_ON_DEMAND:=false}"
: "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE:=true}"
: "${MAGENTO_MODE:=default}"
: "${MAGENTO_THEMES:=}"
: "${MAGENTO_LANGUAGES:=}"
: "${MAGENTO_SKIP_REINDEX:=true}"
: "${MAGENTO_SKIP_UPGRADE:=false}"

: "${MAGENTO_USERNAME:=admin}"
: "${MAGENTO_PASSWORD:=ASDqwe123}"
: "${MAGENTO_ACTIVATE_INACTIVE_ADMIN_USER:=true}"
: "${MAGENTO_FIRST_NAME:=admin}"
: "${MAGENTO_LAST_NAME:=admin}"
: "${MAGENTO_EMAIL:=admin@example.com}"

: "${MAGENTO_SHARED_FILES:=app/etc/env.php}"

magento() {
  "$MAGENTO_COMMAND" "$@"
}

magerun() {
  "$MAGERUN_COMMAND" "$@"
}

composer() {
  "$COMPOSER_COMMAND" "$@"
}

check_requirements() {
  check_command "composer"
  check_command "mr"
}

command_before_install() {
  if [[ -z "$COMMAND_BEFORE_INSTALL" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "$COMMAND_BEFORE_INSTALL"
}

command_after_install() {
  if [[ -z "$COMMAND_AFTER_INSTALL" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "$COMMAND_AFTER_INSTALL"
}

composer_configure() {
  if [[ -n "$COMPOSER_AUTH" ]]; then
    # HACK: workaround for
    # https://github.com/composer/composer/issues/12084
    # shellcheck disable=SC2016
    log '$COMPOSER_AUTH is set, skipping Composer configuration'

    return 0
  fi

  log "Configuring Composer"

  if [[ -n "$MAGENTO_PUBLIC_KEY" ]] && [[ -n "$MAGENTO_PRIVATE_KEY" ]]; then
    composer global config http-basic.repo.magento.com "$MAGENTO_PUBLIC_KEY" "$MAGENTO_PRIVATE_KEY"
  fi

  if [[ -n "$GITHUB_USER" ]] && [[ -n "$GITHUB_TOKEN" ]]; then
    composer global config http-basic.github.com "$GITHUB_USER" "$GITHUB_TOKEN"
  fi

  if [[ -n "$BITBUCKET_PUBLIC_KEY" ]] && [[ -n "$BITBUCKET_PRIVATE_KEY" ]]; then
    composer global config bitbucket-oauth.bitbucket.org "$BITBUCKET_PUBLIC_KEY" "$BITBUCKET_PRIVATE_KEY"
  fi

  if [[ -n "$GITLAB_TOKEN" ]]; then
    composer global config gitlab-token.gitlab.com "$GITLAB_TOKEN"
  fi
}

composer_configure_home_for_magento() {
  mkdir -p "$(app_path)/var/composer_home"

  local composer_home
  composer_home="$(composer config --global home)"
  if [[ -n "$composer_home" ]]; then
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

bootstrap_check() {
  if [[ "$MAGENTO_SKIP_BOOTSTRAP" != "true" ]] && [[ "$SKIP_BOOTSTRAP" != "true" ]]; then
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
  ARGS+=(
    "--base-url=${MAGENTO_BASE_URL}"
    "--base-url-secure=${MAGENTO_BASE_URL_SECURE}"
  )
  if [[ "$MAGENTO_ENABLE_HTTPS" == "true" ]]; then
    ARGS+=(
      "--use-secure=1"
    )
  fi
  if [[ "$MAGENTO_ENABLE_ADMIN_HTTPS" == "true" ]]; then
    ARGS+=(
      "--use-secure-admin=1"
    )
  fi
  if [[ "$MAGENTO_USE_REWRITES" == "true" ]]; then
    ARGS+=(
      "--use-rewrites=1"
    )
  fi
}

magento_args_defaults() {
  ARGS+=(
    "--key=${MAGENTO_KEY}"
    "--backend-frontname=${MAGENTO_ADMIN_URL_PREFIX}"
  )
}

magento_args_db() {
  ARGS+=(
    "--db-host=${MAGENTO_DATABASE_HOST}:${MAGENTO_DATABASE_PORT}"
    "--db-name=${MAGENTO_DATABASE_NAME}"
    "--db-user=${MAGENTO_DATABASE_USER}"
    "--db-password=${MAGENTO_DATABASE_PASSWORD}"
  )
}

magento_args_redis() {
  # Configure Redis
  if [[ "$MAGENTO_REDIS_ENABLED" != "true" ]]; then
    ARGS+=(
      "--session-save=files"
    )
    return 0
  fi

  if [[ "$MAGENTO_SESSION_SAVE" == "redis" ]]; then
    ARGS+=(
      "--session-save=redis"
      "--session-save-redis-host=${MAGENTO_SESSION_SAVE_REDIS_HOST}"
      "--session-save-redis-port=${MAGENTO_SESSION_SAVE_REDIS_PORT}"
      "--session-save-redis-db=${MAGENTO_SESSION_SAVE_REDIS_DB}"
      "--session-save-redis-max-concurrency=${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY}"
    )

    if [[ -n "$MAGENTO_REDIS_PASSWORD" ]] || [[ -n "$MAGENTO_SESSION_SAVE_REDIS_PASSWORD" ]]; then
      ARGS+=(
        "--session-save-redis-password=${MAGENTO_SESSION_SAVE_REDIS_PASSWORD}"
      )
    fi
  fi

  if [[ "$MAGENTO_CACHE_BACKEND" == "redis" ]]; then
    ARGS+=(
      "--cache-backend=redis"
      "--cache-backend-redis-server=${MAGENTO_CACHE_BACKEND_REDIS_SERVER}"
      "--cache-backend-redis-port=${MAGENTO_CACHE_BACKEND_REDIS_PORT}"
      "--cache-backend-redis-db=${MAGENTO_CACHE_BACKEND_REDIS_DB}"
    )

    if [[ -n "$MAGENTO_REDIS_PASSWORD" ]] || [[ -n "$MAGENTO_CACHE_BACKEND_REDIS_PASSWORD" ]]; then
      ARGS+=(
        "--cache-backend-redis-password=${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD}"
      )
    fi
  fi

  if [[ "$MAGENTO_PAGE_CACHE" == "redis" ]]; then
    ARGS+=(
      "--page-cache=redis"
      "--page-cache-redis-server=${MAGENTO_PAGE_CACHE_REDIS_SERVER}"
      "--page-cache-redis-port=${MAGENTO_PAGE_CACHE_REDIS_PORT}"
      "--page-cache-redis-db=${MAGENTO_PAGE_CACHE_REDIS_DB}"
    )
    if [[ -n "$MAGENTO_REDIS_PASSWORD" ]] || [[ -n "$MAGENTO_PAGE_CACHE_REDIS_PASSWORD" ]]; then
      ARGS+=(
        "--page-cache-redis-password=${MAGENTO_PAGE_CACHE_REDIS_PASSWORD}"
      )
    fi
  fi
}

magento_args_varnish() {
  # Configure Varnish
  if [[ "$MAGENTO_VARNISH_ENABLED" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--http-cache-hosts=${MAGENTO_VARNISH_HOST}:${MAGENTO_VARNISH_PORT}"
  )
}

magento_args_rabbitmq() {
  # Configure RabbitMQ
  if [[ "$MAGENTO_RABBITMQ_ENABLED" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--amqp-host=${MAGENTO_AMQP_HOST}"
    "--amqp-port=${MAGENTO_AMQP_PORT}"
    "--amqp-user=${MAGENTO_AMQP_USER}"
    "--amqp-password=${MAGENTO_AMQP_PASSWORD}"
    "--amqp-virtualhost=${MAGENTO_AMQP_VIRTUAL_HOST}"
  )

  if version_gt "$MAGENTO_VERSION" "2.3.99"; then
    ARGS+=(
      "--consumers-wait-for-messages=${MAGENTO_CONSUMERS_WAIT_FOR_MESSAGES}"
    )
  fi
}

magento_args_search() {
  if version_gt "2.3.99" "$MAGENTO_VERSION" && [[ "$MAGENTO_ELASTICSEARCH_ENABLED" != "true" ]] && [[ "$MAGENTO_OPENSEARCH_ENABLED" != "true" ]]; then
    return 0
  fi

  if [[ "$MAGENTO_OPENSEARCH_ENABLED" == "true" ]] || [[ "$MAGENTO_SEARCH_ENGINE" == "opensearch" ]]; then
    ARGS+=(
      "--search-engine=${MAGENTO_SEARCH_ENGINE:-opensearch}"
      "--opensearch-host=${MAGENTO_OPENSEARCH_HOST}"
      "--opensearch-port=${MAGENTO_OPENSEARCH_PORT}"
      "--opensearch-index-prefix=${MAGENTO_OPENSEARCH_INDEX_PREFIX}"
      "--opensearch-enable-auth=${MAGENTO_OPENSEARCH_ENABLE_AUTH}"
      "--opensearch-timeout=${MAGENTO_OPENSEARCH_TIMEOUT}"
    )

    search_configured

    return 0
  fi

  # Elasticsearch 7 is required for Magento 2.4.0+ or later (if not using OpenSearch)
  if version_gt "$MAGENTO_VERSION" "2.3.99" || [[ "$MAGENTO_ELASTICSEARCH_ENABLED" == "true" ]]; then
    MAGENTO_ELASTICSEARCH_ENABLED="true"

    ARGS+=(
      "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
      "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST}"
      "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT}"
      "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX}"
      "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH}"
      "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT}"
    )
  fi

  search_configured
}

search_configured() {
  export SEARCH_CONFIGURED=true
}

magento_args_sample_data() {
  if [[ "$MAGENTO_DEPLOY_SAMPLE_DATA" != "true" ]]; then
    return 0
  fi

  ARGS+=(
    "--use-sample-data"
  )
}

magento_args_extra() {
  if [[ -z "$MAGENTO_EXTRA_INSTALL_ARGS" ]]; then
    return 0
  fi

  ARGS+=(
    "$MAGENTO_EXTRA_INSTALL_ARGS"
  )
}

magento_setup_install() {
  if [[ "$MAGENTO_SKIP_INSTALL" == "true" ]]; then
    return 0
  fi

  log "Installing Magento ${MAGENTO_VERSION}"

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
  magento setup:install "${ARGS[@]}"

  magento_configure_search
}

magento_configure() {
  log "Configuring Magento"

  install_date="$(magerun config:env:show install.date 2>/dev/null || true)"
  cache_graphql_id_salt="$(magerun config:env:show cache.graphql.id_salt 2>/dev/null || true)"

  rm -f "$(app_path)/app/etc/env.php"

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
  magento setup:config:set "${ARGS[@]}"

  if [[ -z "${install_date:+x}" ]]; then install_date=$(date -R); fi
  magerun config:env:set install.date "$install_date"

  if [[ -n "$cache_graphql_id_salt" ]]; then magerun config:env:set cache.graphql.id_salt "$cache_graphql_id_salt"; fi

  magento_app_config_import

  magento_configure_search
}

magento_search_configurable() {
  if ! magento setup:config:set --help | grep -q '\-\-search-engine'; then
    false
  fi
}

magento_app_config_import() {
  magento app:config:import
}

magento_configure_search() {
  if [[ "${SEARCH_CONFIGURED:-false}" == "true" ]]; then
    return 0
  fi

  if [[ "$MAGENTO_ELASTICSEARCH_ENABLED" != "true" ]] && [[ "$MAGENTO_OPENSEARCH_ENABLED" != "true" ]]; then
    return 0
  fi

  if [[ "$MAGENTO_OPENSEARCH_ENABLED" == "true" ]] || [[ "$MAGENTO_SEARCH_ENGINE" == "opensearch" ]]; then
    magento config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-opensearch}"
    magento config:set --lock-env "catalog/search/opensearch_server_hostname" "$MAGENTO_OPENSEARCH_HOST"
    magento config:set --lock-env "catalog/search/opensearch_server_port" "$MAGENTO_OPENSEARCH_PORT"
    magento config:set --lock-env "catalog/search/opensearch_index_prefix" "$MAGENTO_OPENSEARCH_INDEX_PREFIX"
    magento config:set --lock-env "catalog/search/opensearch_enable_auth" "$MAGENTO_OPENSEARCH_ENABLE_AUTH"
    magento config:set --lock-env "catalog/search/opensearch_server_timeout" "$MAGENTO_OPENSEARCH_TIMEOUT"

    return 0
  fi

  # Elasticsearch 7 is required for Magento 2.4.0+ or later (if not using OpenSearch)
  if version_gt "$MAGENTO_VERSION" "2.3.99"; then
    magento config:set --lock-env "catalog/search/engine" "${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_hostname" "$MAGENTO_ELASTICSEARCH_HOST"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_port" "$MAGENTO_ELASTICSEARCH_PORT"
    magento config:set --lock-env "catalog/search/elasticsearch7_index_prefix" "$MAGENTO_ELASTICSEARCH_INDEX_PREFIX"
    magento config:set --lock-env "catalog/search/elasticsearch7_enable_auth" "$MAGENTO_ELASTICSEARCH_ENABLE_AUTH"
    magento config:set --lock-env "catalog/search/elasticsearch7_server_timeout" "$MAGENTO_ELASTICSEARCH_TIMEOUT"
  fi
}

magento_setup_di_compile() {
  # Skip DI compile if it's not enabled explicitly (as it should be part of the build by default)
  if [[ "$MAGENTO_DI_COMPILE" != "true" ]] && [[ "$MAGENTO_DI_COMPILE_ON_DEMAND" != "true" ]]; then
    return 0
  fi

  log "Compiling Magento dependencies"
  magento setup:di:compile
}

magento_setup_static_content_deploy() {
  if [[ "$MAGENTO_STATIC_CONTENT_DEPLOY" != "true" ]] && [[ "$MAGENTO_SCD_ON_DEMAND" != "true" ]]; then
    return 0
  fi

  local ARGS=("--jobs=$(nproc)")

  local SCD_ARGS="-v"
  if version_gt "$MAGENTO_VERSION" "2.3.99" && [[ "$MAGENTO_STATIC_CONTENT_DEPLOY_FORCE" == "true" ]]; then
    SCD_ARGS="-fv"
  fi
  ARGS+=("$SCD_ARGS")

  if [[ -n ${MAGENTO_THEMES} ]]; then
    read -r -a themes <<<"$MAGENTO_THEMES"
    local THEME_ARGS=$(printf -- '--theme=%s ' "${themes[@]}")
    # Remove trailing space
    THEME_ARGS=${THEME_ARGS% }
    ARGS+=("$THEME_ARGS")
  fi

  if [[ -n "$MAGENTO_LANGUAGES" ]]; then
    ARGS+=("$MAGENTO_LANGUAGES")
  fi

  log "Deploying static content"
  magento setup:static-content:deploy "${ARGS[@]}"
}

magento_cache_enable() {
  log "Enabling cache"
  magento cache:enable
}

magento_reindex() {
  if [[ "$MAGENTO_SKIP_REINDEX" == "true" ]]; then
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
  if [[ "$MAGENTO_SKIP_UPGRADE" == "true" ]]; then
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
  if [[ "$MAGENTO_MODE" == "default" ]]; then
    return 0
  fi

  log "Setting Magento deploy mode to ${MAGENTO_MODE}"
  magento deploy:mode:set "$MAGENTO_MODE"
}

magento_secure_frontend() {
  if [[ "$MAGENTO_ENABLE_HTTPS" != "true" ]]; then
    return 0
  fi

  log "Enabling HTTPS for frontend"
  magento config:set "web/secure/use_in_frontend" 1
}

magento_secure_backend() {
  if [[ "$MAGENTO_ENABLE_ADMIN_HTTPS" != "true" ]]; then
    return 0
  fi

  log "Enabling HTTPS for admin"
  magento config:set "web/secure/use_in_adminhtml" 1
}

magento_use_rewrites() {
  if [[ "$MAGENTO_USE_REWRITES" != "true" ]]; then
    return 0
  fi

  log "Enabling rewrites"
  magento config:set "web/seo/use_rewrites" 1
}

magento_admin_user_exists() {
  if ! magerun admin:user:list --format=csv | tail -n +2 | awk -F',' '{print $2}' | grep "^${MAGENTO_USERNAME}$" >/dev/null; then
    false
  fi
}

magento_admin_user_inactive() {
  if ! magerun admin:user:list --format=csv | tail -n +2 | awk -F',' '{print $2,$4}' | grep "^${MAGENTO_USERNAME} inactive$" >/dev/null; then
    false
  fi
}

magento_admin_user() {
  if magento_admin_user_exists; then
    log "Admin user already exists, updating password"
    magerun admin:user:change-password "$MAGENTO_USERNAME" "$MAGENTO_PASSWORD"

    if [[ "$MAGENTO_ACTIVATE_INACTIVE_ADMIN_USER" == "true" ]] && magento_admin_user_inactive; then
      log "Admin user is inactive, activating"
      magerun admin:user:activate "$MAGENTO_USERNAME"
    fi

    return 0
  fi

  log "Creating admin user"
  local ARGS=("")
  ARGS=(
    "--admin-firstname=${MAGENTO_FIRST_NAME}"
    "--admin-lastname=${MAGENTO_LAST_NAME}"
    "--admin-email=${MAGENTO_EMAIL}"
    "--admin-user=${MAGENTO_USERNAME}"
    "--admin-password=${MAGENTO_PASSWORD}"
  )

  magerun admin:user:delete --force "$MAGENTO_USERNAME" || true
  magerun admin:user:delete --force "$MAGENTO_EMAIL" || true
  magerun admin:user:create "${ARGS[@]}"
}

magento_disable_deploy_sample_data() {
  log "Disabling sample data deployment"
  export MAGENTO_DEPLOY_SAMPLE_DATA=false
}

magento_deploy_sample_data() {
  if [[ "$MAGENTO_DEPLOY_SAMPLE_DATA" != "true" ]] && [[ "$MAGENTO_FORCE_DEPLOY_SAMPLE_DATA" != "true" ]]; then
    return 0
  fi

  log "Deploying sample data"
  magento sampledata:deploy
  magento setup:upgrade --keep-generated
  magento_setup_static_content_deploy
}

magento_publish_shared_files() {
  log "Publishing config"

  local _shared_files="$MAGENTO_SHARED_FILES"

  publish_shared_files
}

main() {
  conditional_sleep

  LOCKFILE="$(shared_config_path)/.deploy.lock"
  readonly LOCKFILE

  trap 'lock_cleanup ${LOCKFILE}' EXIT
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  lock_acquire "$LOCKFILE"

  run_hooks "pre-install"

  check_requirements

  command_before_install
  bootstrap_check
  composer_configure
  composer_configure_home_for_magento
  composer_configure_plugins

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
  magento_publish_shared_files

  command_after_install

  run_hooks "post-install"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

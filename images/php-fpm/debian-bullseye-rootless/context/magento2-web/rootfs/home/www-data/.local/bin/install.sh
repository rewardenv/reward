#!/usr/bin/env bash
set -e
function version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

if [ -n "${COMMAND_BEFORE_INSTALL-}" ]; then eval "${COMMAND_BEFORE_INSTALL-}"; fi

if [ "${MAGENTO_SKIP_BOOTSTRAP:-false}" = "true" ]; then
  if [ -n "${COMMAND_AFTER_INSTALL-}" ]; then eval "${COMMAND_AFTER_INSTALL-}"; fi
  exit
fi

if [ "${MAGENTO_SKIP_INSTALL:-false}" != "true" ]; then
  MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.4'}
  MAGENTO_HOST=${MAGENTO_HOST:-'magento.test'}
  MAGENTO_BASE_URL=${MAGENTO_BASE_URL:-"http://$MAGENTO_HOST"}
  MAGENTO_BASE_URL_SECURE=${MAGENTO_BASE_URL_SECURE:-"https://$MAGENTO_HOST"}

  ARGS=()
  ARGS+=(
    "--base-url=${MAGENTO_BASE_URL}"
    "--base-url-secure=${MAGENTO_BASE_URL_SECURE}"
    "--key=${MAGENTO_KEY:-12345678901234567890123456789012}"
    "--backend-frontname=${MAGENTO_ADMIN_URL_PREFIX:-admin}"
    "--db-host=${MAGENTO_DATABASE_HOST:-db}"
    "--db-name=${MAGENTO_DATABASE_NAME:-magento}"
    "--db-user=${MAGENTO_DATABASE_USER:-magento}"
    "--db-password=${MAGENTO_DATABASE_PASSWORD:-magento}"
  )

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

  # Configure Varnish
  if [ "${MAGENTO_VARNISH_ENABLED:-true}" = "true" ]; then
    ARGS+=(
      "--http-cache-hosts=${MAGENTO_VARNISH_HOST:-varnish}:${MAGENTO_VARNISH_PORT:-80}"
    )
  fi

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

  # Configure Elasticsearch
  if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" = "true" ]; then
    if version_gt "${MAGENTO_VERSION}" "2.3.99"; then
      ARGS+=(
        "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
        "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST:-opensearch}"
        "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT:-9200}"
        "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
        "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
        "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
      )
    fi
  fi

  if [ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
    ARGS+=(
      "--use-sample-data"
    )
  fi

  if [ -n "${MAGENTO_EXTRA_INSTALL_ARGS:-}" ]; then
    ARGS+=(
      "${MAGENTO_EXTRA_INSTALL_ARGS}"
    )
  fi

  mr setup:install --no-interaction ${ARGS[@]}

  if [ "${MAGENTO_DI_COMPILE:-false}" = "true" ] || [ "${MAGENTO_DI_COMPILE_ON_DEMAND:-false}" = "true" ]; then
    mr setup:di:compile --no-interaction --ansi
  fi

  if [ "${MAGENTO_STATIC_CONTENT_DEPLOY:-false}" = "true" ] || [ "${MAGENTO_SCD_ON_DEMAND:-false}" = "true" ]; then
    if [ "${MAGENTO_STATIC_CONTENT_DEPLOY_FORCE}" = "true" ]; then
      bin/magento setup:static-content:deploy --no-interaction --jobs=$(nproc) -fv ${MAGENTO_LANGUAGES:-}
    else
      bin/magento setup:static-content:deploy --no-interaction --jobs=$(nproc) -v ${MAGENTO_LANGUAGES:-}
    fi
  fi
fi

if [ "${MAGENTO_MODE:-default}" != "default" ]; then
  mr deploy:mode:set --no-interaction "${MAGENTO_MODE}"
fi

if [ "${MAGENTO_ENABLE_HTTPS:-true}" = "true" ]; then
  mr config:set --no-interaction "web/secure/use_in_frontend" 1
fi

if [ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" = "true" ]; then
  mr config:set --no-interaction "web/secure/use_in_adminhtml" 1
fi

if [ "${MAGENTO_USE_REWRITES:-true}" = "true" ]; then
  mr config:set --no-interaction "web/seo/use_rewrites" 1
fi

if mr admin:user:list --no-interaction --format=csv | tail -n +2 | awk -F',' '{print $2}' | grep "^${MAGENTO_USERNAME:-admin}$" >/dev/null; then
  mr admin:user:change-password --no-interaction ${MAGENTO_USERNAME:-admin} ${MAGENTO_PASSWORD:-ASDqwe123}
else
  ARGS=()
  ARGS+=(
    "--admin-firstname=${MAGENTO_FIRST_NAME:-admin}"
    "--admin-lastname=${MAGENTO_LAST_NAME:-admin}"
    "--admin-email=${MAGENTO_EMAIL:-admin@example.com}"
    "--admin-user=${MAGENTO_USERNAME:-admin}"
    "--admin-password=${MAGENTO_PASSWORD:-ASDqwe123}"
  )
  mr admin:user:delete --force --no-interaction "${MAGENTO_USERNAME:-admin}" || true
  mr admin:user:delete --force --no-interaction "${MAGENTO_EMAIL:-admin@example.com}" || true
  mr admin:user:create --no-interaction "${ARGS[@]}"
fi

if [ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
  mr sampledata:deploy --no-interaction
  mr setup:upgrade --no-interaction --keep-generated
fi

if [ "${MAGENTO_DEPLOY_STATIC_CONTENT:-false}" = "true" ]; then
  mr setup:static-content:deploy --no-interaction --jobs="$(nproc)" -fv "${MAGENTO_LANGUAGES:-}"
fi

if [ "${MAGENTO_SKIP_REINDEX:-true}" != "true" ]; then
  mr indexer:reindex --no-interaction
fi

if [ -n "${COMMAND_AFTER_INSTALL-}" ]; then eval "${COMMAND_AFTER_INSTALL-}"; fi

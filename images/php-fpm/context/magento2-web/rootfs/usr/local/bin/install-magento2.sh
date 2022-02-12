#!/bin/bash
set -e
function version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

if [ "${MAGENTO_SKIP_BOOTSTRAP:-false}" == "true" ]; then
  exit
fi

MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.3-p1'}

MAGENTO_HOST=${MAGENTO_HOST:-'magento.test'}

MAGENTO_BASE_URL=${MAGENTO_BASE_URL:-"http://$MAGENTO_HOST"}
MAGENTO_BASE_URL_SECURE=${MAGENTO_BASE_URL_SECURE:-"https://$MAGENTO_HOST"}

ARGS=()
ARGS+=(
  "--base-url=${MAGENTO_BASE_URL}"
  "--base-url-secure=${MAGENTO_BASE_URL_SECURE}"
  "--key=${MAGENTO_KEY:-12345678901234567890123456789012}"
  "--backend-frontname=${MAGENTO_ADMIN_URL_PREFIX:-admin}"
  "--db-host=${MAGENTO_DATABASE_HOST:-mysql}"
  "--db-name=${MAGENTO_DATABASE_NAME:-magento}"
  "--db-user=${MAGENTO_DATABASE_USER:-magento}"
  "--db-password=${MAGENTO_DATABASE_PASSWORD:-magento}"
  "--admin-firstname=${MAGENTO_FIRST_NAME:-admin}"
  "--admin-lastname=${MAGENTO_LAST_NAME:-admin}"
  "--admin-email=${MAGENTO_EMAIL:-admin\@example.com}"
  "--admin-user=${MAGENTO_USERNAME:-admin}"
  "--admin-password=${MAGENTO_PASSWORD:-ASDFqwer1234}"
)

MAGENTO_DEPLOY_STATIC_CONTENT
MAGENTO_SKIP_REINDEX

# Configure Redis
if [ "${MAGENTO_REDIS_ENABLED:-true}" == "true" ]; then
  MAGENTO_REDIS_HOST=${MAGENTO_REDIS_HOST:-redis}
  MAGENTO_REDIS_PORT=${MAGENTO_REDIS_PORT:-6379}

  ARGS+=(
    "--session-save=redis"
    "--session-save-redis-host=${MAGENTO_SESSION_SAVE_REDIS_HOST:-$MAGENTO_REDIS_HOST}"
    "--session-save-redis-port=${MAGENTO_SESSION_SAVE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
    "--session-save-redis-db=${MAGENTO_SESSION_SAVE_REDIS_SESSION_DB:-2}"
    "--session-save-redis-max-concurrency=${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY:-20}"
    "--cache-backend=redis"
    "--cache-backend-redis-server=${MAGENTO_CACHE_BACKEND_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
    "--cache-backend-redis-port=${MAGENTO_CACHE_BACKEND_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
    "--cache-backend-redis-db=${MAGENTO_CACHE_BACKEND_REDIS_DB:-0}"
    "--page-cache=redis"
    "--page-cache-redis-server=${MAGENTO_PAGE_CACHE_REDIS_SERVER:-$MAGENTO_REDIS_HOST}"
    "--page-cache-redis-port=${MAGENTO_PAGE_CACHE_REDIS_PORT:-$MAGENTO_REDIS_PORT}"
    "--page-cache-redis-db=${MAGENTO_PAGE_CACHEC_REDIS_DB:-1}"
  )

  if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_SESSION_REDIS_PASSWORD} ]]; then
    ARGS+=(
      "--session-save-redis-password=${MAGENTO_SESSION_SAVE_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
    )
  fi
  if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD} ]]; then
    ARGS+=(
      "--cache-backend-redis-password=${MAGENTO_CACHE_BACKEND_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
    )
  fi
  if [[ -n "$MAGENTO_REDIS_PASSWORD" || -n ${MAGENTO_PAGE_CACHE_REDIS_PASSWORD} ]]; then
    ARGS+=(
      "--cache-backend-redis-password=${MAGENTO_PAGE_CACHE_REDIS_PASSWORD:-$MAGENTO_REDIS_PASSWORD}"
    )
  fi
else
  ARGS+=(
    "--session-save=files"
  )
fi

# Configure Varnish
if [ "${MAGENTO_VARNISH_ENABLED:-true}" == "true" ]; then
  ARGS+=(
    "--http-cache-hosts=${MAGENTO_VARNISH_HOST:-varnish}:${MAGENTO_VARNISH_PORT:-80}"
  )
fi

# Configure RabbitMQ
if [ "${MAGENTO_RABBITMQ_ENABLED:-true}" == "true" ]; then
  ARGS+=(
    "--amqp-host=${MAGENTO_AMQP_HOST:-rabbitmq}"
    "--amqp-port=${MAGENTO_AMQP_PORT:-5672}"
    "--amqp-user=${MAGENTO_AMQP_USER:-guest}"
    "--amqp-password=${MAGENTO_AMQP_PASSWORD:-guest}"
  )

  if version_gt "${MAGENTO_VERSION}" "2.3.99"; then
    ARGS+=(
      "--consumers-wait-for-messages=0"
    )
  fi
fi

# Configure Elasticsearch
if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" == "true" ]; then
  ARGS+=(
    "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
    "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST:-elasticsearch}"
    "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT:-9200}"
    "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
    "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
    "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
  )
fi

if [ "${MAGENTO_DEPLOY_SAMPLE_DATA:-false}" == "true" ]; then
  ARGS+=(
    "--use-sample-data"
  )
fi

if [ -n "${MAGENTO_EXTRA_INSTALL_ARGS:-}" ]; then
  ARGS+=(
    "${MAGENTO_EXTRA_INSTALL_ARGS}"
  )
fi

php bin/magento setup:install "${ARGS[@]}"

if [ "${MAGENTO_MODE:-default}" != "default" ]; then
  php bin/magento deploy:mode:set "${MAGENTO_MODE}"
fi

if [ "${MAGENTO_ENABLE_HTTPS:-true}" == "true" ]; then
  php bin/magento config:set "web/secure/use_in_frontend" 1
fi

if [ "${MAGENTO_ENABLE_ADMIN_HTTPS:-true}" == "true" ]; then
  php bin/magento config:set "web/secure/use_in_adminhtml" 1
fi

if [ "${MAGENTO_USE_REWRITES:-true}" == "true" ]; then
  php bin/magento config:set "web/seo/use_rewrites" 1
fi

if [ "${MAGENTO_SKIP_REINDEX:-true}" != "true" ]; then
  php bin/magento indexer:reindex
fi


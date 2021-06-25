#!/bin/bash
MAGENTO_VERSION=${MAGENTO_VERSION:-'2.4.2'}

MAGENTO_DOMAIN=${MAGENTO_DOMAIN:-'magento.test'}

MAGENTO_BASE_URL=${MAGENTO_BASE_URL:-"http://$MAGENTO_DOMAIN"}
MAGENTO_BASE_URL_SECURE=${MAGENTO_BASE_URL_SECURE:-"https://$MAGENTO_DOMAIN"}

ARGS=()
ARGS+=(
  "--base-url=http://${MAGENTO_BASE_URL}"
  "--base-url-secure=https://${MAGENTO_BASE_URL_SECURE}"
  "--key=${MAGENTO_KEY:-12345678901234567890123456789012}"
  "--backend-frontname=${MAGENTO_BACKEND_FRONTNAME:-admin}"
  "--db-host=${MAGENTO_DB_HOST:-mysql}"
  "--db-name=${MAGENTO_DB_NAME:-magento}"
  "--db-user=${MAGENTO_DB_USER:-magento}"
  "--db-password=${MAGENTO_DB_PASSWORD:-magento}"
)

# Configure Redis
if [ "${MAGENTO_REDIS_ENABLED:-true}" = true ]; then
  ARGS+=(
    "--session-save=redis"
    "--session-save-redis-host=${MAGENTO_SESSION_SAVE_REDIS_HOST:-redis}"
    "--session-save-redis-port=${MAGENTO_SESSION_SAVE_REDIS_PORT:-6379}"
    "--session-save-redis-db=${MAGENTO_SESSION_SAVE_REDIS_SESSION_DB:-2}"
    "--session-save-redis-max-concurrency=${MAGENTO_SESSION_SAVE_REDIS_MAX_CONCURRENCY:-20}"
    "--cache-backend=redis"
    "--cache-backend-redis-server=${MAGENTO_CACHE_BACKEND_REDIS_SERVER:-redis}"
    "--cache-backend-redis-db=${MAGENTO_CACHE_BACKEND_REDIS_DB:-0}"
    "--cache-backend-redis-port=${MAGENTO_CACHE_BACKEND_REDIS_PORT:6379}"
    "--page-cache=redis"
    "--page-cache-redis-server=${MAGENTO_PAGE_CACHE_REDIS_SERVER:-redis}"
    "--page-cache-redis-db=${MAGENTO_PAGE_CACHEC_REDIS_DB:-1}"
    "--page-cache-redis-port=${MAGENTO_PAGE_CACHE_REDIS_PORT:-6379}"
  )
else
  ARGS+=(
    "--session-save=files"
  )
fi

# Configure Varnish
if [ "${MAGENTO_VARNISH_ENABLED:-true}" = true ]; then
  ARGS+=(
    "--http-cache-hosts=${MAGENTO_VARNISH_HOST:-varnish}:${MAGENTO_VARNISH_PORT:-80}"
  )
fi

# Configure RabbitMQ
if [ "${MAGENTO_RABBITMQ_ENABLED:-true}" = true ]; then
  ARGS+=(
    "--amqp-host=${MAGENTO_AMQP_HOST:-rabbitmq}"
    "--amqp-port=${MAGENTO_AMQP_PORT:-5672}"
    "--amqp-user=${MAGENTO_AMQP_USER:-guest}"
    "--amqp-password=${MAGENTO_AMQP_PASSWORD:-guest}"
  )

  if [ "${MAGENTO_VERSION//./}" -ge 240 ]; then
    ARGS+=(
      "--consumers-wait-for-messages=0"
    )
  fi
fi

# Configure Elasticsearch
if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-true}" = true ]; then
  ARGS+=(
    "--search-engine=${MAGENTO_SEARCH_ENGINE:-elasticsearch7}"
    "--elasticsearch-host=${MAGENTO_ELASTICSEARCH_HOST:-elasticsearch}"
    "--elasticsearch-port=${MAGENTO_ELASTICSEARCH_PORT:-9200}"
    "--elasticsearch-index-prefix=${MAGENTO_ELASTICSEARCH_INDEX_PREFIX:-magento2}"
    "--elasticsearch-enable-auth=${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}"
    "--elasticsearch-timeout=${MAGENTO_ELASTICSEARCH_TIMEOUT:-15}"
  )
fi

php bin/magento setup:install ${ARGS[@]}

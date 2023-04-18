#!/bin/bash
set -e

echo "Checking database connection..."
if mysql -h"${MAGENTO_DATABASE_HOST:-db}" -P"${MAGENTO_DATABASE_PORT:-3306}" -u"${MAGENTO_DATABASE_USER:-magento}" -p"${MAGENTO_DATABASE_PASSWORD:-magento}" -e "CREATE DATABASE IF NOT EXISTS ${MAGENTO_DATABASE_NAME:-magento}; "; then
  echo "Database connection ready."
else
  echo "Database connection failed."
  exit 1
fi

if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-false}" = "true" ]; then
  echo "Checking Elasticsearch connection..."
  if curl --connect-timeout 10 -fsSL -X GET "http://${MAGENTO_ELASTICSEARCH_HOST:-opensearch}:${MAGENTO_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty" &>/dev/null; then
    echo "Elasticsearch connection ready."
  else
    echo "Elasticsearch connection failed."
    exit 1
  fi
fi

if [ "${MAGENTO_REDIS_ENABLED:-false}" = "true" ]; then
  echo "Checking Redis connection..."
  if nc -v -z "${MAGENTO_REDIS_HOST:-redis}" "${MAGENTO_REDIS_PORT:-6379}"; then
    echo "Redis connection ready."
  else
    echo "Redis connection failed."
    exit 1
  fi
fi

if [ "${MAGENTO_VARNISH_ENABLED:-false}" = "true" ]; then
  echo "Checking Varnish connection..."
  if nc -v -z "${MAGENTO_VARNISH_HOST:-varnish}" "${MAGENTO_VARNISH_PORT:-80}"; then
    echo "Varnish connection ready."
  else
    echo "Varnish connection failed."
    exit 1
  fi
fi

if [ "${MAGENTO_RABBITMQ_ENABLED:-false}" = "true" ]; then
  echo "Checking RabbitMQ connection..."
  if nc -v -z "${MAGENTO_AMQP_HOST:-rabbitmq}" "${MAGENTO_AMQP_PORT:-5672}"; then
    echo "RabbitMQ connection ready."
  else
    echo "RabbitMQ connection failed."
    exit 1
  fi
fi

echo "All connections are ready."
exit 0

#!/bin/bash
set -e

mysql -h"${MAGENTO_DATABASE_HOST:-db}" -P"${MAGENTO_DATABASE_PORT:-3306}" -u"${MAGENTO_DATABASE_USER:-magento}" -p"${MAGENTO_DATABASE_PASSWORD:-magento}" -e "CREATE DATABASE IF NOT EXISTS ${MAGENTO_DATABASE_NAME:-magento}; "

if [ "${MAGENTO_ELASTICSEARCH_ENABLED:-false}" = "true" ]; then
  if [ "${MAGENTO_ELASTICSEARCH_ENABLE_AUTH:-0}" = "0" ]; then
    curl -X GET "http://${MAGENTO_ELASTICSEARCH_HOST:-opensearch}:${MAGENTO_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty"
  else
    curl -X GET "https://${MAGENTO_ELASTICSEARCH_HOST:-opensearch}:${MAGENTO_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty"
  fi
fi

if [ "${MAGENTO_REDIS_ENABLED:-false}" = "true" ]; then
  nc -v -z "${MAGENTO_REDIS_HOST:-redis}" "${MAGENTO_REDIS_PORT:-6379}"
fi

if [ "${MAGENTO_VARNISH_ENABLED:-false}" = "true" ]; then
  nc -v -z "${MAGENTO_VARNISH_HOST:-varnish}" "${MAGENTO_VARNISH_PORT:-80}"
fi

if [ "${MAGENTO_RABBITMQ_ENABLED:-false}" = "true" ]; then
  nc -v -z "${MAGENTO_AMQP_HOST:-rabbitmq}" "${MAGENTO_AMQP_PORT:-5672}"
fi

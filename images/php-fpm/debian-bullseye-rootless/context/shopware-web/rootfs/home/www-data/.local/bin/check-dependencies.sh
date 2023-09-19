#!/usr/bin/env bash
set -e

echo "Checking database connection..."
if mysql -h"${SHOPWARE_DATABASE_HOST:-db}" -P"${SHOPWARE_DATABASE_PORT:-3306}" -u"${SHOPWARE_DATABASE_USER:-app}" -p"${SHOPWARE_DATABASE_PASSWORD:-app}" -e "CREATE DATABASE IF NOT EXISTS ${SHOPWARE_DATABASE_NAME:-shopware}; "; then
  echo "Database connection ready."
else
  echo "Database connection failed."
  exit 1
fi

if [ "${SHOPWARE_ELASTICSEARCH_ENABLED:-false}" = "true" ]; then
  echo "Checking Elasticsearch connection..."
  if curl --connect-timeout 10 -fsSL -X GET "http://${SHOPWARE_ELASTICSEARCH_HOST:-opensearch}:${SHOPWARE_ELASTICSEARCH_PORT:-9200}/_cat/health?pretty"; then
    echo "Elasticsearch connection ready."
  else
    echo "Elasticsearch connection failed."
    exit 1
  fi
fi

echo "All connections are ready."
exit 0

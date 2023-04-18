#!/bin/bash
set -e

echo "Checking database connection..."
if mysql -h"${WORDPRESS_DATABASE_HOST:-db}" -P"${WORDPRESS_DATABASE_PORT:-3306}" -u"${WORDPRESS_DATABASE_NAME:-wordpress}" -p"${WORDPRESS_DATABASE_PASSWORD:-wordpress}" -e "CREATE DATABASE IF NOT EXISTS ${WORDPRESS_DATABASE_NAME:-wordpress}; "; then
  echo "Database connection ready."
else
  echo "Database connection failed."
  exit 1
fi

echo "All connections are ready."
exit 0

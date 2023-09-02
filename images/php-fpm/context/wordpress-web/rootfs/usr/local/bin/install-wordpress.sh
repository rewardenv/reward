#!/bin/bash
set -e

if [ -n "${COMMAND_BEFORE_INSTALL-}" ]; then eval "${COMMAND_BEFORE_INSTALL-}"; fi

if [ "${WORDPRESS_SKIP_BOOTSTRAP:-false}" = "true" ]; then
  if [ -n "${COMMAND_AFTER_INSTALL-}" ]; then eval "${COMMAND_AFTER_INSTALL-}"; fi
  exit
fi

if [ "${WORDPRESS_CONFIG:-true}" = "true" ]; then
  ARGS=()
  ARGS+=(
    "--force"
    "--dbhost=${WORDPRESS_DATABASE_HOST:-db}"
    "--dbname=${WORDPRESS_DATABASE_NAME:-wordpress}"
    "--dbuser=${WORDPRESS_DATABASE_USER:-wordpress}"
    "--dbpass=${WORDPRESS_DATABASE_PASSWORD:-wordpress}"
    "--dbprefix=${WORDPRESS_DATABASE_PREFIX:-wp_}"
    "--dbcharset=${WORDPRESS_DATABASE_CHARSET:-utf8}"
    "--dbcollate=${WORDPRESS_DATABASE_COLLATE:-}"
    "--locale=${WORDPRESS_LOCALE:-}"
  )

  wp core config "${ARGS[@]}" --extra-php <<PHP
${WORDPRESS_EXTRA_PHP}
PHP
fi

if [ "${WORDPRESS_SKIP_INSTALL:-false}" != "true" ]; then
  WORDPRESS_SCHEME="${WORDPRESS_SCHEME:-https}"
  WORDPRESS_HOST="${WORDPRESS_HOST:-wp.test}"
  WORDPRESS_URL="${WORDPRESS_URL:-"$WORDPRESS_SCHEME://$WORDPRESS_HOST"}"

  ARGS=()
  ARGS+=(
    "--url=${WORDPRESS_URL}"
    "--title=${WORDPRESS_BLOG_NAME:-wordpress}"
    "--admin_user=${WORDPRESS_USER:-admin}"
    "--admin_password=${WORDPRESS_PASSWORD:-ASDqwe12345}"
    "--admin_email=${WORDPRESS_EMAIL:-admin@example.com}"
  )

  wp core install "${ARGS[@]}"
fi

if [ "${WORDPRESS_DEPLOY_SAMPLE_DATA:-false}" = "true" ]; then
  wp plugin install --activate wordpress-importer

  curl -O https://raw.githubusercontent.com/manovotny/wptest/master/wptest.xml
  wp import wptest.xml --authors=create
  rm wptest.xml

  wp theme install twentytwentytwo --activate
fi

if [ -n "${COMMAND_AFTER_INSTALL-}" ]; then eval "${COMMAND_AFTER_INSTALL-}"; fi

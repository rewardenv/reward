#!/bin/bash
set -e

WORDPRESS_SHARED_CONFIG_PATH=${WORDPRESS_SHARED_CONFIG_PATH:-/config/wp-config.php}
if [ -f "${WORDPRESS_SHARED_CONFIG_PATH}" ]; then
  cp "${WORDPRESS_SHARED_CONFIG_PATH}" /var/www/html/wp-config.php
fi

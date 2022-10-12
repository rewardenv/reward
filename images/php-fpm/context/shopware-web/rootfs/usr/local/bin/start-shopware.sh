#!/bin/bash
set -e

SHOPWARE_SHARED_CONFIG_PATH=${SHOPWARE_SHARED_CONFIG_PATH:-/config/.env}
if [ -f "${SHOPWARE_SHARED_CONFIG_PATH}" ]; then
  cp "${SHOPWARE_SHARED_CONFIG_PATH}" /var/www/html/.env
fi

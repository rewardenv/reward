#!/bin/bash
set -e

MAGENTO_SHARED_CONFIG_PATH=${MAGENTO_SHARED_CONFIG_PATH:-/config/app/etc/env.php}
if [ -f "${MAGENTO_SHARED_CONFIG_PATH}" ]; then
  cp "${MAGENTO_SHARED_CONFIG_PATH}" /var/www/html/app/etc/env.php
fi

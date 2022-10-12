#!/bin/bash
set -e

if [ "${SHOPWARE_INSTALL:-false}" = "true" ]; then
  echo $'const:\n  APP_ENV: "dev"\n  APP_URL: "https://your-awesome-shopware-project.test"\n  DB_HOST: "db"\n  DB_NAME: "shopware"\n  DB_USER: "app"\n  DB_PASSWORD: "app"' > .psh.yaml.override
  psh.phar install
fi

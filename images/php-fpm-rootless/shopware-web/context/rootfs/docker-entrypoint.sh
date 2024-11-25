#!/bin/bash
set -e

version_gt() { test "$(printf '%s\n' "${@#v}" | sort -V | head -n 1)" != "${1#v}"; }

shopt -s expand_aliases
if [ -f "${HOME}/.bash_alias" ]; then
  source "${HOME}/.bash_alias"
fi

# Supervisor: Fix Permissions
if [ "${FIX_PERMISSIONS:-true}" = "true" ] && [ -f /etc/supervisor/available.d/permission.conf.template ]; then
  gomplate </etc/supervisor/available.d/permission.conf.template >/etc/supervisor/conf.d/permission.conf
fi

# Supervisor: Cron
if [ "${CRON_ENABLED:-false}" = "true" ] && [ -f /etc/supervisor/available.d/cron.conf.template ]; then
  gomplate </etc/supervisor/available.d/cron.conf.template >/etc/supervisor/conf.d/cron.conf
fi

# Supervisor: Socat
if [ "${SOCAT_ENABLED:-false}" = "true" ] &&
  [ -S /run/host-services/ssh-auth.sock ] &&
  [ "${SSH_AUTH_SOCK}" != "/run/host-services/ssh-auth.sock" ] &&
  [ -f /etc/supervisor/available.d/socat.conf.template ]; then
  gomplate </etc/supervisor/available.d/socat.conf.template >/etc/supervisor/conf.d/socat.conf
fi

# Supervisor: Nginx
if [ "${NGINX_ENABLED:-true}" = "true" ] && [ -f /etc/supervisor/available.d/nginx.conf.template ]; then
  gomplate </etc/supervisor/available.d/nginx.conf.template >/etc/supervisor/conf.d/nginx.conf
  find /etc/nginx -name '*.template' -exec sh -c 'gomplate <${1} > ${1%.*}' sh {} \;
fi

# Supervisor: PHP-FPM
if [ "${PHP_FPM_ENABLED:-true}" = "true" ] && [ -f /etc/supervisor/available.d/php-fpm.conf.template ]; then
  gomplate </etc/supervisor/available.d/php-fpm.conf.template >/etc/supervisor/conf.d/php-fpm.conf
fi

# Supervisor: Gotty
if [ "${GOTTY_ENABLED:-false}" = "true" ] && [ -f /etc/supervisor/available.d/gotty.conf.template ]; then
  gomplate </etc/supervisor/available.d/gotty.conf.template >/etc/supervisor/conf.d/gotty.conf
fi

# PHP
PHP_PREFIX="/etc/php"
PHP_PREFIX_LONG="${PHP_PREFIX}/${PHP_VERSION}"

# Configure PHP Global Settings
gomplate <"${PHP_PREFIX}/mods-available/docker.ini.template" >"${PHP_PREFIX_LONG}/mods-available/docker.ini"
phpenmod docker

# Configure PHP Opcache
gomplate <"${PHP_PREFIX}/mods-available/opcache.ini.template" >"${PHP_PREFIX_LONG}/mods-available/opcache.ini"
phpenmod opcache

# Configure PHP Cli
if [ -f "${PHP_PREFIX}/cli/conf.d/php-cli.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/cli/conf.d/php-cli.ini.template" >"${PHP_PREFIX_LONG}/cli/conf.d/php-cli.ini"
fi

# Configure PHP-FPM
if [ -f "${PHP_PREFIX}/fpm/conf.d/php-fpm.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/fpm/conf.d/php-fpm.ini.template" >"${PHP_PREFIX_LONG}/fpm/conf.d/php-fpm.ini"
fi

# Configure PHP-FPM Pool
if [ -f "${PHP_PREFIX}/fpm/pool.d/zz-docker.conf.template" ]; then
  gomplate <"${PHP_PREFIX}/fpm/pool.d/zz-docker.conf.template" >"${PHP_PREFIX_LONG}/fpm/pool.d/zz-docker.conf"
fi

# Update Reward Root Certificate if exist
if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
  cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
  update-ca-certificates
fi

if [ -f "${HOME}/msmtprc.template" ]; then
  gomplate <"${HOME}/msmtprc.template" >"${HOME}/.msmtprc"
  chmod 600 "${HOME}/.msmtprc"
fi

# Install requested node version if not already installed
NODE_INSTALLED="$(node -v | perl -pe 's/^v([0-9]+)\..*$/$1/')"
if [ "${NODE_INSTALLED}" -ne "${NODE_VERSION}" ] || [ "${NODE_VERSION}" = "latest" ] || [ "${NODE_VERSION}" = "lts" ]; then
  n install "${NODE_VERSION}"
fi

# Configure composer version
if [ "${COMPOSER_VERSION:-}" = "1" ]; then
  alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer1"
elif [ "${COMPOSER_VERSION:-}" = "2" ]; then
  alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
else
  if version_gt "${COMPOSER_VERSION:-}" "2.0"; then
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update "${COMPOSER_VERSION:-}"
  fi
fi

if [ "${CRON_ENABLED:-false}" = "true" ]; then
  printf "PATH=/home/www-data/.composer/vendor/bin:/home/www-data/bin:/home/www-data/.local/bin:/var/www/html/node_modules/.bin:/home/www-data/node_modules/.bin:/home/www-data/.local/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin\nSHELL=/bin/bash\n" |
    crontab -u www-data -

  # If CRONJOBS is set, write it to the crontab
  if [ -n "${CRONJOBS}" ]; then
    crontab -l -u www-data |
      {
        cat
        printf "%s\n" "${CRONJOBS}"
      } |
      crontab -u www-data -
  fi
fi

# If the first arg is `-D` or `--some-option` pass it to php-fpm.
if [ $# -eq 0 ] || [ "${1#-}" != "$1" ] || [ "${1#-}" != "$1" ]; then
  set -- supervisord -c /etc/supervisor/supervisord.conf "$@"
# If the first arg is supervisord call it normally.
elif [ "${1}" = "supervisord" ]; then
  set -- "$@"
# If the first arg is anything else
else
  set -- "$@"
fi

exec "$@"

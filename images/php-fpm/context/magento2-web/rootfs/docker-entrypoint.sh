#!/bin/bash
set -e

# Disable sudo for www-data if it's not explicitly configured to be enabled.
if [ "${SUDO_ENABLED}" != "true" ]; then
  echo >/etc/sudoers.d/nopasswd
  echo >/etc/sudoers.d/www-data
fi

# Supervisor: Fix Permissions
if [ "${FIX_PERMISSIONS:-true}" = "true" ]; then
  gomplate </etc/supervisor/available.d/permission.conf.template >/etc/supervisor/conf.d/permission.conf
fi

# Supervisor: Cron
if [ "${CRON_ENABLED:-false}" = "true" ]; then
  gomplate </etc/supervisor/available.d/cron.conf.template >/etc/supervisor/conf.d/cron.conf
fi

# Supervisor: Socat
if [ "${SOCAT_ENABLED:-false}" = "true" ] &&
  [ -S /run/host-services/ssh-auth.sock ] &&
  [ "${SSH_AUTH_SOCK}" != "/run/host-services/ssh-auth.sock" ]; then
  gomplate </etc/supervisor/available.d/socat.conf.template >/etc/supervisor/conf.d/socat.conf
fi

# Supervisor: Nginx
if [ "${NGINX_ENABLED:-true}" = "true" ]; then
  gomplate </etc/supervisor/available.d/nginx.conf.template >/etc/supervisor/conf.d/nginx.conf
  find /etc/nginx -name '*.template' -exec sh -c 'gomplate <${1} > ${1%.*}' sh {} \;
  ln -sf /proc/self/fd/1 /var/log/nginx/access.log && ln -sf /proc/self/fd/2 /var/log/nginx/error.log
fi

# Supervisor: PHP-FPM
gomplate </etc/supervisor/available.d/php-fpm.conf.template >/etc/supervisor/conf.d/php-fpm.conf

# Debian
if [ -x "$(command -v apt-get)" ]; then
  PHP_PREFIX="/etc/php/${PHP_VERSION}"

  # Configure PHP Global Settings
  gomplate <"${PHP_PREFIX}/mods-available/docker.ini.template" >"${PHP_PREFIX}/mods-available/docker.ini"
  phpenmod docker

  # Configure PHP Opcache
  gomplate <"${PHP_PREFIX}/mods-available/opcache.ini.template" >"${PHP_PREFIX}/mods-available/opcache.ini"
  phpenmod opcache

  # Configure PHP-FPM
  printf "[www]
user=%s
group=%s
;listen=%s
;listen.owner=%s
;listen.group=%s
" "${UID}" "${GID}" "${NGINX_UPSTREAM_PORT}" "${UID}" "${GID}" >"${PHP_PREFIX}/fpm/zzz-docker.conf"

  # Update Reward Root Certificate if exist
  if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
    cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
    update-ca-certificates
  fi

# CentOS
elif [ -x "$(command -v dnf)" ] || [ -x "$(command -v yum)" ]; then
  PHP_PREFIX="/etc/php.d"

  # Configure PHP Global Settings
  gomplate <"${PHP_PREFIX}/docker.ini.template" >"${PHP_PREFIX}/01-docker.ini"

  # Configure PHP Opcache
  gomplate <"${PHP_PREFIX}/opcache.ini.template" >"${PHP_PREFIX}/10-opcache.ini"

  # Configure PHP-FPM
  printf "[www]
user=%s
group=%s
listen=%s
;listen.owner=%s
;listen.group=%s
" "${UID}" "${GID}" "${NGINX_UPSTREAM_PORT}" "${UID}" "${GID}" >/etc/php-fpm.d/zzz-docker.conf

  # Update Reward Root Certificate if exist
  if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
    cp /etc/ssl/reward-rootca-cert/ca.cert.pem /etc/pki/ca-trust/source/anchors/reward-rootca-cert.pem
    update-ca-trust
  fi
fi

# Install requested node version if not already installed
NODE_INSTALLED="$(node -v | perl -pe 's/^v([0-9]+)\..*$/$1/')"
if [ "${NODE_INSTALLED}" -ne "${NODE_VERSION}" ] || [ "${NODE_VERSION}" = "latest" ] || [ "${NODE_VERSION}" = "lts" ]; then
  n "${NODE_VERSION}"
fi

# Configure composer version
if [ "${COMPOSER_VERSION:-}" = "1" ]; then
  alternatives --set composer /usr/bin/composer1
elif [ "${COMPOSER_VERSION:-}" = "2" ]; then
  alternatives --set composer /usr/bin/composer2
fi

# If command is not specified, run supervisord as root
if [ $# -eq 0 ]; then
  set -- supervisord -c /etc/supervisor/supervisord.conf
else
  # Drop privilege and run the called command as www-data
  set -- gosu www-data "$@"
fi

exec "$@"

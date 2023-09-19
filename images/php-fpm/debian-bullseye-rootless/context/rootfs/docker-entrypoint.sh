#!/usr/bin/env bash
set -e
shopt -s expand_aliases
if [ -f "${HOME}/.bash_alias" ]; then
  source "${HOME}/.bash_alias"
fi

# PHP
PHP_PREFIX="$HOME/.local/etc/php"
PHP_PREFIX_LONG="${PHP_PREFIX}/${PHP_VERSION}"

mkdir -p "${PHP_PREFIX_LONG}/mods-available" "${PHP_PREFIX_LONG}/cli/conf.d" "${PHP_PREFIX_LONG}/fpm/conf.d" "${PHP_PREFIX_LONG}/fpm/pool.d"

# Configure PHP Global Settings
if [ -f "${PHP_PREFIX}/mods-available/docker.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/mods-available/docker.ini.template" >"${PHP_PREFIX_LONG}/mods-available/docker.ini"
  phpenmod docker
fi

# Configure PHP Opcache
if [ -f "${PHP_PREFIX}/mods-available/opcache.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/mods-available/opcache.ini.template" >"${PHP_PREFIX_LONG}/mods-available/opcache.ini"
  phpenmod opcache
fi

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

# Configure PHP XDebug
if [ -f "${PHP_PREFIX}/mods-available/xdebug.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/mods-available/xdebug.ini.template" >"${PHP_PREFIX_LONG}/mods-available/xdebug.ini"
  phpenmod xdebug
fi

# Configure PHP Blackfire
if [ -f "${PHP_PREFIX}/mods-available/blackfire.ini.template" ]; then
  gomplate <"${PHP_PREFIX}/mods-available/blackfire.ini.template" >"${PHP_PREFIX_LONG}/mods-available/blackfire.ini"
  phpenmod blackfire
fi

# Update Reward Root Certificate if exist
if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
  cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
  update-ca-certificates
fi

# Start Cron
cron

# start socat process in background to connect sockets used for agent access within container environment
# shellcheck disable=SC2039
if [ -S /run/host-services/ssh-auth.sock ] && [ "${SSH_AUTH_SOCK}" != "/run/host-services/ssh-auth.sock" ]; then
  bash -c "nohup socat UNIX-CLIENT:/run/host-services/ssh-auth.sock \
    UNIX-LISTEN:${SSH_AUTH_SOCK},fork,user=www-data,group=www-data 1>/var/log/socat-ssh-auth.log 2>&1 &"
fi

# Install requested node version if not already installed
NODE_INSTALLED="$(node -v | perl -pe 's/^v([0-9]+)\..*$/$1/')"
if [ "${NODE_INSTALLED}" -ne "${NODE_VERSION}" ] || [ "${NODE_VERSION}" = "latest" ] || [ "${NODE_VERSION}" = "lts" ]; then
  n install "${NODE_VERSION}"
fi

# Configure composer version
if [ "${COMPOSER_VERSION:-}" = "1" ]; then
  alternatives --set composer "${HOME}/.local/bin/composer1"
elif [ "${COMPOSER_VERSION:-}" = "2" ]; then
  alternatives --set composer "${HOME}/.local/bin/composer2"
fi

# If the first arg is `-D` or `--some-option` pass it to php-fpm.
if [ "${1#-}" != "$1" ] || [ "${1#-}" != "$1" ]; then
  set -- php-fpm "$@"
# If the first arg is php-fpm call it normally.
else
  set -- "$@"
fi

exec "$@"

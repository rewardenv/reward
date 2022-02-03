#!/bin/bash
set -e

# Debian
if [ -x "$(command -v apt-get)" ]; then
  PHP_PREFIX="/etc/php/${PHP_VERSION}"

  # Configure PHP Global Settings
  gomplate <"${PHP_PREFIX}/mods-available/docker.ini.template" >"${PHP_PREFIX}/mods-available/docker.ini"
  phpenmod docker

  # Configure PHP Opcache
  gomplate <"${PHP_PREFIX}/mods-available/opcache.ini.template" >"${PHP_PREFIX}/mods-available/opcache.ini"
  phpenmod opcache

  # Configure PHP Cli
  if [ -f "${PHP_PREFIX}/cli/php-cli.ini.template" ]; then
    gomplate <"${PHP_PREFIX}/cli/php-cli.ini.template" >"${PHP_PREFIX}/cli/php-cli.ini"
  fi

  # Configure PHP XDebug
  if [ -f "${PHP_PREFIX}/mods-available/xdebug.ini.template" ]; then
    gomplate <"${PHP_PREFIX}/mods-available/xdebug.ini.template" >"${PHP_PREFIX}/mods-available/xdebug.ini"
    phpenmod xdebug
  fi

  # Configure PHP Blackfire
  if [ -f "${PHP_PREFIX}/mods-available/blackfire.ini.template" ]; then
    gomplate <"${PHP_PREFIX}/mods-available/blackfire.ini.template" >"${PHP_PREFIX}/mods-available/blackfire.ini"
    phpenmod blackfire
  fi

  # Update Reward Root Certificate if exist
  if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
    cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
    update-ca-certificates
  fi

  # Start Cron
  cron

# CentOS
elif [ -x "$(command -v dnf)" ] || [ -x "$(command -v yum)" ]; then
  PHP_PREFIX="/etc/php.d"

  # Configure PHP Global Settings
  gomplate <"${PHP_PREFIX}/docker.ini.template" >"${PHP_PREFIX}/01-docker.ini"

  # Configure PHP Opcache
  gomplate <"${PHP_PREFIX}/opcache.ini.template" >"${PHP_PREFIX}/10-opcache.ini"

  # Configure PHP Cli
  if [ -f "/etc/php-cli.ini.template" ]; then
    gomplate <"/etc/php-cli.ini.template" >"/etc/php-cli.ini"
  fi

  # Configure PHP XDebug
  if [ -f "${PHP_PREFIX}/xdebug.ini.template" ]; then
    gomplate <"${PHP_PREFIX}/xdebug.ini.template" >"${PHP_PREFIX}/15-xdebug.ini"
  fi

  # Configure PHP Blackfire
  if [ -f "${PHP_PREFIX}/blackfire.ini.template" ]; then
    gomplate <"${PHP_PREFIX}/blackfire.ini.template" >"${PHP_PREFIX}/10-blackfire.ini"
  fi

  # Update Reward Root Certificate if exist
  if [ -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]; then
    cp /etc/ssl/reward-rootca-cert/ca.cert.pem /etc/pki/ca-trust/source/anchors/reward-rootca-cert.pem
    update-ca-trust
  fi

  # Start Cron
  crond
fi

# start socat process in background to connect sockets used for agent access within container environment
# shellcheck disable=SC2039
if [ -S /run/host-services/ssh-auth.sock ] && [ "${SSH_AUTH_SOCK}" != "/run/host-services/ssh-auth.sock" ]; then
  bash -c "nohup socat UNIX-CLIENT:/run/host-services/ssh-auth.sock \
    UNIX-LISTEN:${SSH_AUTH_SOCK},fork,user=www-data,group=www-data 1>/var/log/socat-ssh-auth.log 2>&1 &"
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

# Resolve permission issues with directories auto-created by volume mounts; to use set CHOWN_DIR_LIST to
# a list of directories (relative to working directory) to chown, walking up the paths to also chown each
# specified parent directory. Example: "dir1/dir2 dir3" will chown dir1/dir2, then dir1 followed by dir3
# shellcheck disable=SC2039
for DIR in ${CHOWN_DIR_LIST:-}; do
  if [ -d "${DIR}" ]; then
    while :; do
      chown www-data:www-data "${DIR}"
      DIR=$(dirname "${DIR}")
      if [ "${DIR}" = "." ] || [ "${DIR}" = "/" ]; then
        break
      fi
    done
  fi
done

# Resolve permission issue with /var/www/html being owned by root as a result of volume mounted on php-fpm
# and nginx combined with nginx running as a different uid/gid than php-fpm does. This condition, when it
# surfaces would cause mutagen sync failures (on initial startup) on macOS environments.
chown www-data:www-data /var/www/html

# If the first arg is `-F` or `--some-option` pass it to php-fpm.
if [ "${1#-}" != "$1" ] || [ "${1#-}" != "$1" ]; then
  set -- php-fpm "$@"
# If the first arg is php-fpm call it normally.
elif [ "${1}" == "php-fpm" ]; then
  set -- "$@"
# If the first arg is anything else, drop privilege and run the called command as www-data.
else
  set -- gosu www-data "$@"
fi

exec "$@"

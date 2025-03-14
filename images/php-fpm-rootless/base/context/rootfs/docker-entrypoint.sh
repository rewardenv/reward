#!/bin/bash
set -e

shopt -s expand_aliases
if [[ -f "${HOME}/.bash_alias" ]]; then
  source "${HOME}/.bash_alias"
fi

configure_php() {
  local PHP_PREFIX="${PHP_PREFIX:-$HOME/.local/etc/php}"
  local PHP_PREFIX_LONG="${PHP_PREFIX}/${PHP_VERSION?required}"

  prepare_php_directories
  configure_php_settings
  configure_php_opcache
  configure_php_cli
  configure_php_fpm
  configure_php_fpm_pool
  configure_php_xdebug
  configure_php_blackfire
  configure_php_spx
}

prepare_php_directories() {
  mkdir -p "${PHP_PREFIX_LONG}/mods-available" "${PHP_PREFIX_LONG}/cli/conf.d" "${PHP_PREFIX_LONG}/fpm/conf.d" "${PHP_PREFIX_LONG}/fpm/pool.d"
}

configure_php_settings() {
  if [[ ! -f "${PHP_PREFIX}/mods-available/docker.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/mods-available/docker.ini.template" >"${PHP_PREFIX_LONG}/mods-available/docker.ini"
  phpenmod docker
}

configure_php_opcache() {
  if [[ ! -f "${PHP_PREFIX}/mods-available/opcache.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/mods-available/opcache.ini.template" >"${PHP_PREFIX_LONG}/mods-available/opcache.ini"
  phpenmod opcache
}

configure_php_cli() {
  if [[ ! -f "${PHP_PREFIX}/cli/conf.d/php-cli.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/cli/conf.d/php-cli.ini.template" >"${PHP_PREFIX_LONG}/cli/conf.d/php-cli.ini"
}

configure_php_fpm() {
  if [[ ! -f "${PHP_PREFIX}/fpm/conf.d/php-fpm.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/fpm/conf.d/php-fpm.ini.template" >"${PHP_PREFIX_LONG}/fpm/conf.d/php-fpm.ini"
}

configure_php_fpm_pool() {
  if [[ ! -f "${PHP_PREFIX}/fpm/pool.d/zz-docker.conf.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/fpm/pool.d/zz-docker.conf.template" >"${PHP_PREFIX_LONG}/fpm/pool.d/zz-docker.conf"
}

configure_php_xdebug() {
  if [[ ! -f "${PHP_PREFIX}/mods-available/xdebug.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/mods-available/xdebug.ini.template" >"${PHP_PREFIX_LONG}/mods-available/xdebug.ini"
  phpenmod xdebug
}

configure_php_blackfire() {
  if [[ ! -f "${PHP_PREFIX}/mods-available/blackfire.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/mods-available/blackfire.ini.template" >"${PHP_PREFIX_LONG}/mods-available/blackfire.ini"
  phpenmod blackfire
}

configure_php_spx() {
  if [[ ! -f "${PHP_PREFIX}/mods-available/spx.ini.template" ]]; then
    return 0
  fi

  gomplate <"${PHP_PREFIX}/mods-available/spx.ini.template" >"${PHP_PREFIX_LONG}/mods-available/spx.ini"
  phpenmod spx
}

configure_reward_root_certificate() {
  if [[ ! -f /etc/ssl/reward-rootca-cert/ca.cert.pem ]]; then
    return 0
  fi

  cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
  update-ca-certificates
}

configure_msmtp() {
  if [[ ! -f "${HOME}/msmtprc.template" ]]; then
    return 0
  fi

  gomplate <"${HOME}/msmtprc.template" >"${HOME}/.msmtprc"
  chmod 600 "${HOME}/.msmtprc"
}

configure_node_version() {
  NODE_INSTALLED="$(node -v | perl -pe 's/^v([0-9]+)\..*$/$1/')"
  if [[ "${NODE_INSTALLED}" == "${NODE_VERSION}" ]]; then
    return 0
  fi

  n install "${NODE_VERSION}"
}

configure_composer_version() {
  if [[ -z "${COMPOSER_VERSION:-}" ]]; then
    return 0
  fi

  case "${COMPOSER_VERSION:-}" in
  "1")
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer1"
    composer self-update --1
    ;;
  "2")
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update --2
    ;;
  "2.2")
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update --2.2
    ;;
  "stable")
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update --stable
    ;;
  *)
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update "${COMPOSER_VERSION:-}"
    ;;
  esac
}

start_socat() {
  # start socat process in background to connect sockets used for agent access within container environment
  # shellcheck disable=SC2039
  if [[ ! -S /run/host-services/ssh-auth.sock ]] || [[ "${SSH_AUTH_SOCK}" == "/run/host-services/ssh-auth.sock" ]]; then
    return 0
  fi

  bash -c "nohup socat UNIX-CLIENT:/run/host-services/ssh-auth.sock \
    UNIX-LISTEN:${SSH_AUTH_SOCK},fork,user=www-data,group=www-data 1>/var/log/socat-ssh-auth.log 2>&1 &"
}

configure_cron() {
  if [[ "${CRON_ENABLED:-false}" != "true" ]]; then
    return 0
  fi

  printf "PATH=/home/www-data/.composer/vendor/bin:/home/www-data/bin:/home/www-data/.local/bin:/var/www/html/node_modules/.bin:/home/www-data/node_modules/.bin:/home/www-data/.local/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin\nSHELL=/bin/bash\n" |
    crontab -u www-data -

  # If CRONJOBS is set, write it to the crontab
  if [[ -n "${CRONJOBS:-}" ]]; then
    crontab -l -u www-data |
      {
        cat
        printf "%s\n" "${CRONJOBS:-}"
      } |
      crontab -u www-data -
  fi
}

start_cron() {
  if [[ "${CRON_ENABLED:-false}" == "true" ]]; then
    cron
  fi
}

main() {
  configure_php
  configure_reward_root_certificate
  configure_msmtp
  configure_node_version
  configure_composer_version

  start_socat

  configure_cron
  start_cron

  # If the first arg is `-D` or `--some-option` pass it to php-fpm.
  if [[ "${1#-}" != "$1" ]] || [[ "${1#-}" != "$1" ]]; then
    set -- php-fpm "$@"
  # If the first arg is php-fpm call it normally.
  else
    set -- "$@"
  fi

  exec "$@"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

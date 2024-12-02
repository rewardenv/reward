#!/bin/bash
set -e

version_gt() { test "$(printf "%s\n" "${@#v}" | sort -V | head -n 1)" != "${1#v}"; }

shopt -s expand_aliases
if [[ -f "${HOME}/.bash_alias" ]]; then
  source "${HOME}/.bash_alias"
fi

configure_supervisord() {
  configure_supervisord_fix_permissions
  configure_supervisord_cron
  configure_supervisord_socat
  configure_supervisord_nginx
  configure_supervisord_php_fpm
  configure_supervisord_gotty
}

configure_supervisord_fix_permissions() {
  if [[ "${FIX_PERMISSIONS:-true}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/permission.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/permission.conf.template >/etc/supervisor/conf.d/permission.conf
}

configure_supervisord_cron() {
  if [[ "${CRON_ENABLED:-false}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/cron.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/cron.conf.template >/etc/supervisor/conf.d/cron.conf
}

configure_supervisord_socat() {
  if [[ "${SOCAT_ENABLED:-false}" != "true" ]] ||
    [[ ! -S /run/host-services/ssh-auth.sock ]] ||
    [[ "${SSH_AUTH_SOCK}" == "/run/host-services/ssh-auth.sock" ]] ||
    [[ ! -f /etc/supervisor/available.d/socat.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/socat.conf.template >/etc/supervisor/conf.d/socat.conf
}

configure_supervisord_nginx() {
  if [[ "${NGINX_ENABLED:-true}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/nginx.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/nginx.conf.template >/etc/supervisor/conf.d/nginx.conf
  find /etc/nginx -name '*.template' -exec sh -c 'gomplate <${1} > ${1%.*}' sh {} \;
}

configure_supervisord_php_fpm() {
  if [[ "${PHP_FPM_ENABLED:-true}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/php-fpm.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/php-fpm.conf.template >/etc/supervisor/conf.d/php-fpm.conf
}

configure_supervisord_gotty() {
  if [[ "${GOTTY_ENABLED:-false}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/gotty.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/gotty.conf.template >/etc/supervisor/conf.d/gotty.conf
}

configure_php() {
  local PHP_PREFIX="${PHP_PREFIX:-$HOME/.local/etc/php}"
  local PHP_PREFIX_LONG="${PHP_PREFIX}/${PHP_VERSION?required}"

  prepare_php_directories
  configure_php_settings
  configure_php_opcache
  configure_php_cli
  configure_php_fpm
  configure_php_fpm_pool
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
  if [[ "${COMPOSER_VERSION:-}" == "1" ]]; then
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer1"
    return $?
  fi

  if [[ "${COMPOSER_VERSION:-}" == "2" ]]; then
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    return $?
  fi

  if version_gt "${COMPOSER_VERSION:-}" "2.0"; then
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --set composer "${HOME}/.local/bin/composer2"
    composer self-update "${COMPOSER_VERSION:-}"
  fi
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

main() {
  configure_php
  configure_reward_root_certificate
  configure_msmtp
  configure_node_version
  configure_composer_version

  configure_cron

  configure_supervisord

  # If the first arg is `-D` or `--some-option` pass it to supervisord.
  if [[ $# -eq 0 ]] || [[ "${1#-}" != "$1" ]] || [[ "${1#-}" != "$1" ]]; then
    set -- supervisord -c /etc/supervisor/supervisord.conf "$@"
  # If the first arg is supervisord call it normally.
  elif [[ "${1}" == "supervisord" ]]; then
    set -- "$@"
  # If the first arg is anything else
  else
    set -- "$@"
  fi

  exec "$@"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

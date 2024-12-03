#!/bin/bash
set -e

shopt -s expand_aliases
if [[ -f "${HOME}/.bash_alias" ]]; then
  source "${HOME}/.bash_alias"
fi

configure_supervisord() {
  configure_supervisord_sudo
  configure_supervisord_fix_permissions
  configure_supervisord_cron
  configure_supervisord_socat
  configure_supervisord_nginx
  configure_supervisord_php_fpm
  configure_supervisord_gotty
}

configure_supervisord_sudo() {
  if [[ "${SET_SUDO:-true}" != "true" ]] || [[ ! -f /etc/supervisor/available.d/sudo.conf.template ]]; then
    return 0
  fi

  gomplate </etc/supervisor/available.d/sudo.conf.template >/etc/supervisor/conf.d/sudo.conf
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
  local PHP_PREFIX="${PHP_PREFIX:-/etc/php}"
  local PHP_PREFIX_LONG="${PHP_PREFIX}/${PHP_VERSION?required}"

  configure_php_settings
  configure_php_opcache
  configure_php_cli
  configure_php_fpm
  configure_php_fpm_pool
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

  sudo cp /etc/ssl/reward-rootca-cert/ca.cert.pem /usr/local/share/ca-certificates/reward-rootca-cert.pem
  sudo update-ca-certificates
}

configure_msmtp() {
  if [[ ! -f "/etc/msmtprc.template" ]]; then
    return 0
  fi

  gomplate </etc/msmtprc.template | tee /home/www-data/.msmtprc | sudo tee /etc/msmtprc >/dev/null
  sudo chmod 0600 /etc/msmtprc /home/www-data/.msmtprc
}

configure_node_version() {
  NODE_INSTALLED="$(node -v | perl -pe 's/^v([0-9]+)\..*$/$1/')"
  if [[ "${NODE_INSTALLED}" == "${NODE_VERSION}" ]]; then
    return 0
  fi

  sudo n install "${NODE_VERSION}"
}

configure_composer_version() {
  if [[ -z "${COMPOSER_VERSION:-}" ]]; then
    return 0
  fi

  case "${COMPOSER_VERSION:-}" in
  "1")
    sudo alternatives --set composer "/usr/local/bin/composer1"
    sudo composer self-update --1
    ;;
  "2")
    sudo alternatives --set composer "/usr/local/bin/composer2"
    sudo composer self-update --2
    ;;
  "2.2")
    sudo alternatives --set composer "/usr/local/bin/composer2"
    sudo composer self-update --2.2
    ;;
  "stable")
    sudo alternatives --set composer "/usr/local/bin/composer2"
    sudo composer self-update --stable
    ;;
  *)
    sudo alternatives --set composer "/usr/local/bin/composer2"
    sudo composer self-update "${COMPOSER_VERSION:-}"
    ;;
  esac
}

configure_cron() {
  if [[ "${CRON_ENABLED:-false}" != "true" ]]; then
    return 0
  fi

  printf "PATH=/home/www-data/bin:/home/www-data/.local/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin\nSHELL=/bin/bash\n" |
    crontab -u www-data -

  # If CRONJOBS is set, write it to the crontab
  if [[ -n "${CRONJOBS:-}" ]]; then
    crontab -l -u www-data |
      {
        cat
        printf "%s\n" "${CRONJOBS:-}"
      } |
      crontab -u www-data -
  else
    # If CRONJOBS is not set, set default Magento cron
    printf "* * * * * /usr/bin/test ! -e /var/www/html/var/.maintenance.flag -a ! -e /var/www/html/var/.cron-disable && cd /var/www/html && /usr/bin/php /var/www/html/bin/magento cron:run 2>&1 | grep -v 'Ran jobs by schedule' >> /var/www/html/var/log/magento.cron.log\n" |
      crontab -u www-data -
  fi
}

change_wwwdata_password() {
  if [[ -n "${WWWDATA_PASSWORD:-}" ]]; then
    echo "www-data:${WWWDATA_PASSWORD:-}" | sudo /usr/sbin/chpasswd
    unset WWWDATA_PASSWORD
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

  change_wwwdata_password

  # If the first arg is `-D` or `--some-option` pass it to supervisord.
  if [[ $# -eq 0 ]] || [[ "${1#-}" != "$1" ]] || [[ "${1#-}" != "$1" ]]; then
    set -- sudo supervisord -c /etc/supervisor/supervisord.conf "$@"
  # If the first arg is supervisord call it normally.
  elif [[ "${1}" == "supervisord" ]]; then
    set -- sudo "$@"
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

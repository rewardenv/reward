#!/bin/bash
[[ "${DEBUG:-false}" == "true" ]] && set -x
set -eEu -o pipefail -o errtrace
shopt -s extdebug

SCRIPT_DIR="$(dirname "$(realpath "${BASH_SOURCE[0]}")")"
FUNCTIONS_FILE="${SCRIPT_DIR}/functions.sh"
if [[ ! -f "${FUNCTIONS_FILE}" ]]; then
  FUNCTIONS_FILE="$(command -v functions.sh)"
fi
readonly FUNCTIONS_FILE

if [[ -f "${FUNCTIONS_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${FUNCTIONS_FILE}"
else
  printf "\033[1;31m%s ERROR: Required file %s not found\033[0m\n" "$(date --iso-8601=seconds)" "${FUNCTIONS_FILE}" >&2
  exit 1
fi

PHP_ARGS="-derror_reporting=${PHP_ERROR_REPORTING:-E_ALL} -dmemory_limit=${PHP_MEMORY_LIMIT:-2G}"

_console_command="bin/ci"
CONSOLE_COMMAND="${CONSOLE_COMMAND:-php ${PHP_ARGS} ${_console_command} --no-ansi --no-interaction}"
readonly CONSOLE_COMMAND
unset _console_command

_composer_command="composer"
if command -v composer &>/dev/null; then
  _composer_command="$(command -v composer 2>/dev/null)"
fi
COMPOSER_COMMAND="${COMPOSER_COMMAND:-php ${PHP_ARGS} ${_composer_command} --no-ansi --no-interaction}"
readonly COMPOSER_COMMAND
unset _composer_command

console() {
  ${CONSOLE_COMMAND} "$@"
}

composer() {
  ${COMPOSER_COMMAND} "$@"
}

check_requirements() {
  check_command "composer"
}

command_before_install() {
  if [[ -z "${COMMAND_BEFORE_INSTALL:-}" ]]; then
    return 0
  fi

  log "Executing custom command before installation"
  eval "${COMMAND_BEFORE_INSTALL:-}"
}

command_after_install() {
  if [[ -z "${COMMAND_AFTER_INSTALL:-}" ]]; then
    return 0
  fi

  log "Executing custom command after installation"
  eval "${COMMAND_AFTER_INSTALL:-}"
}

composer_configure() {
  if [[ -n "${COMPOSER_AUTH:-}" ]]; then
    # HACK: workaround for
    # https://github.com/composer/composer/issues/12084
    # shellcheck disable=SC2016
    log '$COMPOSER_AUTH is set, skipping Composer configuration'

    return 0
  fi

  log "Configuring Composer"

  if [[ -n "${GITHUB_USER:-}" ]] && [[ -n "${GITHUB_TOKEN:-}" ]]; then
    composer global config http-basic.github.com "${GITHUB_USER:-}" "${GITHUB_TOKEN:-}"
  fi

  if [[ -n "${BITBUCKET_PUBLIC_KEY:-}" ]] && [[ -n "${BITBUCKET_PRIVATE_KEY:-}" ]]; then
    composer global config bitbucket-oauth.bitbucket.org "${BITBUCKET_PUBLIC_KEY:-}" "${BITBUCKET_PRIVATE_KEY:-}"
  fi

  if [[ -n "${GITLAB_TOKEN:-}" ]]; then
    composer global config gitlab-token.gitlab.com "${GITLAB_TOKEN:-}"
  fi
}

bootstrap_check() {
  if [[ "${SHOPWARE_SKIP_BOOTSTRAP:-false}" != "true" ]] && [[ "${SKIP_BOOTSTRAP:-false}" != "true" ]]; then
    return 0
  fi

  log "Skipping Shopware bootstrap"
  command_after_install
  exit
}

shopware_version() {
  if [[ -n "${shopware_version:-}" ]]; then
    echo "${shopware_version}"
    return 0
  fi

  shopware_version=$(jq '.packages[] | select (.name == "shopware/core") | .version' -r <"$(app_path)/composer.lock")
  # shellcheck disable=SC2081,SC3010
  if [[ $shopware_version != v6* ]]; then
    shopware_version="v$(jq '.packages[] | select (.name == "shopware/core") | .extra."branch-alias"."dev-trunk"' -r <"$(app_path)/composer.lock")"
  fi

  echo "${shopware_version}"
}

shopware_deployment_helper() {
  # If deployment helper is available, run it
  if [[ -f "$(app_path)/vendor/bin/shopware-deployment-helper" ]]; then
    log "Executing Shopware deployment helper"
    exec "$(app_path)/vendor/bin/shopware-deployment-helper" run
  fi
}

shopware_is_installed() {
  if (version_gt "$(shopware_version)" "6.5.99"); then
    if ! console system:is-installed; then
      false
    fi
    return $?
  fi

  if ! console system:config:get shopware.installed &>/dev/null; then
    false
  fi
}

shopware_args_defaults() {
  SHOPWARE_HOST=${SHOPWARE_HOST:-'shopware.test'}
  SHOPWARE_SCHEME=${SHOPWARE_SCHEME:-'https'}

  SHOPWARE_APP_URL=${SHOPWARE_APP_URL:-"$SHOPWARE_SCHEME://$SHOPWARE_HOST"}

  ARGS+=(
    "--app-env=${SHOPWARE_APP_ENV:-prod}"
    "--app-url=${SHOPWARE_APP_URL}"
    "--database-url=mysql://${SHOPWARE_DATABASE_USER:-app}:${SHOPWARE_DATABASE_PASSWORD:-app}@${SHOPWARE_DATABASE_HOST:-db}:${SHOPWARE_DATABASE_PORT:-3306}/${SHOPWARE_DATABASE_NAME:-shopware}"
    "--cdn-strategy=${SHOPWARE_CDN_STRATEGY:-physical_filename}"
    "--mailer-url=${SHOPWARE_MAILER_URL:-native://default}"
  )
}

shopware_args_elasticsearch() {
  # Configure Elasticsearch
  if [[ "${SHOPWARE_ELASTICSEARCH_ENABLED:-false}" != "true" ]]; then
    return 0
  fi
  local SHOPWARE_ELASTICSEARCH_HOSTS="${SHOPWARE_ELASTICSEARCH_HOSTS:-${SHOPWARE_ELASTICSEARCH_HOST:-elasticsearch}:${SHOPWARE_ELASTICSEARCH_PORT:-9200}}"

  ARGS+=(
    "--es-enabled=1"
    "--es-hosts=${SHOPWARE_ELASTICSEARCH_HOSTS}"
  )

  if [[ "${SHOPWARE_ELASTICSEARCH_INDEXING_ENABLED:-true}" == "true" ]]; then
    ARGS+=(
      "--es-indexing-enabled=1"
    )
  fi
}

shopware_args_opensearch() {
  # Configure Opensearch
  if [[ "${SHOPWARE_OPENSEARCH_ENABLED:-true}" != "true" ]]; then
    return 0
  fi

  local SHOPWARE_OPENSEARCH_HOSTS="${SHOPWARE_OPENSEARCH_HOSTS:-${SHOPWARE_OPENSEARCH_HOST:-opensearch}:${SHOPWARE_OPENSEARCH_PORT:-9200}}"

  ARGS+=(
    "--es-enabled=1"
    "--es-hosts=${SHOPWARE_OPENSEARCH_HOSTS}"
  )

  if [[ "${SHOPWARE_OPENSEARCH_INDEXING_ENABLED:-true}" == "true" ]]; then
    ARGS+=(
      "--es-indexing-enabled=1"
    )
  fi
}

shopware_args_extra() {
  if [[ -z "${SHOPWARE_EXTRA_INSTALL_ARGS:-}" ]]; then
    return 0
  fi

  ARGS+=(
    "${SHOPWARE_EXTRA_INSTALL_ARGS:-}"
  )
}

shopware_env_puppeteer_skip_chromium_download() {
  if [[ "${SHOPWARE_PUPPETEER_SKIP_CHROMIUM_DOWNLOAD:-true}" != "true" ]]; then
    return 0
  fi

  export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1
}

shopware_env_ci() {
  if [[ "${SHOPWARE_CI:-true}" == "true" ]]; then
    export CI=1
  else
    export CI=0
  fi
}

shopware_env_skip_bundle_dump() {
  if [[ "${SHOPWARE_SKIP_BUNDLE_DUMP:-false}" == "true" ]]; then
    export SHOPWARE_SKIP_BUNDLE_DUMP=1
  else
    export SHOPWARE_SKIP_BUNDLE_DUMP=0
  fi
}

shopware_env_disable_admin_compilation_typecheck() {
  if [[ "${SHOPWARE_DISABLE_ADMIN_COMPILATION_TYPECHECK:-true}" == "true" ]]; then
    export DISABLE_ADMIN_COMPILATION_TYPECHECK=1
  fi
}

shopware_configure_lock_dsn() {
  if grep -q "LOCK_DSN" "$(app_path)/.env"; then
    return 0
  fi

  echo "LOCK_DSN=${SHOPWARE_LOCK_DSN:-flock://var/lock}" >>"$(app_path)/.env"
}

shopware_maintenance_enable() {
  console sales-channel:maintenance:enable --all
}

shopware_maintenance_disable() {
  console sales-channel:maintenance:disable --all
}

shopware_bundle_dump() {
  console bundle:dump
  console theme:dump
}

shopware_skip_asset_build_flag() {
  if [[ "${SHOPWARE_SKIP_ASSET_COPY:-false}" == "true" ]]; then
    echo "--skip-asset-build"
  fi
}

shopware_list_plugins_not_installed() {
  console plugin:list --json | jq 'map(select(.installedAt == null)) | .[].name' -r
}

shopware_install_all_plugins() {
  log "Installing all plugins"
  list_with_updates=$(shopware_list_plugins_not_installed)

  for plugin in $list_with_updates; do
    console plugin:install --activate "$plugin"
  done
}

shopware_list_plugins_with_updates() {
  console plugin:list --json | jq 'map(select(.upgradeVersion != null)) | .[].name' -r
}

shopware_update_all_plugins() {
  log "Updating plugins"
  if (version_gt "$(shopware_version)" "6.5.99"); then
    if [[ -n "$(shopware_skip_asset_build_flag)" ]]; then
      console plugin:update:all "$(shopware_skip_asset_build_flag)"
      return $?
    fi

    console plugin:update:all
    return $?
  fi

  list_with_updates="$(shopware_list_plugins_with_updates)"

  for plugin in $list_with_updates; do
    console plugin:update "$plugin"
  done
}

shopware_configure() {
  if [[ "${SHOPWARE_SKIP_INSTALL:-false}" == "true" ]]; then
    return 0
  fi

  log "Installing Shopware"

  local ARGS=("")

  shopware_args_defaults
  shopware_args_elasticsearch
  shopware_args_opensearch
  shopware_args_extra
  shopware_env_puppeteer_skip_chromium_download
  shopware_env_ci
  shopware_env_disable_admin_compilation_typecheck
  shopware_env_skip_bundle_dump

  # shellcheck disable=SC2068
  console system:setup --force ${ARGS[@]}

  shopware_configure_lock_dsn
}

shopware_install() {
  log "Installing Shopware"
  console system:install --force --create-database --basic-setup --shop-locale="${SHOPWARE_LOCALE:-en-GB}" --shop-currency="${SHOPWARE_CURRENCY:-EUR}"
}

shopware_theme_change() {
  log "Changing theme to Storefront"
  console theme:change --all Storefront
}

shopware_system_update_finish() {
  log "Running shopware system:update:finish"
  if [[ -n "$(shopware_skip_asset_build_flag)" ]]; then
    console system:update:finish "$(shopware_skip_asset_build_flag)"
    return $?
  fi

  console system:update:finish
}

shopware_plugin_refresh() {
  if ! (version_gt "$(shopware_version)" "6.5.99"); then
    log "Refreshing plugins"
    console plugin:refresh
  fi
}

shopware_scheduled_task_register() {
  console scheduled-task:register
}

shopware_theme_refresh() {
  console theme:refresh
}

shopware_system_config_set() {
  log "Setting system configuration"
  console system:config:set core.frw.completedAt '2019-10-07T10:46:23+00:00'
  if ! (version_gt "$(shopware_version)" "6.5.99"); then
    console system:config:set core.usageData.shareUsageData false --json
  fi
}

shopware_admin_user_exists() {
  # Below Shopware 6.6.0 cannot list users via console
  if (! version_gt "$(shopware_version)" "6.6"); then
    return 0
  fi

  # If console user:list command is not available, return 0
  if ! console user:list --json 2>/dev/null; then
    return 0
  fi

  if console user:list --json 2>/dev/null | jq -e ".[] | select(.username == \"${SHOPWARE_USERNAME:-admin}\")" >/dev/null; then
    return 0
  fi

  false
}

shopware_admin_user() {
  if shopware_admin_user_exists; then
    log "Admin user already exists, updating password"
    if (! version_gt "$(shopware_version)" "6.6"); then
      log "Below Shopware 6.6.0, admin user cannot be queried, so changing password without checking if user exists"
      console user:change-password "${SHOPWARE_USERNAME:-admin}" --password="${SHOPWARE_PASSWORD:-ASDqwe123}" || true
      return $?
    fi

    console user:change-password "${SHOPWARE_USERNAME:-admin}" --password="${SHOPWARE_PASSWORD:-ASDqwe123}"
    return $?
  fi

  log "Creating admin user"
  declare -a ARGS
  ARGS=(
    "${SHOPWARE_USERNAME:-admin}"
    "--admin"
    "--firstName=${SHOPWARE_FIRST_NAME:-admin}"
    "--lastName=${SHOPWARE_LAST_NAME:-admin}"
    "--email=${SHOPWARE_EMAIL:-admin@example.com}"
    "--password=${SHOPWARE_PASSWORD:-ASDqwe123}"
  )

  # shellcheck disable=SC2068
  console user:create "${ARGS[@]}" >/dev/null
}

shopware_disable_deploy_sample_data() {
  log "Disabling deploy sample data"
  export SHOPWARE_DEPLOY_SAMPLE_DATA=false
}

shopware_deploy_sample_data() {
  if [[ "${SHOPWARE_DEPLOY_SAMPLE_DATA:-false}" != "true" ]]; then
    return 0
  fi

  log "Deploying sample data"
  mkdir -p "$(app_path)/custom/plugins"
  APP_ENV="${SHOPWARE_APP_ENV:-prod}" console store:download -p SwagPlatformDemoData
  console plugin:install SwagPlatformDemoData --activate
  shopware_cache_clear
  shopware_cache_warmup
}

shopware_cache_clear() {
  log "Clearing cache"
  console cache:clear --no-warmup
}

shopware_cache_warmup() {
  log "Warming up cache"
  console cache:warmup
  console http:cache:warm:up
}

shopware_publish_config() {
  log "Publishing config"
  cp -a "$(app_path)/.env" "$(shared_config_path)/.env"
}

main() {
  conditional_sleep

  LOCKFILE="$(shared_config_path)/.deploy.lock"
  readonly LOCKFILE

  trap 'lock_cleanup ${LOCKFILE}' EXIT
  trap 'trapinfo $LINENO ${BASH_LINENO[*]}' ERR

  lock_acquire "${LOCKFILE}"

  run_hooks "pre-install"

  check_requirements

  command_before_install
  bootstrap_check
  composer_configure

  shopware_deployment_helper
  local shopware_version
  shopware_configure

  if shopware_is_installed; then
    shopware_maintenance_enable
    shopware_system_update_finish
    shopware_plugin_refresh
    shopware_update_all_plugins
    shopware_disable_deploy_sample_data
    shopware_maintenance_disable
  else
    shopware_install
    shopware_theme_change
    shopware_system_config_set
  fi

  shopware_install_all_plugins

  shopware_admin_user
  shopware_deploy_sample_data
  shopware_publish_config

  command_after_install

  run_hooks "post-install"
}

(return 0 2>/dev/null) && sourced=1

if [[ -z "${sourced:-}" ]]; then
  main "$@"
fi

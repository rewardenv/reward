#!/bin/bash

function setup() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/install.sh"
}

function test_wordpress_configure_default() {
  # Default
  setup

  mock wp true
  spy wp
  wordpress_configure
  assert_have_been_called_with "core config --force --dbhost=db --dbname=wordpress --dbuser=wordpress --dbpass=wordpress --dbprefix=wp_ --dbcharset=utf8" wp
}

function test_wordpress_configure_skip() {
  # Skip configure
  local WORDPRESS_CONFIG="false"
  setup

  spy wp
  wordpress_configure
  assert_have_been_called_times 0 wp
}

function test_wordpress_configure_custom() {
  # Custom values
  local WORDPRESS_DATABASE_HOST="localhost"
  local WORDPRESS_DATABASE_NAME="wp"
  local WORDPRESS_DATABASE_USER="root"
  local WORDPRESS_DATABASE_PASSWORD="rootpw"
  local WORDPRESS_DATABASE_PREFIX="wp"
  local WORDPRESS_DATABASE_CHARSET="utf8mb4"
  local WORDPRESS_DATABASE_COLLATE="utf8mb4_unicode_ci"
  local WORDPRESS_LOCALE="en_US"
  local WORDPRESS_EXTRA_PHP="define( 'WP_DEBUG', true );"
  setup

  spy wp
  wordpress_configure
  assert_have_been_called_with "core config --force --dbhost=localhost --dbname=wp --dbuser=root --dbpass=rootpw --dbprefix=wp --dbcharset=utf8mb4 --dbcollate=utf8mb4_unicode_ci --locale=en_US --extra-php" wp

  mock wp false
  assert_exit_code 1 "$(wordpress_configure)"
}

function test_wordpress_install_default() {
  # Default
  setup

  mock wp true
  spy wp
  wordpress_install
  assert_have_been_called_with "core install --url=https://wp.test --title=wordpress --admin_user=admin --admin_password=ASDqwe12345 --admin_email=admin@example.com" wp
}

function test_wordpress_install_skip() {
  # Skip install
  local WORDPRESS_SKIP_INSTALL="true"
  setup

  spy wp
  wordpress_install
  assert_have_been_called_times 0 wp
}

function test_wordpress_install_custom() {
  # Custom values
  local WORDPRESS_SCHEME="http"
  local WORDPRESS_HOST="example.com"
  local WORDPRESS_USER="johndoe"
  local WORDPRESS_PASSWORD="johndoepw"
  local WORDPRESS_EMAIL="johndoe@example.com"
  setup

  spy wp
  wordpress_install
  assert_have_been_called_with "core install --url=http://example.com --title=wordpress --admin_user=johndoe --admin_password=johndoepw --admin_email=johndoe@example.com" wp
}

function test_wordpress_publish_shared_files_default() {
  setup

  # Test with a valid SHARED_CONFIG_PATH
  local SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/wp-config.php"
  wordpress_publish_shared_files
  assert_file_exists "test-data/config/wp-config.php"
  rm -fr "./test-data"
  unset SHARED_CONFIG_PATH
}

function test_wordpress_publish_shared_files_not_writable() {
  setup

  # Test if SHARED_CONFIG_PATH is not writable (/config by default)
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/wp-config.php"
  wordpress_publish_shared_files
  assert_file_exists "/tmp/wp-config.php"
  rm -fr "/tmp/wp-config.php"
  rm -fr "./test-data"
}

function test_command_before_install_default() {
  # Default
  setup

  spy eval
  assert_exit_code 0 "$(command_before_install)"
}

function test_command_before_install_custom() {
  # Custom command
  local COMMAND_BEFORE_INSTALL="echo 'test'"
  setup

  spy eval
  command_before_install
  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_install_default() {
  # Default
  setup
  spy eval

  assert_exit_code 0 "$(command_before_install)"
}

function test_command_after_install_custom() {
  # Custom command
  local COMMAND_AFTER_INSTALL="echo 'test'"
  setup

  spy eval
  command_after_install
  assert_have_been_called_with "echo 'test'" eval
}

function test_bootstrap_check_default() {
  setup
  # If both are false it should not call the command_after_install
  assert_empty "$(bootstrap_check)"

}

function test_bootstrap_check_skip_bootstrap_but_command_after_install_set() {
  # If WORDPRESS_SKIP_BOOTSTRAP is true, it should just run the COMMAND_AFTER_INSTALL and exit
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local WORDPRESS_SKIP_BOOTSTRAP="true"
  setup

  assert_contains "test-123" "$(bootstrap_check)"

  # If SKIP_BOOTSTRAP is true, it should run the COMMAND_AFTER_INSTALL and exit
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
}

function test_bootstrap_check_both_enabled() {
  # If both are true it should run the COMMAND_AFTER_INSTALL
  local COMMAND_AFTER_INSTALL="echo 'test-123'"
  local WORDPRESS_SKIP_BOOTSTRAP="true"
  local SKIP_BOOTSTRAP="true"
  setup

  assert_contains "test-123" "$(bootstrap_check)"
}

function test_composer_configure_default() {
  # Default
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer
}

function test_composer_configure_github() {
  local GITHUB_USER="user"
  local GITHUB_TOKEN="token"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_bitbucket() {
  local BITBUCKET_PUBLIC_KEY="public"
  local BITBUCKET_PRIVATE_KEY="private"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_gitlab() {
  local GITLAB_TOKEN="token"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 1 composer
}

function test_composer_configure_auth() {
  local COMPOSER_AUTH="test"
  setup

  spy composer
  composer_configure
  assert_have_been_called_times 0 composer
}

function test_wordpress_is_installed_true() {
  setup

  mock wp true
  assert_exit_code 0 "$(wordpress_is_installed)"
}

function test_wordpress_is_installed_false() {
  setup

  mock wp false
  assert_exit_code 1 "$(wordpress_is_installed)"
}

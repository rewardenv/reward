#!/bin/bash

function set_up() {
  source "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/install.sh"
}

#function test_magento_admin_user_exists() {
#  # Valid admin
#  mock magerun <<EOF
#iid,username,email,status
#1,admin,admin@example.com,active
#21,admintest,admintest@example.com,active
#39,otheradmin,otheradmin@example.com,active
#39,user,user@example.com,active
#EOF
#  assert_exit_code 0 "$(magento_admin_user_exists)"
#
#  # Invalid admin
#  mock magerun <<EOF
#iid,username,email,status
#21,admintest,admintest@example.com,active
#39,otheradmin,otheradmin@example.com,active
#39,user,user@example.com,active
#EOF
#  assert_exit_code 1 "$(magento_admin_user_exists)"
#
#  # Custom admin username
#  local MAGENTO_USERNAME="johndoe"
#  mock magerun <<EOF
#iid,username,email,status
#5,johndoe,johndoe@example.com,inactive
#21,admintest,admintest@example.com,active
#39,otheradmin,otheradmin@example.com,active
#39,user,user@example.com,active
#EOF
#  assert_exit_code 0 "$(magento_admin_user_exists)"
#  unset MAGENTO_USERNAME
#
#  # Malformed output
#  mock magerun <<EOF
#iid,username,email,status
#admin,active
#admintest,active
#EOF
#  assert_exit_code 1 "$(magento_admin_user_exists)"
#}
#
#function test_magento_admin_user_inactive() {
#  # Admin inactive
#  mock magerun <<EOF
#iid,username,email,status
#1,admin,admin@example.com,inactive
#21,admintest,admintest@example.com,active
#39,otheradmin,otheradmin@example.com,active
#39,user,user@example.com,active
#EOF
#  assert_exit_code 0 "$(magento_admin_user_inactive)"
#
#  # Admin inactive custom admin username
#  local MAGENTO_USERNAME="johndoe"
#  mock magerun <<EOF
#iid,username,email,status
#5,johndoe,johndoe@example.com,inactive
#21,admintest,admintest@example.com,active
#39,otheradmin,otheradmin@example.com,active
#39,user,user@example.com,active
#EOF
#  assert_exit_code 0 "$(magento_admin_user_inactive)"
#  unset MAGENTO_USERNAME
#
#  # Admin user active
#  mock magerun <<EOF
#  iid,username,email,status
#  1,admin,admin@example.com,active
#  21,admintest,admintest@example.com,active
#  39,otheradmin,otheradmin@example.com,active
#  39,user,user@example.com,active
#EOF
#  assert_exit_code 1 "$(magento_admin_user_inactive)"
#
#  # Malformed output inactive
#  mock magerun <<EOF
#iid,username,email,status
#admin,inactive
#admintest,active
#EOF
#  assert_exit_code 1 "$(magento_admin_user_inactive)"
#
#  # Malformed output active
#  mock magerun <<EOF
#iid,username,email,status
#admin,active
#admintest,active
#EOF
#  assert_exit_code 1 "$(magento_admin_user_inactive)"
#}
#
#function test_magento_admin_user() {
#  # Admin user exist and active
#  mock magento_admin_user_exists true
#  spy magerun
#  magento_admin_user
#  assert_have_been_called_times 1 magerun
#
#  # Admin user exist and inactive
#  mock magento_admin_user_exists true
#  mock magento_admin_user_inactive true
#  spy magerun
#  magento_admin_user
#  assert_have_been_called_times 2 magerun
#
#  # Admin user doesn't exist
#  mock magento_admin_user_exists false
#  spy magerun
#  magento_admin_user
#  assert_have_been_called_with "admin:user:create --admin-firstname=admin --admin-lastname=admin --admin-email=admin@example.com --admin-user=admin --admin-password=ASDqwe123" magerun
#
#  # Custom admin user values
#  local MAGENTO_FIRST_NAME="John"
#  local MAGENTO_LAST_NAME="Doe"
#  local MAGENTO_EMAIL="johndoe@example.com"
#  local MAGENTO_USERNAME="johndoe"
#  local MAGENTO_PASSWORD="johndoepw"
#  mock magento_admin_user_exists false
#  spy magerun
#  magento_admin_user
#  assert_have_been_called_with "admin:user:create --admin-firstname=John --admin-lastname=Doe --admin-email=johndoe@example.com --admin-user=johndoe --admin-password=johndoepw" magerun
#}

function test_wordpress_configure() {
  # Skip configure
  local WORDPRESS_CONFIG="false"
  spy wp
  wordpress_configure
  assert_have_been_called_times 0 wp
  unset WORDPRESS_CONFIG

  # Default
  mock wp true
  spy wp
  wordpress_configure
  assert_have_been_called_with "core config --force --dbhost=db --dbname=wordpress --dbuser=wordpress --dbpass=wordpress --dbprefix=wp_ --dbcharset=utf8" wp

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
  spy wp
  wordpress_configure
  assert_have_been_called_with "core config --force --dbhost=localhost --dbname=wp --dbuser=root --dbpass=rootpw --dbprefix=wp --dbcharset=utf8mb4 --dbcollate=utf8mb4_unicode_ci --locale=en_US --extra-php" wp

  mock wp false
  assert_exit_code 1 "$(wordpress_configure)"
}

function test_wordpress_install() {
  # Skip install
  local WORDPRESS_SKIP_INSTALL="true"
  spy wp
  wordpress_install
  assert_have_been_called_times 0 wp
  unset WORDPRESS_SKIP_INSTALL

  # Default
  mock wp true
  spy wp
  wordpress_install
  assert_have_been_called_with "core install --url=https://wp.test --title=wordpress --admin_user=admin --admin_password=ASDqwe12345 --admin_email=admin@example.com" wp

  # Custom values
  local WORDPRESS_SCHEME="http"
  local WORDPRESS_HOST="example.com"
  local WORDPRESS_USER="johndoe"
  local WORDPRESS_PASSWORD="johndoepw"
  local WORDPRESS_EMAIL="johndoe@example.com"
  spy wp
  wordpress_install
  assert_have_been_called_with "core install --url=http://example.com --title=wordpress --admin_user=johndoe --admin_password=johndoepw --admin_email=johndoe@example.com" wp
}

function test_wordpress_publish_config() {
  # Test with a valid SHARED_CONFIG_PATH
  local SHARED_CONFIG_PATH="./test-data/config"
  mkdir -p "${SHARED_CONFIG_PATH}"
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/wp-config.php"
  wordpress_publish_config
  assert_file_exists "test-data/config/wp-config.php"
  rm -fr "./test-data"
  unset SHARED_CONFIG_PATH

  # Test if SHARED_CONFIG_PATH is not writable (/config by default)
  local APP_PATH="./test-data/var/www/html"
  mkdir -p "${APP_PATH}"
  touch "${APP_PATH}/wp-config.php"
  wordpress_publish_config
  assert_file_exists "/tmp/wp-config.php"
  rm -fr "/tmp/wp-config.php"
  rm -fr "./test-data"
}

function test_command_before_install() {
  # Default
  assert_exit_code 0 "$(command_before_install)"

  # Custom command
  local COMMAND_BEFORE_INSTALL="echo 'test'"
  spy eval

  command_before_install

  assert_have_been_called_with "echo 'test'" eval
}

function test_command_after_install() {
  # Default
  assert_exit_code 0 "$(command_before_install)"

  # Custom command
  local COMMAND_AFTER_INSTALL="echo 'test'"
  spy eval

  command_after_install

  assert_have_been_called_with "echo 'test'" eval
}

function test_bootstrap_check() {
  local COMMAND_AFTER_INSTALL="echo 'test-123'"

  # If WORDPRESS_SKIP_BOOTSTRAP is true, it should just run the COMMAND_AFTER_INSTALL and exit
  local WORDPRESS_SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset WORDPRESS_SKIP_BOOTSTRAP

  # If SKIP_BOOTSTRAP is true, it should run the COMMAND_AFTER_INSTALL and exit
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset SKIP_BOOTSTRAP

  # If both are true it should run the COMMAND_AFTER_INSTALL
  local WORDPRESS_SKIP_BOOTSTRAP="true"
  local SKIP_BOOTSTRAP="true"
  assert_contains "test-123" "$(bootstrap_check)"
  unset SKIP_BOOTSTRAP
  unset WORDPRESS_SKIP_BOOTSTRAP

  # If both are false it should not call the command_after_install
  assert_empty "$(bootstrap_check)"
}

function test_composer_configure() {
  # Default
  mock composer echo
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  # Test if only GITHUB_USER is set
  local GITHUB_USER="user"
  spy composer
  composer_configure
  assert_have_been_called_times 0 composer

  # Test if only GITHUB_TOKEN is set
  local GITHUB_USER="user"
  local GITHUB_TOKEN="token"
  spy composer
  composer_configure
  assert_have_been_called_times 1 composer

  local BITBUCKET_PUBLIC_KEY="public"
  local BITBUCKET_PRIVATE_KEY="private"
  spy composer
  composer_configure
  assert_have_been_called_times 2 composer

  local GITLAB_TOKEN="token"
  spy composer
  composer_configure
  assert_have_been_called_times 3 composer
}

function test_wordpress_is_installed() {
  mock wp true
  assert_exit_code 0 "$(wordpress_is_installed)"

  mock wp false
  assert_exit_code 1 "$(wordpress_is_installed)"
}

<?php
define( 'WP_HOME',          'https://{{.traefik_domain}}/' );
define( 'WP_SITEURL',       'https://{{.traefik_domain}}/' );
define( 'DB_NAME',          '{{ default "wordpress" .wordpress_db_name }}' );
define( 'DB_USER',          '{{ default "wordpress" .wordpress_db_user }}' );
define( 'DB_PASSWORD',      '{{ default "wordpress" .wordpress_db_password }}' );
define( 'DB_HOST',          '{{ default "db" .wordpress_db_host }}' );
define( 'DB_CHARSET',       '{{ default "utf8" .wordpress_db_charset }}' );
define( 'DB_COLLATE',       '{{ default "" .wordpress_db_collate }}' );
define( 'AUTH_KEY',         '{{ default "put your unique phrase here" .wordpress_auth_key }}' );
define( 'SECURE_AUTH_KEY',  '{{ default "put your unique phrase here" .wordpress_secure_auth_key }}' );
define( 'LOGGED_IN_KEY',    '{{ default "put your unique phrase here" .wordpress_logged_in_key }}' );
define( 'NONCE_KEY',        '{{ default "put your unique phrase here" .wordpress_nonce_key }}' );
define( 'AUTH_SALT',        '{{ default "put your unique phrase here" .wordpress_auth_salt }}' );
define( 'SECURE_AUTH_SALT', '{{ default "put your unique phrase here" .wordpress_secure_auth_key }}' );
define( 'LOGGED_IN_SALT',   '{{ default "put your unique phrase here" .wordpress_logged_in_salt }}' );
define( 'NONCE_SALT',       '{{ default "put your unique phrase here" .wordpress_nonce_salt }}' );
$table_prefix = '{{ default "wp_" .wordpress_table_prefix }}';
define( 'WP_DEBUG', {{ default "false" .wordpress_debug }} );
if ( ! defined( 'ABSPATH' ) ) {
	define( 'ABSPATH', __DIR__ . '/' );
}
require_once ABSPATH . 'wp-settings.php';

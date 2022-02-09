#!/bin/bash
set -e

cat <EOF> /var/www/html/wp-config.php
<?php
define( 'WP_HOME',          'https://${WORDPRESS_HOME:-TODO}/' );
define( 'WP_SITEURL',       'https://${WORDPRESS_SITEURL:-TODO}/' );
define( 'DB_NAME',          '${WORDPRESS_DB_NAME:-wordpress}"' );
define( 'DB_USER',          '${WORDPRESS_DB_USER:-wordpress}' );
define( 'DB_PASSWORD',      '${WORDPRESS_DB_PASSWORD:-wordpress}' );
define( 'DB_HOST',          '${WORDPRESS_DB_USER:-mysql}' );
define( 'DB_CHARSET',       '${WORDPRESS_DB_CHARSET:-utf8}' );
define( 'DB_COLLATE',       '${WORDPRESS_DB_COLLATE:-}' );
define( 'AUTH_KEY',         '${WORDPRESS_AUTH_KEY:-oocie1ihiet1Ucu3us3fiequ9yu2aep5}' );
define( 'SECURE_AUTH_KEY',  '${WORDPRESS_SECURE_AUTH_KEY:-aumahShoophei9ahlae4beeng0Iechud}' );
define( 'LOGGED_IN_KEY',    '${WORDPRESS_LOGGED_IN_KEY:-aisuish5aepaedae1sahtee1Voo6bee3}' );
define( 'NONCE_KEY',        '${WORDPRESS_NONCE_KEY:-thiegailuoTee5chikicaCii5ichie2S}' );
define( 'AUTH_SALT',        '${WORDPRESS_AUTH_SALT:-iaquaeH1chaid3Oom6eisaemoPhawozi}' );
define( 'SECURE_AUTH_SALT', '${WORDPRESS_SECURE_AUTH_SALT:-foo0Ilai4Ocie5Iophohyaja0iLie6ei}' );
define( 'LOGGED_IN_SALT',   '${WORDPRESS_LOGGED_IN_SALT:-eequaeFie5ohngab0aeg4at6Ootheeth}' );
define( 'NONCE_SALT',       '${WORDPRESS_NONCE_SALT:-xootei5uut3ii2lieno0eeFengo2vuJe}' );
$table_prefix = '${WORDPRESS_TABLE_PREFIX:-wp_}';
define( 'WP_DEBUG', ${WORDPRESS_DEBUG:-false} );
if ( ! defined( 'ABSPATH' ) ) {
	define( 'ABSPATH', __DIR__ . '/' );
}
require_once ABSPATH . 'wp-settings.php';
EOF

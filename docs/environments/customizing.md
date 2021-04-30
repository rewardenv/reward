## Customizing An Environment

Further information on customizing or extending an environment is forthcoming. For now, this section is limited to very simple and somewhat common customizations.

To configure your project with a non-default PHP version, add the following to the project's `.env` file and run `reward env up` to re-create the affected containers:

    PHP_VERSION=7.2

The versions of MariaDB, Elasticsearch, Varnish, Redis, and NodeJS may also be similarly configured using variables in the `.env` file:

* `MARIADB_VERSION`
* `ELASTICSEARCH_VERSION`
* `REDIS_VERSION`
* `VARNISH_VERSION`
* `RABBITMQ_VERSION`
* `NODE_VERSION`

Start of some environments could be skipped by using variables in `.env` file:

* `REWARD_DB=0`
* `REWARD_REDIS=0`

### Magento 2 Specific Customizations

The following variables can be added to the project's `.env` file to enable additional database containers for use with the Magento 2 (Commerce Only) [split-database solution](https://devdocs.magento.com/guides/v2.3/config-guide/multi-master/multi-master.html).

* `REWARD_SPLIT_SALES=1`
* `REWARD_SPLIT_CHECKOUT=1`

Start of some Magento 2 specific environments could be skipped by using variables in `.env` file:

* `REWARRD_ELASTICSEARCH=0`
* `REWARD_VARNISH=0`
* `REWARD_RABBITMQ=0`

### Database Specific Customizations

You can change the DB Server Character Set or Collation in the `.env` file:

* `MYSQL_CHARACTER_SET_SERVER=utf8mb4`
* `MYSQL_COLLATION_SERVER=utf8mb4_unicode_ci`

To disable Strict Mode in MySQL you will have to add the following line to the `.env` file:

* `MYSQL_DISABLE_STRICT_MODE=1`

### PHP Specific Customizations

The following PHP environment variables can be set in the `.env` file:

* `PHP_EXPOSE=Off`
* `PHP_ERROR_REPORTING=E_ALL`
* `PHP_DISPLAY_ERRORS=On`
* `PHP_DISPLAY_STARTUP_ERRORS=On`
* `PHP_LOG_ERRORS=On`
* `PHP_LOG_ERRORS_MAX_LEN=1024`
* `PHP_MAX_EXECUTION_TIME=3600`
* `PHP_MAX_INPUT_VARS=10000`
* `PHP_POST_MAX_SIZE=25M`
* `PHP_UPLOAD_MAX_FILESIZE=25M`
* `PHP_MAX_FILE_UPLOADS=20`
* `PHP_MEMORY_LIMIT=2G`
* `PHP_SESSION_AUTO_START=Off`
* `PHP_REALPATH_CACHE_SIZE=10M`
* `PHP_REALPATH_CACHE_TTL=7200`
* `PHP_DATE_TIMEZONE=UTC`
* `PHP_ZEND_ASSERTIONS=1`
* `PHP_SENDMAIL_PATH="/usr/local/bin/mhsendmail --smtp-addr='${MAILHOG_HOST:-mailhog}:${MAILHOG_PORT:-1025}'"`

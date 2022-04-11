## Database

You can change the DB Server Character Set or Collation in the `.env` file:

* `MYSQL_CHARACTER_SET_SERVER=utf8mb4`
* `MYSQL_COLLATION_SERVER=utf8mb4_unicode_ci`

To configure InnoDB Buffer Pool size, add the following line to the `.env` file:

* `MYSQL_INNODB_BUFFER_POOL_SIZE=256m`

To disable Strict Mode in MySQL you will have to add the following line to the `.env` file:

* `MYSQL_DISABLE_STRICT_MODE=1`

You can also set additional arguments to `mysqld` using the setting below in the `.env` file. To add multiple arguments,
set them as a single **space separated** string. See the available arguments in
the [MariaDB Docs](https://mariadb.com/kb/en/server-system-variables/).

* `MYSQL_ARGS="--innodb-buffer-pool-instances=4 --key-buffer-size=256M"`

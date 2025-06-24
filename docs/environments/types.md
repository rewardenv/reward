### Environment Types

Reward currently supports 8 environment types.

* Magento 1
* Magento 2
* PWA Studio (for Magento 2)
* Laravel
* Symfony
* Shopware
* WordPress
* Generic PHP
* Local

  These types are passed to `env-init` when configuring a project for local development for the first time. This list of
  environment types can also be seen by running `reward env-init --help` on your command line. The `docker compose`
  configuration used to assemble each environment type can be found in
  the [templates directory](https://github.com/itgcloud/reward/tree/main/templates) on GitHub.

#### Magento 1

The `magento1` environment type supports development of Magento 1 projects, launching containers including:

* Nginx
* PHP-FPM (5.6 or 7.0+)
* MariaDB
* Redis

Files are currently mounted using a delegated mount on macOS/Windows and natively on Linux.

#### Magento 2

The `magento2` environment type provides necessary containerized services for running Magento 2 in a local development
context including:

* Nginx
* Varnish
* PHP-FPM (7.0+)
* MariaDB
* OpenSearch (or Elasticsearch, disabled by default)
* RabbitMQ
- Valkey (or Redis, disabled by default)

In order to achieve a well performing experience on macOS and Windows, files in the webroot are synced into the
container using a Mutagen sync session except `pub/media` which remains mounted using a delegated mount.

#### PWA Studio

The `pwa-studio` environment type provides necessary containerized services for running PWA in a local development
context including:

* NodeJS (with yarn)

#### Laravel

The `laravel` environment type supports development of Laravel projects, launching containers including:

* Nginx
* PHP-FPM
* MariaDB
* Valkey (or Redis, disabled by default)

Files are currently mounted using a delegated mount on macOS/Windows and natively on Linux.

#### Symfony

The `symfony` environment type supports development of Symfony 4+ projects, launching containers including:

* Nginx
* PHP-FPM
* MariaDB
* Valkey (or Redis, disabled by default)
* RabbitMQ (disabled by default)
* Varnish (disabled by default)
* OpenSearch (or Elasticsearch, disabled by default)

Files are currently mounted using a delegated mount on macOS/Windows and natively on Linux.

#### Shopware

The `shopware` environment type supports development of Shopware 6 projects, launching containers including:

* Nginx
* PHP-FPM
* MariaDB
* Valkey (or Redis, disabled by default)
* RabbitMQ (disabled by default)
* Varnish (disabled by default)
* OpenSearch (or Elasticsearch, disabled by default)

In order to achieve a well performing experience on macOS and Windows, files in the webroot are synced into the
container using a Mutagen sync session except `public/media` which remains mounted using a delegated mount.

#### WordPress

The `wordpress` environment type supports development of WordPress 5 projects, launching containers including:

* Nginx
* PHP-FPM
* MariaDB
* Valkey (or Redis, disabled by default)

In order to achieve a well performing experience on macOS and Windows, files in the webroot are synced into the
container using a Mutagen sync session except `wp-content/uploads` which remains mounted using a delegated mount.

#### Generic PHP

The `generic-php` environment type contains nginx, php-fpm, php-debug, database (and an optional redis) containers.

Using this type, you will get a more generic development environment, with just serving the files from the current
directory.

It is useful for any other PHP frameworks and raw PHP development.

#### Local

The `local` environment type does nothing more than declare the `docker compose` version and label the project network
so Reward will recognize it as belonging to an environment orchestrated by Reward.

When this type is used, a `.reward/reward-env.yml` may be placed in the root directory of the project workspace to
define the desired containers, volumes, etc needed for the project. An example of a `local` environment type being used
can be found in the [Initializing a Custom Node Environment in a Subdomain](custom-environment.md) or
in [m2demo project](https://github.com/davidalger/m2demo).

Similar to the other environment type's base definitions, Reward supports a `reward-env.darwin.yml`,
`reward-env.linux.yml` and `reward-env.windows.yml`

#### Commonalities

In addition to the above, each environment type (except the `local` type) come with PHP setup to use `msmtp` to
ensure outbound email does not inadvertently leave your network and to support simpler testing of email functionality.
Mailbox may be accessed by navigating to [https://mailbox.reward.test/](https://mailbox.reward.test/) in a browser.

Where PHP is specified in the above list, there should be two `fpm` containers, `php-fpm` and `php-debug` in order to
provide Xdebug support. Use of Xdebug is enabled by setting the `XDEBUG_SESSION` cookie in your browser to direct the
request to the `php-debug` container. Shell sessions opened in the debug container via `reward debug` will also connect
PHP processes for commands on the CLI to Xdebug.

The configuration of each environment leverages a `base` configuration YAML file, and optionally a `darwin` and `linux`
file to add to `base` configuration anything which may be specific to a given host architecture (this is, for example,
how the `magento2` environment type works seamlessly on macOS with Mutagen sync sessions while using native filesystem
mounts on Linux hosts).

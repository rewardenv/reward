# How to build specific images locally

The workdir should be the Reward Project root

* nginx, varnish, elasticsearch, etc
``` bash
images/scripts/build.sh nginx
```

* PHP (build all php images)
``` bash
VERSION_LIST="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1" images/scripts/build.sh php
```

* PHP-CLI (build all php images for Debian)
``` bash
VERSION_LIST="5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1" images/scripts/build.sh php
VARIANT_LIST="cli fpm" VERSION_LIST="7.4 8.0 8.1" images/scripts/build.sh php
DOCKER_BASE_IMAGES="centos7 centos8" VARIANT_LIST="cli fpm" VERSION_LIST="7.4 8.0 8.1" images/scripts/build.sh php

DOCKER_BASE_IMAGES="debian" VARIANT_LIST="cli-loaders" VERSION_LIST="7.1" images/scripts/build.sh php
```

* PHP-FPM for Magento 2 for specific PHP version
``` bash
PHP_VERSION=7.4 images/scripts/build.sh php-fpm/magento2
```

## Command line options

* `--dry-run`: only print the commands the build script would run
* DEBUG=true: environment variable to call the bash script with setopt -x

Example: 

``` bash
$ DEBUG=true VERSION_LIST="7.4" images/scripts/build.sh --dry-run php
```

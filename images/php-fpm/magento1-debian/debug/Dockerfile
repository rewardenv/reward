ARG PHP_VERSION

FROM rewardenv/php-fpm:${PHP_VERSION}-magento1-debian

ARG PHP_VERSION

USER root

RUN set -eux \
  && apt-get update && apt-get install -y php${PHP_VERSION}-xdebug \
  && rm -rf /var/lib/apt/lists/*

COPY debug/etc/*.ini /etc/php/${PHP_VERSION}/cli
COPY debug/etc/php.d/15-xdebug.ini /etc/php/${PHP_VERSION}/mods-available/xdebug.ini

USER www-data

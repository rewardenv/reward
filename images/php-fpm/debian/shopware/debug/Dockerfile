ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-shopware-debian

ARG PHP_VERSION

USER root

RUN set -eux \
  && apt-get update && apt-get install -y php${PHP_VERSION}-xdebug \
  && rm -rf /var/lib/apt/lists/*

COPY debug/etc/*.ini /etc/php/${PHP_VERSION}/cli
COPY debug/etc/php.d/15-xdebug.ini /etc/php/${PHP_VERSION}/mods-available/xdebug.ini

USER www-data

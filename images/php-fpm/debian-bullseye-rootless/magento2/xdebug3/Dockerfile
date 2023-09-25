ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2-${IMAGE_BASE}

ARG PHP_VERSION

USER root

COPY xdebug3/rootfs/. /

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    php${PHP_VERSION}-xdebug \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && chown -R www-data: ~www-data

USER www-data

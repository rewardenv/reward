ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bookworm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-wordpress-${IMAGE_BASE}

ARG PHP_VERSION

USER root

COPY xdebug2/rootfs/. /

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    php${PHP_VERSION}-xdebug \
    && chown -R www-data: /etc/php /var/lib/php \
    && rm -rf /var/lib/apt/lists/* /var/log/apt

USER www-data

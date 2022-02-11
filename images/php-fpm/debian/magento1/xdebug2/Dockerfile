ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento1-${IMAGE_BASE}

ARG PHP_VERSION

USER root

COPY xdebug2/rootfs/. /

RUN set -eux \
    && sudo apt-get update && sudo apt-get install -y php${PHP_VERSION}-dev php-pear make \
    && sudo pecl install -f xdebug-2.9.8 \
    && chown -R www-data: /etc/php /var/lib/php \
    && rm -rf /var/lib/apt/lists/* /var/log/apt

USER www-data

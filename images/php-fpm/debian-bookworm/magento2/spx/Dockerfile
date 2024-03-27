ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bookworm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE} as builder

ARG PHP_VERSION

USER root


RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    php${PHP_VERSION}-dev \
    zlib1g-dev \
    && chown -R www-data: /etc/php /var/lib/php \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && cd /tmp \
    && git clone https://github.com/NoiseByNorthwest/php-spx.git \
    && cd php-spx \
    && git checkout release/latest \
    && phpize \
    && ./configure \
    && make \
    && make install

USER www-data

FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2-${IMAGE_BASE}

COPY spx/rootfs/. /
COPY --from=builder /tmp/php-spx/assets /usr/share/misc/php-spx/assets
COPY --from=builder /tmp/php-spx/modules/spx.so /tmp/php/extensions

RUN set -eux \
    && mv /tmp/php/extensions/spx.so "$(php-config --extension-dir)" \
    && chown -R www-data: /etc/php /var/lib/php

ENV SPX_ENABLED=1

USER www-data

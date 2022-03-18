ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER root

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    ack \
    build-essential \
    make \
    && rm -rf /var/lib/apt/lists/* /var/log/apt

USER www-data

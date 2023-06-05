ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="ubuntu-jammy"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

ARG NODE_VERSION=16
ENV NODE_VERSION ${NODE_VERSION}

USER root

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    ack \
    build-essential \
    make \
    && n install ${NODE_VERSION} \
    && rm -rf /var/lib/apt/lists/* /var/log/apt

USER www-data

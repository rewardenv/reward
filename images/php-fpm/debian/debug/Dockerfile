ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

ARG PHP_VERSION

USER root

RUN set -eux \
  && apt-get update && apt-get install -y php${PHP_VERSION}-xdebug \
  && rm -rf /var/lib/apt/lists/*

COPY debug/rootfs/. /

ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-debian

USER root

RUN set -eux \
  && apt-get update && apt-get install -y --no-install-recommends \
     ack \
     build-essential \
     make \
  && rm -rf /var/lib/apt/lists/*

USER www-data

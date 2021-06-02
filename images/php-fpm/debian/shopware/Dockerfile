ARG PHP_VERSION

FROM rewardenv/php-fpm:${PHP_VERSION}-debian

USER root

RUN set -eux \
  && apt-get update && apt-get install -y --no-install-recommends \
     ack \
     build-essential \
     make \
  && rm -rf /var/lib/apt/lists/*

USER www-data

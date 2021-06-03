ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}
USER root

RUN set -eux \
  && yum install -y \
     ack \
     gcc \
     gcc-c++ \
     make \
  && yum clean all && rm -rf /var/cache/yum

USER www-data

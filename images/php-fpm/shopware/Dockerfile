ARG PHP_VERSION
FROM rewardenv/php-fpm:${PHP_VERSION}
USER root

RUN set -eux \
  && yum install -y \
     ack \
     gcc \
     gcc-c++ \
     make \
  && yum clean all && rm -rf /var/cache/yum

USER www-data

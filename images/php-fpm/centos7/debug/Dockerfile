ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION

FROM ${IMAGE_NAME}:${PHP_VERSION}

USER root

RUN set -eux \
  && yum install -y php-pecl-xdebug \
  && yum clean all && rm -rf /var/cache/yum

COPY debug/etc/*.ini /etc/
COPY debug/etc/php.d/xdebug.ini /etc/php.d/15-xdebug.ini

USER www-data

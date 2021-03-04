ARG PHP_VERSION

FROM rewardenv/php-fpm:${PHP_VERSION}

USER root

RUN set -eux \
  && yum install -y php-pecl-xdebug \
  && yum clean all && rm -rf /var/cache/yum

COPY debug/etc/*.ini /etc/
COPY debug/etc/php.d/xdebug.ini /etc/php.d/15-xdebug.ini

USER www-data

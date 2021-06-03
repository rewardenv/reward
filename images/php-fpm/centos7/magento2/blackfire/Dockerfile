ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2
USER root

RUN set -eux \
  && yum install -y pygpgme \
  && curl -fsSLo- "http://packages.blackfire.io/fedora/blackfire.repo" | sudo tee /etc/yum.repos.d/blackfire.repo \
  && yum install -y blackfire-php \
  && yum clean all && rm -rf /var/cache/yum

COPY blackfire/etc/php.d/*.ini /etc/php.d/

USER www-data

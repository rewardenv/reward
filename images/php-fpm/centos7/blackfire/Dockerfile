ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION

FROM ${IMAGE_NAME}:${PHP_VERSION}

USER root

RUN set -eux \
  && yum install -y pygpgme \
  && curl -fsSLo- "http://packages.blackfire.io/fedora/blackfire.repo" | sudo tee /etc/yum.repos.d/blackfire.repo \
  && yum install -y blackfire-php \
  && yum clean all && rm -rf /var/cache/yum \
  && mkdir -p /tmp/blackfire \
  && curl -fsSLo- https://blackfire.io/api/v1/releases/client/linux_static/amd64 | tar zxp -C /tmp/blackfire \
  && mv /tmp/blackfire/blackfire /usr/bin/blackfire \
  && rm -rf /tmp/blackfire

COPY blackfire/etc/php.d/blackfire.ini.template /etc/php.d/

USER www-data

ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-wordpress-${IMAGE_BASE}

ARG PHP_VERSION

USER root

RUN set -eux \
  && apt-get update && apt-get install -y gnupg2 \
  && wget -q -O - https://packages.blackfire.io/gpg.key | sudo apt-key add - \
  && echo "deb http://packages.blackfire.io/debian any main" | sudo tee /etc/apt/sources.list.d/blackfire.list \
  && apt-get update && apt-get install blackfire-php \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /tmp/blackfire \
  && curl -fsSL https://blackfire.io/api/v1/releases/client/linux_static/amd64 | tar zxp -C /tmp/blackfire \
  && mv /tmp/blackfire/blackfire /usr/bin/blackfire \
  && rm -rf /tmp/blackfire

COPY blackfire/rootfs/. /

ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2-${IMAGE_BASE}

ARG PHP_VERSION

USER root

COPY blackfire/rootfs/. /

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    gnupg2 \
    && wget -q -O - https://packages.blackfire.io/gpg.key | apt-key add - \
    && echo "deb http://packages.blackfire.io/debian any main" >/etc/apt/sources.list.d/blackfire.list \
    && apt-get update && apt-get install -y --no-install-recommends \
    blackfire-php \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && mkdir -p /tmp/blackfire \
    && curl -fsSL https://blackfire.io/api/v1/releases/client/linux_static/amd64 | tar zxp -C /tmp/blackfire \
    && mv /tmp/blackfire/blackfire /usr/bin/blackfire \
    && chown -R www-data: /etc/php /var/lib/php \
    && rm -rf /tmp/blackfire

USER www-data

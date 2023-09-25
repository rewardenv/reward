ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-shopware-${IMAGE_BASE}

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
    && chown -R www-data: ~www-data

USER www-data

RUN set -eux \
    && mkdir -p /tmp/blackfire "${HOME}/.local/bin" \
    && curl -fsSL https://blackfire.io/api/v1/releases/client/linux_static/amd64 | tar zxp -C /tmp/blackfire \
    && mv /tmp/blackfire/blackfire "${HOME}/.local/bin/blackfire" \
    && chmod +x "${HOME}/.local/bin/blackfire" \
    && rm -rf /tmp/blackfire

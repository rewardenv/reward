ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION

FROM golang:alpine AS builder-supervisor
RUN set -eux \
    && apk add --no-cache --update git gcc rust \
    && mkdir -p /src \
    && cd /src \
    && git clone https://github.com/ochinchina/supervisord.git . \
    && if [ "$(apk --print-arch)" = "aarch64" ]; \
        then BUILD_ARCH="arm64"; \
        else BUILD_ARCH="amd64"; \
    fi \
    && CGO_ENABLED=0 GOOS=linux GOARCH=$BUILD_ARCH go build -a -ldflags "-linkmode external -extldflags -static" -o /usr/local/bin/supervisord github.com/ochinchina/supervisord

FROM ${IMAGE_NAME}:${PHP_VERSION}-shopware-${IMAGE_BASE}
USER root

ENV CRON_ENABLED            false
ENV SOCAT_ENABLED           false
ENV GOTTY_ENABLED           true
ENV GOTTY_USERNAME          shopware
ENV GOTTY_PASSWORD          shopware
ENV CHOWN_DIR_LIST          wp-content/uploads
ENV UID                     1000
ENV GID                     1000

ENV NGINX_UPSTREAM_HOST           127.0.0.1
ENV NGINX_UPSTREAM_PORT           9000
ENV NGINX_UPSTREAM_DEBUG_HOST     php-debug
ENV NGINX_UPSTREAM_DEBUG_PORT     9000
ENV NGINX_UPSTREAM_BLACKFIRE_HOST php-blackfire
ENV NGINX_UPSTREAM_BLACKFIRE_PORT 9000
ENV NGINX_ROOT                    /var/www/html
ENV NGINX_PUBLIC                  '/public'
ENV NGINX_TEMPLATE                shopware.conf
ENV XDEBUG_CONNECT_BACK_HOST      '""'
ENV WWWDATA_PASSWORD              ""

COPY shopware-web/rootfs/. /
COPY --from=builder-supervisor /usr/local/bin/supervisord /usr/bin/

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    gnupg2 \
    && echo "deb https://nginx.org/packages/debian/ $(lsb_release -sc) nginx" >/etc/apt/sources.list.d/nginx.list \
    && wget -q -O - https://nginx.org/keys/nginx_signing.key | apt-key add - \
    && apt-get update && apt-get install -y --no-install-recommends \
    nginx \
    netcat \
    && BUILD_ARCH="$(dpkg --print-architecture)" \
    && if [ "${BUILD_ARCH}" = "arm64" ]; \
        then GOTTY_ARCH="arm64"; \
        else GOTTY_ARCH="amd64"; \
    fi \
    && wget -q -O /tmp/gotty.tar.gz \
      "https://github.com/sorenisanerd/gotty/releases/download/v1.5.0/gotty_v1.5.0_linux_${GOTTY_ARCH}.tar.gz" \
    && tar -zxvf /tmp/gotty.tar.gz -C /usr/bin \
    && rm -f /tmp/gotty.tar.gz \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && rm -f /etc/supervisor/supervisord.conf.dpkg-dist \
    && mkdir -p /etc/supervisor/conf.d \
    && find /etc/supervisor /etc/nginx /etc/php /var/cache/nginx /var/lib/php /etc/ssl /var/run -exec sh -c "chgrp -v staff {} +; chmod -v g+w {} +" \; \
    && find /var/log -exec sh -c "chgrp -v adm {} +; chmod -v g+w {} +" \; \
    && ln -sf /etc/supervisor ~www-data/.local/etc \
    && ln -sf /etc/nginx ~www-data/.local/etc \
    && ln -sf /var/cache/nginx ~www-data/.local/var/cache \
    && chmod +x ~www-data/.local/bin/check-dependencies.sh ~www-data/.local/bin/install.sh ~www-data/.local/bin/stop-supervisor.sh \
    && chmod +x /docker-entrypoint.sh \
    && chown -R www-data: ~www-data

USER www-data

EXPOSE 4200
EXPOSE 8080

CMD ["supervisord", "-c", "/etc/supervisor/supervisord.conf"]

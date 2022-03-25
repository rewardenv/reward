ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2-${IMAGE_BASE}
USER root

ENV CRON_ENABLED   false
ENV SOCAT_ENABLED  false
ENV CHOWN_DIR_LIST pub/media
ENV UID            1000
ENV GID            1000

ENV NGINX_UPSTREAM_HOST           127.0.0.1
ENV NGINX_UPSTREAM_PORT           9000
ENV NGINX_UPSTREAM_DEBUG_HOST     php-debug
ENV NGINX_UPSTREAM_DEBUG_PORT     9000
ENV NGINX_UPSTREAM_BLACKFIRE_HOST php-blackfire
ENV NGINX_UPSTREAM_BLACKFIRE_PORT 9000
ENV NGINX_ROOT                    /var/www/html
ENV NGINX_PUBLIC                  '/pub'
ENV NGINX_TEMPLATE                magento2.conf
ENV XDEBUG_CONNECT_BACK_HOST      '""'
ENV SUDO_ENABLED                  "false"
ENV SUPERVISORD_VERSION           "0.7.3"

COPY magento2-web/rootfs/. /

USER root

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    gnupg2 \
    && echo "deb https://nginx.org/packages/debian/ $(lsb_release -sc) nginx" >/etc/apt/sources.list.d/nginx.list \
    && wget -q -O - https://nginx.org/keys/nginx_signing.key | apt-key add - \
    && BUILD_ARCH="$(dpkg --print-architecture)" \
    && if [ "${BUILD_ARCH}" = "arm64" ]; \
        then SUPERVISORD_ARCH="Linux_ARMv7"; \
        else SUPERVISORD_ARCH="Linux_64-bit"; \
    fi \
    && wget -q -O - https://github.com/ochinchina/supervisord/releases/download/v${SUPERVISORD_VERSION}/supervisord_${SUPERVISORD_VERSION}_${SUPERVISORD_ARCH}.tar.gz | tar zxv -C /tmp \
    && mv /tmp/supervisor*/supervisord /usr/bin/ \
    && rm -fr /tmp/supervisor* \
    && apt-get update && apt-get install -y --no-install-recommends \
    nginx \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && usermod -aG $GID nginx \
    && rm -f /etc/supervisor/supervisord.conf.dpkg-dist \
    && mkdir -p /etc/supervisor/conf.d \
    && chmod +x /usr/local/bin/install-magento2.sh /usr/local/bin/stop-supervisor.sh \
    && chown -R www-data: /etc/supervisor /etc/nginx /etc/php /var/log/nginx /var/cache/nginx /var/lib/php \
    && chmod +x /docker-entrypoint.sh
#    && ln -sf /dev/stdout /var/log/nginx/access.log && ln -sf /dev/stderr /var/log/nginx/error.log

USER www-data

EXPOSE 8080

CMD ["supervisord", "-c", "/etc/supervisor/supervisord.conf"]

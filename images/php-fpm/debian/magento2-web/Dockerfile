ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2-${IMAGE_BASE}
USER root

ENV CRON_ENABLED   true
ENV SOCAT_ENABLED  true
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

COPY magento2-web/docker-entrypoint magento2-web/stop-supervisor.sh /usr/local/bin/
COPY magento2-web/etc/. /etc/

RUN set -eux \
  && apt-get update && apt-get install -y gnupg2 \
  && echo "deb https://nginx.org/packages/debian/ $(lsb_release -sc) nginx" >/etc/apt/sources.list.d/nginx.list \
  && wget -q -O - https://nginx.org/keys/nginx_signing.key | apt-key add - \
  && apt-get update && apt-get install -y --no-install-recommends \
     nginx \
     supervisor \
  && rm -rf /var/lib/apt/lists/* \
  && usermod -aG $GID nginx \
  && rm -f /etc/supervisor/supervisord.conf.dpkg-dist \
  && mkdir -p /etc/supervisor/conf.d

USER www-data

EXPOSE 80

CMD ["sudo", "supervisord", "-c", "/etc/supervisor/supervisord.conf"]

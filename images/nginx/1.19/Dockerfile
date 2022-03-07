FROM nginx:1.19-alpine

COPY --from=hairyhenderson/gomplate:slim /gomplate /usr/local/bin/gomplate

RUN set -eux \
    && apk add --no-cache bash shadow \
    && groupmod -g 1000 www-data \
    && usermod -aG www-data nginx

ENV NGINX_UPSTREAM_HOST           php-fpm
ENV NGINX_UPSTREAM_PORT           9000
ENV NGINX_UPSTREAM_DEBUG_HOST     php-debug
ENV NGINX_UPSTREAM_DEBUG_PORT     9000
ENV NGINX_UPSTREAM_BLACKFIRE_HOST php-blackfire
ENV NGINX_UPSTREAM_BLACKFIRE_PORT 9000
ENV NGINX_ROOT                    /var/www/html
ENV NGINX_PUBLIC                  ''
ENV NGINX_TEMPLATE                application.conf
ENV XDEBUG_CONNECT_BACK_HOST      '""'
ENV NGINX_RESOLVER                127.0.0.11

COPY rootfs/. /

CMD find /etc/nginx -name '*.template' -exec sh -c 'gomplate <${1} >${1%.*}' sh {} \; \
    && nginx -g "daemon off;"

WORKDIR /var/www/html

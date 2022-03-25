FROM nginx:1.19-alpine

RUN set -eux \
    && apk add --no-cache bash shadow \
    && groupmod -g 1000 www-data \
    && usermod -aG www-data nginx \
    && BUILD_ARCH="$(apk --print-arch)" \
    && if [ "${BUILD_ARCH}" = "aarch64" ]; \
        then GOMPLATE_ARCH="arm64"; \
        else GOMPLATE_ARCH="amd64"; \
    fi \
    && wget -q -O /usr/local/bin/gomplate \
      "https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-${GOMPLATE_ARCH}" \
    && chmod +x /usr/local/bin/gomplate

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

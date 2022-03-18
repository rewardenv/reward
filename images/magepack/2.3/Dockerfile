FROM zenika/alpine-chrome:with-puppeteer

USER root

WORKDIR /var/www/html

COPY rootfs/. /

RUN set -eux \
    && npm install -g magepack@^2.3 \
    && npm cache clean --force

CMD tail -f /dev/null

USER chrome

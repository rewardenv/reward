# syntax=docker/dockerfile:1
FROM zenika/alpine-chrome:with-puppeteer

USER root

WORKDIR /var/www/html

COPY rootfs/. /

RUN <<-EOF
    set -eux
    npm install -g magepack@^{{ getenv "MAGEPACK_VERSION" "2.3" }}
    npm cache clean --force
EOF

CMD ["tail", "-f", "/dev/null"]

USER chrome

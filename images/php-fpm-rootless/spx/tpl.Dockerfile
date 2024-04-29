# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
{{- $PHP_VARIANT := getenv "PHP_VARIANT" "" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VARIANT="{{ if $PHP_VARIANT }}-{{ $PHP_VARIANT }}{{ end }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}${PHP_VARIANT}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless as builder

ARG PHP_VERSION

# hadolint ignore=DL3002
USER root

WORKDIR /tmp/php-spx

RUN <<-EOF
    set -eux
    apt-get update
    apt-get install -y --no-install-recommends \
      make \
      php${PHP_VERSION}-dev \
      zlib1g-dev
    chown -R www-data: /etc/php /var/lib/php
    rm -rf /var/lib/apt/lists/* /var/log/apt
    git clone https://github.com/NoiseByNorthwest/php-spx.git .
    git checkout release/latest
    phpize
    ./configure
    make
    make install
EOF

FROM ${IMAGE_NAME}:${PHP_VERSION}${PHP_VARIANT}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless

USER root

COPY rootfs/. /
COPY --from=builder /tmp/php-spx/assets /usr/share/misc/php-spx/assets
COPY --from=builder /tmp/php-spx/modules/spx.so /tmp/php-spx/modules/spx.so

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN <<-EOF
    set -eux
    mv /tmp/php-spx/modules/spx.so "$(php -i | grep extension_dir  | cut -d ' ' -f 5)"
    chown -R www-data: /etc/php /var/lib/php
EOF

ENV SPX_ENABLED=1

USER www-data

# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
{{- $PHP_VARIANT := getenv "PHP_VARIANT" "" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VARIANT="{{ if $PHP_VARIANT }}-{{ $PHP_VARIANT }}{{ end }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}${PHP_VARIANT}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless

ARG PHP_VERSION

USER root

COPY rootfs/. /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN <<-EOF
    set -eux
    apt-get update
    apt-get install -y --no-install-recommends \
      gnupg2
    wget -q -O - https://packages.blackfire.io/gpg.key | apt-key add -
    echo "deb http://packages.blackfire.io/debian any main" >/etc/apt/sources.list.d/blackfire.list
    apt-get update
    # blackfire-php 2026.7.0 dropped the PHP 5.6 probe
    BLACKFIRE_PHP="blackfire-php"
    if [ "${PHP_VERSION}" = "5.6" ]; then BLACKFIRE_PHP="blackfire-php=2026.5.0"; fi
    apt-get install -y --no-install-recommends \
      blackfire \
      "${BLACKFIRE_PHP}"
    rm -rf /var/lib/apt/lists/* /var/log/apt
    chown -R www-data: ~www-data
EOF

USER www-data


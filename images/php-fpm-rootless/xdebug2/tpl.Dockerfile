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

RUN <<-EOF
    set -eux
    apt-get update
    apt-get install -y --no-install-recommends \
      php${PHP_VERSION}-dev \
      php-pear \
      make
    eval 'version_gt() { test "$(printf '%s\n' "${@#v}" | sort -V | head -n 1)" != "${1#v}"; }'
    if version_gt "${PHP_VERSION}" "6.99.99"; \
      then if version_gt "${PHP_VERSION}" "7.0.99"; \
        then pecl install -f xdebug-2.9.8; \
        else pecl install -f xdebug-2.7.2; \
        fi \
      else pecl install -f xdebug-2.5.5; \
    fi
    chown -R www-data: ~www-data
    rm -rf /var/lib/apt/lists/* /var/log/apt
EOF

USER www-data

# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}

USER root

RUN <<-EOF
    set -eux
    npm install -g \
      grunt-cli \
      gulp \
      yarn
    curl -fsSLo /usr/local/bin/n98-magerun \
      https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun.phar
    chmod +x /usr/local/bin/n98-magerun
    curl -fsSLo /etc/bash_completion.d/n98-magerun.phar.bash \
      https://raw.githubusercontent.com/netz98/n98-magerun/master/res/autocompletion/bash/n98-magerun.phar.bash
    # Create mr alias for n98-magerun
    ln -fs /usr/local/bin/n98-magerun /usr/local/bin/mr
EOF

USER www-data

# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}-rootless

USER www-data

RUN <<-EOF
    set -eux
    npm install \
      grunt-cli \
      gulp \
      yarn
    curl -fsSLo "${HOME}/.local/bin/n98-magerun" \
      https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun.phar
    chmod +x "${HOME}/.local/bin/n98-magerun"
    mkdir -p "${HOME}/.local/share/bash-completion/completions"
    curl -fsSLo "${HOME}/.local/share/bash-completion/completions/n98-magerun.phar.bash" \
      https://raw.githubusercontent.com/netz98/n98-magerun/master/res/autocompletion/bash/n98-magerun.phar.bash
    ln -s "${HOME}/.local/bin/n98-magerun" "${HOME}/.local/bin/mr"
EOF

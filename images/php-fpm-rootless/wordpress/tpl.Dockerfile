# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless

USER www-data

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "wp-content/uploads"

RUN <<-EOF
    set -eux
    curl -fsSLo "${HOME}/.local/bin/wp" https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar
    chmod +x "${HOME}/.local/bin/wp"
EOF

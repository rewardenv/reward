# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}

ARG NODE_VERSION=18
ENV NODE_VERSION ${NODE_VERSION}

USER root

RUN <<-EOF
    set -eux
    apt-get update
    apt-get install -y --no-install-recommends \
      ack \
      build-essential \
      make
    n install ${NODE_VERSION}
    rm -rf /var/lib/apt/lists/* /var/log/apt
EOF

USER www-data

# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php-fpm"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless

ARG NODE_VERSION=18
ENV NODE_VERSION ${NODE_VERSION}

USER www-data

RUN <<-EOF
    set -eux
    n install "${NODE_VERSION}"
    rm -rf "${HOME}/.local/n/versions/node"
EOF

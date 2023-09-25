ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

ARG NODE_VERSION=18
ENV NODE_VERSION ${NODE_VERSION}

USER www-data

RUN set -eux \
    && n install "${NODE_VERSION}" \
    && rm -rf "${HOME}/.local/n/versions/node" \

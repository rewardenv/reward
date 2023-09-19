ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER www-data

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "wp-content/uploads"

RUN set -eux \
    && curl -fsSLo "${HOME}/.local/bin/wp" https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar \
    && chmod +x "${HOME}/.local/bin/wp"

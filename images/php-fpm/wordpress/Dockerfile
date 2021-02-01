ARG PHP_VERSION
FROM docker.io/rewardenv/php-fpm:${PHP_VERSION}
USER root

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "wp-content/uploads"

RUN set -eux \
    && curl -s https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar > /usr/bin/wp \
    && chmod +x /usr/bin/wp

USER www-data

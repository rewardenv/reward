ARG PHP_VERSION

FROM rewardenv/php-fpm:${PHP_VERSION}-debian

USER root

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "wp-content/uploads"

RUN set -eux \
  && curl -fsSLo /usr/bin/wp https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar \
  && chmod +x /usr/bin/wp

USER www-data

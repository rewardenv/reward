ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER root

RUN set -eux \
    && npm install -g \
    grunt-cli \
    gulp \
    yarn \
    && curl -fsSLo /usr/local/bin/n98-magerun \
      https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun.phar \
    && chmod +x /usr/local/bin/n98-magerun \
    && curl -fsSLo /etc/bash_completion.d/n98-magerun.phar.bash \
      https://raw.githubusercontent.com/netz98/n98-magerun/master/res/autocompletion/bash/n98-magerun.phar.bash \
    # Create mr alias for n98-magerun
    && ln -s /usr/local/bin/n98-magerun /usr/local/bin/mr

USER www-data

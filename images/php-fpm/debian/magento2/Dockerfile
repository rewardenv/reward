ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER root

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "pub/media"
ENV SUDO_ENABLED   "true"

RUN set -eux \
    && npm install -g \
    grunt-cli \
    gulp \
    yarn \
    && PHP_VERSION_STRIPPED=$(echo $PHP_VERSION | awk -F '.' '{print $1$2}') \
    && if [ "${PHP_VERSION_STRIPPED}" -ge 72 ]; then \
        MAGERUN_PHAR_URL=https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun2.phar; \
      else MAGERUN_PHAR_URL=https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun2-3.2.0.phar; \
      fi \
    && curl -fsSLo /usr/bin/n98-magerun ${MAGERUN_PHAR_URL} \
    && chmod +x /usr/bin/n98-magerun \
    && if [ "${PHP_VERSION_STRIPPED}" -ge 72 ]; then \
        MAGERUN_BASH_REF=master; \
      else MAGERUN_BASH_REF=3.2.0; \
      fi \
    && curl -fsSLo /etc/bash_completion.d/n98-magerun2.phar.bash \
      https://raw.githubusercontent.com/netz98/n98-magerun2/${MAGERUN_BASH_REF}/res/autocompletion/bash/n98-magerun2.phar.bash \
    && perl -pi -e 's/^(complete -o default .*)$/$1 n98-magerun/' /etc/bash_completion.d/n98-magerun2.phar.bash \
    # Create mr alias for n98-magerun
    && ln -s /usr/bin/n98-magerun /usr/bin/mr

USER www-data

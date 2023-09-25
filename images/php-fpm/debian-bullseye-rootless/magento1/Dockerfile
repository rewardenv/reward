ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER www-data

RUN set -eux \
    && npm install \
    grunt-cli \
    gulp \
    yarn \
    && curl -fsSLo "${HOME}/.local/bin/n98-magerun" \
      https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun.phar \
    && chmod +x "${HOME}/.local/bin/n98-magerun" \
    && mkdir -p "${HOME}/.local/share/bash-completion/completions" \
    && curl -fsSLo "${HOME}/.local/share/bash-completion/completions/n98-magerun.phar.bash" \
      https://raw.githubusercontent.com/netz98/n98-magerun/master/res/autocompletion/bash/n98-magerun.phar.bash \
    && ln -s "${HOME}/.local/bin/n98-magerun" "${HOME}/.local/bin/mr"


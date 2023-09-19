ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian-bullseye-rootless"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-${IMAGE_BASE}

USER www-data

# Resolve permission issues stemming from directories auto-created by docker due to mounts in sub-directories
ENV CHOWN_DIR_LIST "pub/media"

RUN set -eux \
    && npm install \
    grunt-cli \
    gulp \
    yarn \
    && PHP_VERSION_STRIPPED=$(echo $PHP_VERSION | awk -F '.' '{print $1$2}') \
    && if [ "${PHP_VERSION_STRIPPED}" -le 71 ]; then \
        MAGERUN_PHAR_URL=https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun2-3.2.0.phar; \
        MAGERUN_BASH_REF=3.2.0; \
      elif [ "${PHP_VERSION_STRIPPED}" -eq 72 ]; then \
        MAGERUN_PHAR_URL=https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun2-4.7.0.phar; \
        MAGERUN_BASH_REF=4.7.0; \
      else \
        MAGERUN_PHAR_URL=https://raw.githubusercontent.com/rewardenv/magerun-mirror/main/n98-magerun2.phar; \
        MAGERUN_BASH_REF=master; \
      fi \
    && curl -fsSLo "${HOME}/.local/bin/n98-magerun" ${MAGERUN_PHAR_URL} \
    && chmod +x "${HOME}/.local/bin/n98-magerun" \
    && mkdir -p "${HOME}/.local/share/bash-completion/completions" \
    && curl -fsSLo "${HOME}/.local/share/bash-completion/completions/n98-magerun2.phar.bash" \
      https://raw.githubusercontent.com/netz98/n98-magerun2/${MAGERUN_BASH_REF}/res/autocompletion/bash/n98-magerun2.phar.bash \
    && perl -pi -e 's/^(complete -o default .*)$/$1 n98-magerun/' "${HOME}/.local/share/bash-completion/completions/n98-magerun2.phar.bash" \
    # Create mr alias for n98-magerun
    && ln -s "${HOME}/.local/bin/n98-magerun" "${HOME}/.local/bin/mr"

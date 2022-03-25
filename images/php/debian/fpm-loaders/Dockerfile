ARG IMAGE_NAME="rewardenv/php"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-fpm-debian

USER root

RUN set -eux \
    # Install and enable Source Gaurdian loader
    && mkdir -p /tmp/sourceguardian \
    && cd /tmp/sourceguardian \
    && BUILD_ARCH="$(dpkg --print-architecture)" \
    && if [ "${BUILD_ARCH}" = "arm64" ]; \
        then SOURCEGUARDIAN_ARCH="aarch64"; \
        else SOURCEGUARDIAN_ARCH="x86_64"; \
    fi \
    && curl -fsSLO "https://www.sourceguardian.com/loaders/download/loaders.linux-${SOURCEGUARDIAN_ARCH}.tar.gz" \
    && tar xzf "loaders.linux-${SOURCEGUARDIAN_ARCH}.tar.gz" \
    && SOURCEGUARDIAN_LOADER_PATH="ixed.${PHP_VERSION}.lin" \
    && if [ -f "${SOURCEGUARDIAN_LOADER_PATH}" ]; then \
        cp "${SOURCEGUARDIAN_LOADER_PATH}" "$(php -i | grep '^extension_dir =' | cut -d' ' -f3)/sourceguardian.so" \
          && printf "; priority=15\nextension=sourceguardian.so" > /etc/php/${PHP_VERSION}/mods-available/sourceguardian.ini \
          && phpenmod sourceguardian; \
      else \
        >&2 printf "\033[33mWARNING\033[0m: SourceGuardian loaders for PHP_VERSION %s could not be found at %s\n" \
          "${PHP_VERSION}" "${SOURCEGUARDIAN_LOADER_PATH}"; \
      fi \
    && rm -rf /tmp/sourceguardian \
    # Install and enable IonCube loader
    && mkdir -p /tmp/ioncube \
    && cd /tmp/ioncube \
    && BUILD_ARCH="$(dpkg --print-architecture)" \
    && if [ "${BUILD_ARCH}" = "arm64" ]; \
        then IONCUBE_ARCH="aarch64"; \
        else IONCUBE_ARCH="x86-64"; \
    fi \
    && curl -fsSLO "https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_${IONCUBE_ARCH}.tar.gz" \
    && tar xzf "ioncube_loaders_lin_${IONCUBE_ARCH}.tar.gz" \
    && IONCUBE_LOADER_PATH="ioncube/ioncube_loader_lin_${PHP_VERSION}.so" \
    && if [ -f "${IONCUBE_LOADER_PATH}" ]; then \
        cp "${IONCUBE_LOADER_PATH}" "$(php -i | grep '^extension_dir =' | cut -d' ' -f3)/ioncube_loader.so" \
          && printf "; priority=01\nzend_extension=ioncube_loader.so" > /etc/php/${PHP_VERSION}/mods-available/ioncube-loader.ini \
          && phpenmod ioncube-loader; \
      else \
        >&2 printf "\033[33mWARNING\033[0m: IonCube loaders for PHP_VERSION %s could not be found at %s\n" \
          "${PHP_VERSION}" "${IONCUBE_LOADER_PATH}"; \
      fi \
    && rm -rf /tmp/ioncube

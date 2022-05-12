ARG IMAGE_NAME="rewardenv/php-fpm"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-shopware-${IMAGE_BASE}

ARG PHP_VERSION

USER root

COPY xdebug2/rootfs/. /

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    php${PHP_VERSION}-dev \
    php-pear \
    make \
    && eval 'version_gt() { test "$(printf "%s\n" "$@" | sort -V | head -n 1)" != "$1"; }' \
    && if version_gt "${PHP_VERSION}" "6.99.99"; \
        then if version_gt "${PHP_VERSION}" "7.0.99"; \
              then pecl install -f xdebug-2.9.8; \
              else pecl install -f xdebug-2.7.2; \
            fi \
        else pecl install -f xdebug-2.5.5; \
      fi \
    && chown -R www-data: /etc/php /var/lib/php \
    && rm -rf /var/lib/apt/lists/* /var/log/apt

USER www-data

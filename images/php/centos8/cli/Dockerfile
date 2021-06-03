FROM centos:8

ARG PHP_VERSION
ARG PHP_EXTENSIONS="bcmath cli common gd gmp intl json mbstring \
    mcrypt sodium mysqlnd opcache pdo pecl-msgpack pecl-amqp pecl-redis pecl-imagick pecl-zip process soap xml xmlrpc"

ENV COMPOSER_ALLOW_SUPERUSER 1
ENV COMPOSER_HOME /tmp/composer

COPY --from=composer:1 /usr/bin/composer /usr/bin/composer1
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer2
COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

RUN set -eux \
  && dnf install -y epel-release \
  && dnf install -y \
     bzip2 \
     git \
     jq \
     npm \
     patch \
     perl \
     unzip \
     which \
  && dnf clean all && rm -rf /var/cache/dnf \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer1 99 \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer2 1

RUN set -eux \
  && dnf install -y https://rpms.remirepo.net/enterprise/remi-release-8.rpm \
  && dnf module reset php \
  && dnf module install -y php:remi-${PHP_VERSION} \
  && PHP_VERSION=$(echo ${PHP_VERSION} | awk -F '.' '{print $1$2}') \
  && PHP_PACKAGES= && for PKG in ${PHP_EXTENSIONS}; do \
       if [ "${PKG}" = "mcrypt" ] && [ "${PHP_VERSION}" -gt 71 ]; then continue; fi; \
       if [ "${PKG}" = "sodium" ] && [ "${PHP_VERSION}" -lt 72 ]; then continue; fi; \
       PHP_PACKAGES="${PHP_PACKAGES:+${PHP_PACKAGES} }php-${PKG} "; \
     done \
  && dnf install -y ${PHP_PACKAGES} \
  && dnf clean all && rm -rf /var/cache/dnf \
  && cp /etc/php.ini /etc/php-cli.ini \
  && perl -pi -e 's/memory_limit(.*)/memory_limit=-1/g' /etc/php-cli.ini

FROM centos:7

ARG PHP_VERSION
ARG PHP_EXTENSIONS="bcmath cli common gd gmp intl json mbstring \
    mcrypt sodium mysqlnd opcache pdo pecl-msgpack pecl-amqp pecl-redis pecl-imagick pecl-zip process soap xml xmlrpc"

ENV PHP_VERSION                 $PHP_VERSION
ENV COMPOSER_ALLOW_SUPERUSER    1
ENV COMPOSER_HOME               /tmp/composer

COPY --from=composer:1 /usr/bin/composer /usr/bin/composer1
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer2
COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

RUN set -eux \
  && yum install -y epel-release \
  && yum install -y https://repo.ius.io/ius-release-el7.rpm \
  && yum-config-manager --disable ius \
  && yum install -y \
     bzip2 \
     jq \
     npm \
     patch \
     perl \
     unzip \
     which \
  && yum install -y --enablerepo=ius git222 \
  && yum clean all && rm -rf /var/cache/yum \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer1 99 \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer2 1

RUN set -eux \
  && yum install -y https://rpms.remirepo.net/enterprise/remi-release-7.rpm \
  && PHP_VERSION_STRIPPED=$(echo ${PHP_VERSION} | awk -F '.' '{print $1$2}') \
  && yum-config-manager --enable remi-php${PHP_VERSION} \
  && PHP_PACKAGES= && for PKG in ${PHP_EXTENSIONS}; do \
       if [ "${PKG}" = "mcrypt" ] && [ "${PHP_VERSION_STRIPPED}" -gt 71 ]; then continue; fi; \
       if [ "${PKG}" = "sodium" ] && [ "${PHP_VERSION_STRIPPED}" -lt 72 ]; then continue; fi; \
       PHP_PACKAGES="${PHP_PACKAGES:+${PHP_PACKAGES} }php-${PKG} "; \
     done \
  && yum install -y ${PHP_PACKAGES} \
  && yum clean all && rm -rf /var/cache/yum \
  && cp /etc/php.ini /etc/php-cli.ini \
  && perl -pi -e 's/memory_limit(.*)/memory_limit=-1/g' /etc/php-cli.ini

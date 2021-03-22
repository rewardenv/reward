FROM debian:buster-slim

ARG PHP_VERSION
ARG PHP_EXTENSIONS="amqp apcu bcmath bz2 cli common curl gd gmp imagick intl json mbstring mcrypt msgpack mysql opcache redis soap xml xmlrpc zip"

ENV DEBIAN_FRONTEND="noninteractive"
ENV COMPOSER_ALLOW_SUPERUSER="1"
ENV COMPOSER_HOME="/tmp/composer"

COPY --from=composer:1 /usr/bin/composer /usr/bin/composer1
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer2
COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

RUN set -eux \
  && echo 'apt::install-recommends "false";' > /etc/apt/apt.conf.d/no-install-recommends \
  && echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/keepconfig \
  && echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections \
  && apt-get update && apt-get install -y --no-install-recommends \
     apt-transport-https \
     bzip2 \
     ca-certificates \
     curl \
     #gettext \
     git \
     lsb-release \
     npm \
     patch \
     perl \
     unzip \
  && rm -rf /var/lib/apt/lists/* \
#  && curl -fsSLo /usr/bin/envsubst \
#     "https://github.com/a8m/envsubst/releases/download/v1.2.0/envsubst-`uname -s`-`uname -m`" \
#  && chmod +x /usr/bin/envsubst \
  # make `alternatives` command behave the same as on centos
  && update-alternatives --install /usr/bin/alternatives alternatives /usr/bin/update-alternatives 1 \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer1 99 \
  && alternatives --install /usr/bin/composer composer /usr/bin/composer2 1

RUN set -eux \
  && curl -fsSLo /etc/apt/trusted.gpg.d/php.gpg https://packages.sury.org/php/apt.gpg \
  && echo "deb https://packages.sury.org/php/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/php.list \
  && PHP_VERSION_STRIPPED=$(echo ${PHP_VERSION} | awk -F '.' '{print $1$2}') \
  && PHP_PACKAGES= && for PKG in ${PHP_EXTENSIONS}; do \
       if [ "${PKG}" = "json" ] && [ "${PHP_VERSION_STRIPPED}" -ge 80 ]; then continue; fi; \
       PHP_PACKAGES="${PHP_PACKAGES:+${PHP_PACKAGES} }php${PHP_VERSION}-${PKG} "; \
     done \
  && apt-get update && apt-get install -y --no-install-recommends ${PHP_PACKAGES} \
  && rm -rf /var/lib/apt/lists/*

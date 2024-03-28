# syntax=docker/dockerfile:1
{{- $IMAGE_NAME := getenv "IMAGE_NAME" "ubuntu" }}
{{- $IMAGE_TAG := getenv "IMAGE_TAG" "jammy" }}
FROM {{ $IMAGE_NAME }}:{{ $IMAGE_TAG }}

ARG PHP_VERSION
ARG PHP_EXTENSIONS="amqp apcu bcmath bz2 cli common curl gd gmp imagick intl json mbstring mcrypt msgpack mysql opcache pgsql redis soap xml xmlrpc zip"

ENV PHP_VERSION                 $PHP_VERSION
ENV DEBIAN_FRONTEND             noninteractive
ENV COMPOSER_ALLOW_SUPERUSER    1
ENV COMPOSER_HOME               /tmp/composer

# hadolint ignore=DL3002
USER root

COPY --from=composer:1 /usr/bin/composer /usr/local/bin/composer1
COPY --from=composer:2 /usr/bin/composer /usr/local/bin/composer2

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN <<-EOF
    set -eux
    echo 'apt::install-recommends "false";' > /etc/apt/apt.conf.d/no-install-recommends
    echo 'force-confold' > /etc/dpkg/dpkg.cfg.d/keepconfig
    echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
    apt-get update
    apt-get install -y --no-install-recommends \
      apt-transport-https \
      bzip2 \
      ca-certificates \
      curl \
      git \
      {{- if eq $IMAGE_NAME "ubuntu" }}
      gpg-agent \
      {{- end }}
      lsb-release \
      npm \
      patch \
      perl \
      unzip
    rm -rf /var/lib/apt/lists/* /var/log/apt
    # make `alternatives` command behave the same as on centos
    update-alternatives --install /usr/bin/alternatives alternatives /usr/bin/update-alternatives 1
    alternatives --install /usr/local/bin/composer composer /usr/local/bin/composer1 1
    alternatives --install /usr/local/bin/composer composer /usr/local/bin/composer2 99
    # PHP Packages
    {{- if eq $IMAGE_NAME "ubuntu" }}
    apt-get update
    apt-get install -y --no-install-recommends software-properties-common
    LC_ALL=C.UTF-8 add-apt-repository -y ppa:ondrej/php
    {{- else if eq $IMAGE_NAME "debian" }}
    curl -fsSLo /etc/apt/trusted.gpg.d/php.gpg https://packages.sury.org/php/apt.gpg
    echo "deb https://packages.sury.org/php/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/php.list
    {{- end }}
    PHP_VERSION_STRIPPED=$(echo ${PHP_VERSION} | awk -F '.' '{print $1$2}')
    PHP_PACKAGES=
    for PKG in ${PHP_EXTENSIONS}; do \
      if [ "${PKG}" = "json" ] && [ "${PHP_VERSION_STRIPPED}" -ge 80 ]; then continue; fi; \
      if [ "${PKG}" = "mcrypt" ] && [ "${PHP_VERSION_STRIPPED}" -ge 83 ]; then continue; fi; \
      PHP_PACKAGES="${PHP_PACKAGES:+${PHP_PACKAGES} }php${PHP_VERSION}-${PKG} "; \
    done
    # Adding apt-get upgrade -y to fix issue with libpcre
    # https://github.com/oerdnj/deb.sury.org/issues/1682
    BUILD_ARCH="$(dpkg --print-architecture)"
    if [ "${BUILD_ARCH}" = "arm64" ]; \
      then GOMPLATE_ARCH="arm64"; \
      else GOMPLATE_ARCH="amd64"; \
    fi
    curl -fsSLo /usr/local/bin/gomplate \
      "https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-${GOMPLATE_ARCH}"
    chmod +x /usr/local/bin/gomplate
    apt-get update
    apt-get upgrade -y
    # shellcheck disable=SC2086
    apt-get install -y --no-install-recommends ${PHP_PACKAGES}
    rm -rf /var/lib/apt/lists/* /var/log/apt
EOF

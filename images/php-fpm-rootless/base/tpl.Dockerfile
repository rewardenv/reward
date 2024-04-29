# syntax=docker/dockerfile:1
{{- $BASE_IMAGE_NAME := getenv "BASE_IMAGE_NAME" "ubuntu" }}
{{- $BASE_IMAGE_TAG := getenv "BASE_IMAGE_TAG" "jammy" }}
ARG IMAGE_NAME="rewardenv/php"
ARG BASE_IMAGE_NAME="{{ $BASE_IMAGE_NAME }}"
ARG BASE_IMAGE_TAG="{{ $BASE_IMAGE_TAG }}"
ARG PHP_VERSION
ARG PHP_VARIANT="fpm-loaders"

FROM ${IMAGE_NAME}:${PHP_VERSION}-${PHP_VARIANT}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}-rootless

ARG PHP_VERSION

# Clear undesired settings from base fpm images
ENV COMPOSER_ALLOW_SUPERUSER=""
ENV COMPOSER_HOME=""

ENV MAILBOX_HOST    mailbox
ENV MAILBOX_PORT    1025
ENV NODE_VERSION    16

COPY rootfs/. /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN <<-EOF
    set -eux
    apt-get update
    apt-get install -y --no-install-recommends \
      autoconf \
      automake \
      bash-completion \
      bsd-mailx \
      cron \
      default-mysql-client \
      dnsutils \
      less \
      jq \
      msmtp \
      msmtp-mta \
      nano \
      python3-pip \
      pwgen \
      rsync \
      socat \
      vim \
      wget
    rm -rf /var/lib/apt/lists/* /var/log/apt
    # Install awscli to support data backfill workflows using S3 storage
    pip3 install {{ if eq $BASE_IMAGE_TAG "bookworm" }}--break-system-packages{{ end }} --no-cache-dir --upgrade pip
    pip3 install {{ if eq $BASE_IMAGE_TAG "bookworm" }}--break-system-packages{{ end }} --no-cache-dir awscli
    # Configure Bash
    { \
      echo; \
      echo 'if [ -d /etc/profile.d ]; then'; \
      echo '  for i in /etc/profile.d/*.sh; do'; \
      echo '    if [ -r "$i" ]; then'; \
      echo '      . $i'; \
      echo '    fi'; \
      echo '  done'; \
      echo '  unset i'; \
      echo 'fi'; \
    } | tee -a /etc/bash.bashrc
    # Configure www-data user as primary php-fpm user for better local dev experience
    useradd www-data || true
    usermod -d /home/www-data -u 1000 --shell /bin/bash www-data
    groupmod -g 1000 www-data
    usermod -aG staff,adm www-data
    mkhomedir_helper www-data
    mkdir -p \
      ~www-data/.local/bin \
      ~www-data/.local/etc \
      ~www-data/.local/share \
      ~www-data/.local/var/cache \
      ~www-data/.local/var/lib \
      ~www-data/.local/var/run
    chown -R www-data: ~www-data
    chmod 0775 ~www-data
    mkdir -p /var/www/html
    chown www-data: /var/www/html
    ln -sf /etc/php ~www-data/.local/etc
    ln -sf /var/lib/php ~www-data/.local/var/lib
    find /var/log -exec sh -c "chgrp -v adm {} +; chmod -v g+w {} +" \;
    find /etc/php /etc/ssl /usr/local/share/ca-certificates /var/lib/php /var/run -exec sh -c "chgrp -v staff {} +; chmod -v g+w {} +" \;
    chmod u+s /usr/sbin/cron
    touch /var/run/crond.pid
    chgrp -v staff /var/run/crond.pid
    chmod +x /docker-entrypoint.sh
EOF

WORKDIR /home/www-data
USER www-data

ENV PATH="/var/www/html/node_modules/.bin:/home/www-data/node_modules/.bin:/home/www-data/.local/bin:${PATH}"
ENV N_PREFIX="/home/www-data/.local"

COPY --from=composer:1 /usr/bin/composer /home/www-data/.local/bin/composer1
COPY --from=composer:2 /usr/bin/composer /home/www-data/.local/bin/composer2

RUN <<-EOF
    set -eux
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --install "${HOME}/.local/bin/composer" composer "${HOME}/.local/bin/composer1" 1
    alternatives --altdir ~/.local/etc/alternatives --admindir ~/.local/var/lib/alternatives --install "${HOME}/.local/bin/composer" composer "${HOME}/.local/bin/composer2" 99
    pip3 install awscli --no-cache-dir
    npm install n
    n install "${NODE_VERSION}"
    rm -rf "${HOME}/.local/n/versions/node"
    perl -pi -e 's/^(user|group) = php-fpm$/$1 = www-data/g' "${HOME}/.local/etc/php/${PHP_VERSION}/fpm/pool.d/www.conf"
EOF

WORKDIR /var/www/html

ARG IMAGE_NAME="rewardenv/php"
ARG IMAGE_BASE="debian"
ARG PHP_VERSION
ARG PHP_VARIANT="fpm-loaders"
FROM ${IMAGE_NAME}:${PHP_VERSION}-${PHP_VARIANT}-${IMAGE_BASE}

ARG PHP_VERSION

# Clear undesired settings from base fpm images
ENV COMPOSER_ALLOW_SUPERUSER=""
ENV COMPOSER_HOME=""

ENV MAILHOG_HOST    mailhog
ENV MAILHOG_PORT    1025
ENV NODE_VERSION    10

COPY rootfs/. /

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    bash-completion \
    cron \
    default-mysql-client \
    dnsutils \
    less \
    jq \
    nano \
    python3-pip \
    pwgen \
    rsync \
    socat \
    sudo \
    vim \
    wget \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    # Install mhsendmail to support routing email through mailhog \
    && BUILD_ARCH="$(dpkg --print-architecture)" \
    && if [ "${BUILD_ARCH}" = "arm64" ]; \
        then MHSENDMAIL_ARCH="arm"; \
        else MHSENDMAIL_ARCH="amd64" ; \
    fi \
    && curl -fsSLo /usr/local/bin/mhsendmail \
      "https://github.com/mailhog/mhsendmail/releases/latest/download/mhsendmail_linux_${MHSENDMAIL_ARCH}" \
    && chmod +x /usr/local/bin/mhsendmail \
    && ln -sf /usr/local/bin/mhsendmail /usr/sbin/sendmail \
    # Install awscli to support data backfill workflows using S3 storage
    && pip3 install --upgrade pip \
    && pip3 install awscli --no-cache-dir \
    # Install node
    && npm install -g n \
    && n install ${NODE_VERSION} \
    && rm -rf /usr/local/n/versions/node \
    # Configure Bash
    && { \
      echo; \
      echo 'if [ -d /etc/profile.d ]; then'; \
      echo '  for i in /etc/profile.d/*.sh; do'; \
      echo '    if [ -r $i ]; then'; \
      echo '      . $i'; \
      echo '    fi'; \
      echo '  done'; \
      echo '  unset i'; \
      echo 'fi'; \
      } | tee -a /etc/bash.bashrc \
    # Configure www-data user as primary php-fpm user for better local dev experience
    && useradd www-data || true \
    && usermod -d /home/www-data -u 1000 --shell /bin/bash www-data \
    && groupmod -g 1000 www-data \
    && mkhomedir_helper www-data \
    && chmod 0755 ~www-data \
    && mkdir -p /var/www/html \
    && PHP_FPM_USER=$(grep -i '^user = ' /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf | awk '{print $3}') \
    && PHP_FPM_GROUP=$(grep -i '^group = ' /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf | awk '{print $3}') \
    && find /var/log /var/lib/php -uid $(id -u ${PHP_FPM_USER}) -exec chown -v www-data {} + \
    && find /var/log /var/lib/php -gid $(id -g ${PHP_FPM_GROUP}) -exec chgrp -v www-data {} + \
    && perl -pi -e 's/^(user|group) = php-fpm$/$1 = www-data/g' /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf \
    && chown www-data:www-data /var/www/html \
    && chown -R www-data: /etc/php /var/lib/php \
    && usermod -aG sudo www-data \
    && echo "%sudo ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/sudo \
    && chmod +x /docker-entrypoint.sh

WORKDIR /var/www/html
USER www-data

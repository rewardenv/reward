#TODO
ARG PHP_VERSION
ARG PHP_VARIANT="fpm-loaders"
FROM rewardenv/php:${PHP_VERSION}-${PHP_VARIANT}

# Clear undesired settings from base fpm images
ENV COMPOSER_ALLOW_SUPERUSER=""
ENV COMPOSER_HOME=""

ENV MAILHOG_HOST    mailhog
ENV MAILHOG_PORT    1025

COPY docker-entrypoint /usr/local/bin/
COPY etc/profile.d/*.sh /etc/profile.d/
COPY etc/php.d/*.ini.template /etc/php.d/

RUN set -eux \
  && yum install -y \
     autoconf \
     automake \
     bash-completion \
     bind-utils \
     crontabs \
     jq \
     nano \
     nmap-ncat \
     python3-pip \
     pwgen \
     rsync \
     socat \
     sudo \
     vim \
     wget \
     which \
  && yum install -y --enablerepo=ius mariadb102 \
  && yum clean all && rm -rf /var/cache/yum \
  # Install mhsendmail to support routing email through mailhog
  && curl -fsSLo /usr/local/bin/mhsendmail \
     https://github.com/mailhog/mhsendmail/releases/latest/download/mhsendmail_linux_amd64 \
  && chmod +x /usr/local/bin/mhsendmail \
  && pip3 install --upgrade pip \
  && pip3 install awscli --no-cache-dir \
  && envsubst < /etc/php.d/opcache.ini.template > /etc/php.d/10-opcache.ini

ENV NODE_VERSION    10

RUN set -eux \
  && npm install -g n \
  && n install ${NODE_VERSION} \
  && rm -rf /usr/local/n/versions/node

# Configure www-data user as primary php-fpm user for better local dev experience
RUN set -eux \
  && useradd www-data || true \
  && mkdir -p /var/www \
  && chmod 0755 ~www-data \
  && mkdir -p /var/www/html \
  && chown www-data:www-data /var/www/html \
  && echo "www-data ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd \
  && PHP_FPM_USER=$(grep -i '^user = ' /etc/php-fpm.d/www.conf | awk '{print $3}') \
  && PHP_FPM_GROUP=$(grep -i '^group = ' /etc/php-fpm.d/www.conf | awk '{print $3}') \
  && find /var/log /var/lib/php -uid $(id -u ${PHP_FPM_USER}) | xargs chown www-data \
  && find /var/log /var/lib/php -gid $(id -g ${PHP_FPM_GROUP}) | xargs chgrp www-data \
  && perl -pi -e 's/^(user|group) = php-fpm$/$1 = www-data/g' /etc/php-fpm.d/www.conf

WORKDIR /var/www/html
USER www-data

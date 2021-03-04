ARG PHP_VERSION

FROM rewardenv/php-fpm:${PHP_VERSION}-magento2-debian

ARG PHP_VERSION

USER root

RUN set -eux \
  && apt-get update && apt-get install -y gnupg2 \
  && wget -q -O - https://packages.blackfire.io/gpg.key | sudo apt-key add - \
  && echo "deb http://packages.blackfire.io/debian any main" | sudo tee /etc/apt/sources.list.d/blackfire.list \
  && apt-get update \
  && apt-get install blackfire-php \
  && rm -rf /var/lib/apt/lists/*

COPY blackfire/etc/php.d/*.ini /etc/php/${PHP_VERSION}/fpm/conf.d/

USER www-data

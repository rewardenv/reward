ARG IMAGE_NAME="rewardenv/php"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-debian

RUN set -eux \
  && PHP_VERSION=$(php -v | head -n1 | cut -d' ' -f2 | cut -d. -f1-2) \
  # Install and enable Source Gaurdian loader
  && mkdir -p /tmp/sourceguardian \
  && cd /tmp/sourceguardian \
  && curl -fsSLO https://www.sourceguardian.com/loaders/download/loaders.linux-x86_64.tar.gz \
  && tar xzf loaders.linux-x86_64.tar.gz \
  && cp ixed.${PHP_VERSION}.lin "$(php -i | grep '^extension_dir =' | cut -d' ' -f3)/sourceguardian.so" \
  && echo "extension=sourceguardian.so" > /etc/php/${PHP_VERSION}/cli/conf.d/15-sourceguardian.ini \
  && rm -rf /tmp/sourceguardian \
  # Install and enable IonCube loader
  && mkdir -p /tmp/ioncube \
  && cd /tmp/ioncube \
  && curl -fsSLO https://downloads.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz \
  && tar xzf ioncube_loaders_lin_x86-64.tar.gz \
  && cp ioncube/ioncube_loader_lin_${PHP_VERSION}.so "$(php -i | grep '^extension_dir =' | cut -d' ' -f3)/ioncube_loader.so" \
  && echo "zend_extension=ioncube_loader.so" > /etc/php/${PHP_VERSION}/cli/conf.d/01-ioncube-loader.ini \
  && rm -rf /tmp/ioncube

ARG IMAGE_NAME="rewardenv/php"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-debian

ARG PHP_VERSION

USER root

COPY rootfs/. /
ENTRYPOINT ["/docker-entrypoint.sh"]

RUN set -eux \
    && apt-get update && apt-get install -y --no-install-recommends \
    php${PHP_VERSION}-fpm \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
	  && cd /etc/php/${PHP_VERSION}/fpm \
	  && { \
      echo '[global]'; \
      echo 'error_log = /proc/self/fd/2'; \
      echo; \
      echo '[www]'; \
      echo '; if we send this to /proc/self/fd/1, it never appears'; \
      echo 'access.log = /proc/self/fd/2'; \
      echo; \
      echo 'clear_env = no'; \
      echo; \
      echo '; Ensure worker stdout and stderr are sent to the main error log.'; \
      echo 'catch_workers_output = yes'; \
      echo; \
      echo '; Enable ping path.'; \
      echo 'ping.path = /ping'; \
      } | tee pool.d/docker.conf \
	  && { \
      echo '[global]'; \
      echo 'daemonize = no'; \
      echo; \
      echo '[www]'; \
      echo 'listen = 9000'; \
      } | tee pool.d/zz-docker.conf \
    && perl -pi -e 's/^(pid|error_log|daemonize)/;$1/g' /etc/php/${PHP_VERSION}/fpm/php-fpm.conf \
    && perl -pi -e 's/^(listen)/;$1/g' /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf \
    && perl -pi -e 's/^(php_admin_(value|flag))/;$1/g' /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf \
    && alternatives --install /usr/sbin/php-fpm php-fpm /usr/sbin/php-fpm${PHP_VERSION} 99 \
    && chmod +x /docker-entrypoint.sh

# Override stop signal to stop process gracefully
# https://github.com/php/php-src/blob/17baa87faddc2550def3ae7314236826bc1b1398/sapi/fpm/php-fpm.8.in#L163
STOPSIGNAL SIGQUIT

EXPOSE 9000
CMD ["php-fpm"]

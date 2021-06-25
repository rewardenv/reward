ARG IMAGE_NAME="rewardenv/php-fpm"
ARG PHP_VERSION
FROM ${IMAGE_NAME}:${PHP_VERSION}-magento2
USER root

ENV CRON_ENABLED   true
ENV SOCAT_ENABLED  true
ENV CHOWN_DIR_LIST pub/media
ENV UID            1000
ENV GID            1000

ENV NGINX_UPSTREAM_HOST           127.0.0.1
ENV NGINX_UPSTREAM_PORT           9000
ENV NGINX_UPSTREAM_DEBUG_HOST     php-debug
ENV NGINX_UPSTREAM_DEBUG_PORT     9000
ENV NGINX_UPSTREAM_BLACKFIRE_HOST php-blackfire
ENV NGINX_UPSTREAM_BLACKFIRE_PORT 9000
ENV NGINX_ROOT                    /var/www/html
ENV NGINX_PUBLIC                  '/pub'
ENV NGINX_TEMPLATE                magento2.conf
ENV XDEBUG_CONNECT_BACK_HOST      '""'

COPY magento2-web/docker-entrypoint magento2-web/stop-supervisor.sh magento2-web/install-magento2.sh /usr/local/bin/
COPY magento2-web/etc/. /etc/

RUN set -eux \
  && { \
		 echo '[nginx-stable]'; \
		 echo 'name=nginx stable repo'; \
		 echo 'baseurl=http://nginx.org/packages/centos/$releasever/$basearch/'; \
		 echo 'gpgcheck=1'; \
		 echo 'enabled=1'; \
		 echo 'gpgkey=https://nginx.org/keys/nginx_signing.key'; \
		 echo 'module_hotfixes=true'; \
	   } | tee /etc/yum.repos.d/nginx.repo \
  && yum install -y \
     nginx \
     supervisor \
  && yum clean all && rm -rf /var/cache/yum \
  && usermod -aG $GID nginx \
  && rm -fr /etc/supervisord.conf /etc/supervisord.d \
  && mkdir -p /etc/supervisor/conf.d \
  && chmod +x /usr/local/bin/install-magento2.sh /usr/local/bin/stop-supervisor.sh

USER www-data

EXPOSE 80

CMD ["sudo", "supervisord", "-c", "/etc/supervisor/supervisord.conf"]

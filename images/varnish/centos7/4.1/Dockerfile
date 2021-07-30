FROM centos:7

ARG RPM_SCRIPT=https://packagecloud.io/install/repositories/varnishcache/varnish41/script.rpm.sh

COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

RUN set -eux \
  && curl -fsSLo- "${RPM_SCRIPT}" | bash \
  && yum install -y epel-release \
  && yum install -y \
     supervisor \
     varnish \
  && yum clean all && rm -rf /var/cache/yum

ENV VCL_CONFIG      /etc/varnish/default.vcl
ENV CACHE_SIZE      256m
ENV VARNISHD_PARAMS -p default_ttl=3600 -p default_grace=3600 \
    -p feature=+esi_ignore_https -p feature=+esi_disable_xml_check

COPY . /

RUN set -eux \
  && chmod +x /usr/local/bin/stop-supervisor.sh

ENV BACKEND_HOST    nginx
ENV BACKEND_PORT    80
ENV ACL_PURGE_HOST  0.0.0.0/0

EXPOSE 	80
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]

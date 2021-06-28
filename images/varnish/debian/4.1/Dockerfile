FROM debian:stretch-slim as builder

ARG DEB_SCRIPT=https://packagecloud.io/install/repositories/varnishcache/varnish41/script.deb.sh

ENV PKG_CONFIG_PATH /usr/lib/pkgconfig
ENV ACLOCAL_PATH    /usr/share/aclocal
ENV VARNISH_VERSION 4.1.11-1~stretch

RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
  && curl -fsSLo- "${DEB_SCRIPT}" | bash \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    build-essential \
    git \
    libtool \
    make \
    pkgconf \
    python3 \
    python-docutils \
    wget \
    unzip \
    libgetdns-dev \
    varnish=${VARNISH_VERSION} \
    varnish-dev=${VARNISH_VERSION} \
  && rm -rf /var/lib/apt/lists/* \
  && VARNISH_VERSION_SHORT="$(echo $VARNISH_VERSION | cut -f1,2 -d'.')" \
  && git clone --single-branch --branch "${VARNISH_VERSION_SHORT}" https://github.com/nigoroll/libvmod-dynamic.git /tmp/libvmod-dynamic \
  && cd "/tmp/libvmod-dynamic" \
  && chmod +x ./autogen.sh \
  && ./autogen.sh \
  && ./configure --prefix=/usr \
  && make -j "$(nproc)" \
  && make install

FROM debian:stretch-slim

COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/varnish/vmods/
COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/x86_64-linux-gnu/varnish/vmods/
COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

ARG DEB_SCRIPT=https://packagecloud.io/install/repositories/varnishcache/varnish41/script.deb.sh

ENV VARNISH_VERSION 4.1.11-1~stretch

RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
  && curl -fsSLo- "${DEB_SCRIPT}" | bash \
  && apt-get update && apt-get install -y --no-install-recommends \
    libgetdns1 \
    varnish=${VARNISH_VERSION} \
  && rm -rf /var/lib/apt/lists/* \
  && PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/sbin" ldconfig -n /usr/lib/varnish/vmods

ENV VCL_CONFIG      /etc/varnish/default.vcl
ENV CACHE_SIZE      256m
ENV VARNISHD_PARAMS -p default_ttl=3600 -p default_grace=3600 \
    -p feature=+esi_ignore_https -p feature=+esi_disable_xml_check

COPY default-dynamic.vcl /etc/varnish/default.vcl.template

# because of env substition the valid values are "true" or empty
ENV PROBE_DISABLED  true

ENV PROBE_URL       healthcheck.php
ENV BACKEND_HOST    nginx
ENV BACKEND_PORT    80
ENV ACL_PURGE_HOST  0.0.0.0/0

EXPOSE 	80
CMD envsubst < /etc/varnish/default.vcl.template > /etc/varnish/default.vcl \
    && varnishd -F -f $VCL_CONFIG -s malloc,$CACHE_SIZE $VARNISHD_PARAMS

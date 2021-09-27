ARG VARNISH_VERSION="4.1.11-1~stretch"
ARG VARNISH_REPO_VERSION="41"
ARG DEB_SCRIPT="https://packagecloud.io/install/repositories/varnishcache/varnish${VARNISH_REPO_VERSION}/script.deb.sh"

FROM debian:stretch-slim AS builder

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG DEB_SCRIPT
ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV ACLOCAL_PATH /usr/local/share/aclocal

RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends --allow-downgrades \
    ca-certificates \
    curl \
  && curl -fsSL "${DEB_SCRIPT}" | bash \
  && apt-get install -y --no-install-recommends --allow-downgrades \
    build-essential \
    autoconf \
    automake \
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
  && VARNISH_VERSION_SHORT="$(echo ${VARNISH_VERSION} | cut -f1,2 -d'.')" \
  && git clone --single-branch --branch "${VARNISH_VERSION_SHORT}" https://github.com/nigoroll/libvmod-dynamic.git /tmp/libvmod-dynamic \
  && cd /tmp/libvmod-dynamic \
  && chmod +x ./autogen.sh \
  && ./autogen.sh \
  && ./configure --prefix=/usr \
  && make -j "$(nproc)" \
  && make install
# Vmods do not exist for Varnish 4.1
#  && git clone --single-branch --branch "${VARNISH_VERSION_SHORT}" https://github.com/varnish/varnish-modules.git /tmp/varnish-modules \
#  && cd /tmp/varnish-modules \
#  && ./bootstrap \
#  && ./configure \
#  && make install

FROM debian:stretch-slim

COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/varnish/vmods/
COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/x86_64-linux-gnu/varnish/vmods/
COPY --from=hairyhenderson/gomplate:slim /gomplate /usr/local/bin/gomplate

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG DEB_SCRIPT

RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends --allow-downgrades \
    ca-certificates \
    curl \
  && curl -fsSL "${DEB_SCRIPT}" | bash \
  && apt-get install -y --no-install-recommends --allow-downgrades \
     libgetdns1 \
     supervisor \
     varnish=${VARNISH_VERSION} \
  && rm -rf /var/lib/apt/lists/* \
  && PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/sbin" ldconfig -n /usr/lib/varnish/vmods

ENV VCL_CONFIG      /etc/varnish/default.vcl
ENV CACHE_SIZE      256m
ENV VARNISHD_PARAMS -p default_ttl=3600 -p default_grace=3600 \
    -p feature=+esi_ignore_https -p feature=+esi_disable_xml_check \
    -p http_resp_hdr_len=65536 -p http_resp_size=98304

COPY . /

RUN set -eux \
  && chmod +x /usr/local/bin/stop-supervisor.sh

# because of env substition the valid values are "true" or empty
ENV PROBE_DISABLED  true

ENV PROBE_URL       healthcheck.php
ENV BACKEND_HOST    nginx
ENV BACKEND_PORT    80
ENV ACL_PURGE_HOST  0.0.0.0/0

EXPOSE 	80
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
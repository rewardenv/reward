ARG VARNISH_VERSION="6.6.2-1~buster"
ARG VARNISH_REPO_VERSION="66"
ARG VARNISH_MODULES_BRANCH="6.6"
ARG DEB_SCRIPT="https://packagecloud.io/install/repositories/varnishcache/varnish${VARNISH_REPO_VERSION}/script.deb.sh"

FROM golang:alpine AS builder-supervisor

RUN set -eux \
    && apk add --no-cache --update git gcc rust \
    && mkdir -p /src \
    && cd /src \
    && git clone https://github.com/ochinchina/supervisord.git . \
    && if [ "$(uname -m)" = "aarch64" ]; \
        then BUILD_ARCH="arm64"; \
        else BUILD_ARCH="amd64"; \
       fi \
    && CGO_ENABLED=0 GOOS=linux GOARCH=$BUILD_ARCH go build -a -ldflags "-linkmode external -extldflags -static" -o /usr/local/bin/supervisord github.com/ochinchina/supervisord

FROM debian:buster-slim AS builder

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG VARNISH_MODULES_BRANCH
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
    python3-docutils \
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
    && make install \
    && git clone --single-branch --branch "${VARNISH_MODULES_BRANCH}" https://github.com/varnish/varnish-modules.git /tmp/varnish-modules \
    && cd /tmp/varnish-modules \
    && ./bootstrap \
    && ./configure \
    && make install

FROM debian:buster-slim

COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/varnish/vmods/
COPY --from=builder-supervisor /usr/local/bin/supervisord /usr/bin/

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG VARNISH_MODULES_BRANCH
ARG DEB_SCRIPT
ENV VCL_CONFIG      /etc/varnish/default.vcl
ENV CACHE_SIZE      256m
ENV VARNISHD_PARAMS -p default_ttl=3600 -p default_grace=3600 \
    -p feature=+esi_ignore_https -p feature=+esi_disable_xml_check \
    -p http_req_size=98304 -p http_req_hdr_len=65536 \
    -p http_resp_size=98304 -p http_resp_hdr_len=65536 \
    -p workspace_backend=131072 -p workspace_client=131072
ENV PROBE_ENABLED         false
ENV PROBE_URL             health_check.php
ENV BACKEND_HOST          nginx
ENV BACKEND_PORT          80
ENV VMOD_DYNAMIC_ENABLED  true
ENV ACL_PURGE_HOST        0.0.0.0/0
ARG SUPERVISORD_VERSION=0.7.3
ENV SUPERVISORD_VERSION=$SUPERVISORD_VERSION

COPY rootfs/. /

RUN set -eux \
    && apt-get update && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends --allow-downgrades \
    ca-certificates \
    curl \
    && curl -fsSL "${DEB_SCRIPT}" | bash \
    && if [ "$(dpkg --print-architecture)" = "arm64" ]; \
        then BUILD_ARCH="arm64"; \
        else BUILD_ARCH="amd64"; \
    fi \
    && curl -fsSLo /usr/local/bin/gomplate \
      "https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-${BUILD_ARCH}" \
    && chmod +x /usr/local/bin/gomplate \
    && apt-get install -y --no-install-recommends --allow-downgrades \
    libgetdns10 \
    varnish=${VARNISH_VERSION} \
    && rm -rf /var/lib/apt/lists/* /var/log/apt \
    && PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/sbin" ldconfig -n /usr/lib/varnish/vmods \
    && chmod +x /usr/local/bin/stop-supervisor.sh

EXPOSE 	80

WORKDIR /etc/varnish

CMD ["supervisord", "-c", "/etc/supervisor/supervisord.conf"]

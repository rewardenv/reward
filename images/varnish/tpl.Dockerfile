# syntax=docker/dockerfile:1
{{- $VARNISH_VERSION := getenv "VARNISH_VERSION" "7.7.2-1" }}
{{- $VARNISH_REPO_VERSION := getenv "VARNISH_REPO_VERSION" "77" }}
{{- $VARNISH_MODULES_BRANCH := getenv "VARNISH_MODULES_BRANCH" "7.7" }}
{{- $DISTRO := getenv "DISTRO" "ubuntu" }}
{{- $DISTRO_RELEASE := getenv "DISTRO_RELEASE" "noble" }}
ARG VARNISH_VERSION={{ printf "%s~%s" $VARNISH_VERSION $DISTRO_RELEASE }}
ARG VARNISH_REPO_VERSION={{ $VARNISH_REPO_VERSION }}
ARG VARNISH_MODULES_BRANCH={{ $VARNISH_MODULES_BRANCH }}
ARG DEB_SCRIPT="https://packagecloud.io/install/repositories/varnishcache/varnish${VARNISH_REPO_VERSION}/script.deb.sh"

FROM {{ $DISTRO }}:{{ $DISTRO_RELEASE | strings.TrimSuffix "-1" }}{{- if eq $DISTRO "debian" }}-slim{{- end }} AS builder

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG VARNISH_MODULES_BRANCH
ARG DEB_SCRIPT
ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV ACLOCAL_PATH /usr/local/share/aclocal

WORKDIR /src/libvmod-dynamic

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN <<-EOF
    set -eux
    apt-get update
    apt-get upgrade -y
    apt-get install -y --no-install-recommends --allow-downgrades \
      ca-certificates \
      curl
    curl -fsSL "${DEB_SCRIPT}" | bash
    apt-get install -y --no-install-recommends --allow-downgrades \
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
      varnish-dev=${VARNISH_VERSION}
    VARNISH_VERSION_SHORT="$(echo ${VARNISH_VERSION} | cut -f1,2 -d'.')"
    git clone --single-branch --branch "${VARNISH_VERSION_SHORT}" https://github.com/nigoroll/libvmod-dynamic.git .
    chmod +x ./autogen.sh
    ./autogen.sh
    ./configure --prefix=/usr
    make -j "$(nproc)"
    make install
EOF

WORKDIR /src/varnish-modules

RUN <<-EOF
    set -eux
    git clone --single-branch --branch "${VARNISH_MODULES_BRANCH}" https://github.com/varnish/varnish-modules.git .
    ./bootstrap
    ./configure
    make install
EOF

FROM {{ $DISTRO }}:{{ $DISTRO_RELEASE | strings.TrimSuffix "-1" }}{{- if eq $DISTRO "debian" }}-slim{{- end }}

COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/varnish/vmods/
COPY --from=rewardenv/supervisord /usr/local/bin/supervisord /usr/bin/

ARG VARNISH_VERSION
ARG VARNISH_REPO_VERSION
ARG VARNISH_MODULES_BRANCH
ARG DEB_SCRIPT
ENV VCL_CONFIG      /etc/varnish/default.vcl
ENV VCL_TEMPLATE    default
ENV CACHE_TYPE      malloc
ENV CACHE_SIZE      256m
ENV VARNISHD_PARAMS -p default_ttl=3600 -p default_grace=3600 \
    -p feature=+esi_ignore_https -p feature=+esi_disable_xml_check \
    -p http_req_size=98304 -p http_req_hdr_len=65536 \
    -p http_resp_size=98304 -p http_resp_hdr_len=65536 \
    -p workspace_backend=131072 -p workspace_client=131072
ENV PROBE_ENABLED         false
ENV PROBE_URL             ""
ENV BACKEND_HOST          nginx
ENV BACKEND_PORT          80
ENV VMOD_DYNAMIC_ENABLED  true
ENV ACL_PURGE_HOST        0.0.0.0/0

COPY rootfs/. /

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN <<-EOF
    set -eux
    apt-get update
    apt-get upgrade -y
    apt-get install -y --no-install-recommends --allow-downgrades \
      ca-certificates \
      curl
    curl -fsSL "${DEB_SCRIPT}" | bash
    if [ "$(dpkg --print-architecture)" = "arm64" ]; \
      then BUILD_ARCH="arm64"; \
      else BUILD_ARCH="amd64"; \
    fi
    curl -fsSLo /usr/local/bin/gomplate \
      "https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-${BUILD_ARCH}"
    chmod +x /usr/local/bin/gomplate
    apt-get install -y --no-install-recommends --allow-downgrades \
      libgetdns10 \
      varnish=${VARNISH_VERSION}
    mkdir -p /var/lib/varnish/cache
    rm -rf /var/lib/apt/lists/* /var/log/apt
    PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/sbin" ldconfig -n /usr/lib/varnish/vmods
    chmod +x /docker-entrypoint.sh /usr/local/bin/stop-supervisor.sh
EOF

EXPOSE 80

WORKDIR /etc/varnish

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["supervisord", "-c", "/etc/supervisor/supervisord.conf"]

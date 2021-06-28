ARG VARNISH_BUILD_VERSION="6.6"
FROM varnish:${VARNISH_BUILD_VERSION} AS builder

# Adapted from https://knplabs.com/en/blog/how2tip-varnish-dynamic-backend-dns-resolution-in-a-docker-swarm-context
# Define env vars for VMOD build
ENV PKG_CONFIG_PATH /usr/local/lib/pkgconfig
ENV ACLOCAL_PATH /usr/local/share/aclocal
ENV VMOD_DYNAMIC_VERSION 2.2.1
RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends build-essential \
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
    varnish-dev="$VARNISH_VERSION" \
  && VARNISH_VERSION_SHORT="$(echo $VARNISH_VERSION | cut -f1,2 -d'.')" \
  && git clone --single-branch --branch "${VARNISH_VERSION_SHORT}" https://github.com/nigoroll/libvmod-dynamic.git /tmp/libvmod-dynamic \
  && cd "/tmp/libvmod-dynamic" \
  && chmod +x ./autogen.sh \
  && ./autogen.sh \
  && ./configure --prefix=/usr \
  && make -j "$(nproc)" \
  && make install

ARG VARNISH_BUILD_VERSION="6.6"
FROM varnish:${VARNISH_BUILD_VERSION}

COPY --from=builder /usr/lib/varnish/vmods/ /usr/lib/varnish/vmods/
COPY --from=janosmiko/envsubst /usr/bin/envsubst /usr/local/bin/envsubst

RUN set -eux \
  && apt-get update && apt-get upgrade -y \
  && apt-get install -y --no-install-recommends libgetdns10 \
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

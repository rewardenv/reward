# syntax=docker/dockerfile:1
# hadolint disable=DL4006
ARG WEBPROC_VERSION=0.4.0

FROM alpine:{{ getenv "IMAGE_TAG" "latest" }}
ARG WEBPROC_VERSION

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

RUN <<-EOF
    set -eux
    BUILD_ARCH="$(apk --print-arch)"
    if [ "${BUILD_ARCH}" = "aarch64" ]; \
      then WEBPROC_ARCH="arm64"; \
      else WEBPROC_ARCH="amd64"; \
    fi
    wget -q -O - "https://github.com/jpillora/webproc/releases/download/v${WEBPROC_VERSION}/webproc_${WEBPROC_VERSION}_linux_${WEBPROC_ARCH}.gz" \
      | gzip -d > /usr/bin/webproc
    chmod 0755 /usr/bin/webproc
    apk add --no-cache \
      dnsmasq
    mkdir -p /etc/default
    printf '%s\n%s\n' "ENABLED=1" "IGNORE_RESOLVCONF=yes" > /etc/default/dnsmasq
EOF

COPY dnsmasq.conf /etc/dnsmasq.conf

# The dhcp.leases files is put here, may want to mount as tmpfs
# FIXME Should this be preserved?
VOLUME ["/var/lib/misc"]

# Ports
#  80: Web interface
#  67: DHCP
#  53: DNS: normal on UDP, transfers on TCP
EXPOSE 80/tcp 67/udp 53/tcp 53/udp

ENTRYPOINT ["webproc", "-p", "80", "-c", "/etc/dnsmasq.conf", "--", "dnsmasq", "--no-daemon"]

HEALTHCHECK --interval=30s \
    --timeout=30s \
    --start-period=10s \
    --retries=3 \
    CMD ["pidof", "dnsmasq"]

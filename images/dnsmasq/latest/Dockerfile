ARG WEBPROC_VERSION=0.4.0

FROM alpine:latest
ARG WEBPROC_VERSION

RUN set -eux -o pipefail \
    && BUILD_ARCH="$(apk --print-arch)" \
    && if [ "${BUILD_ARCH}" = "aarch64" ]; \
        then WEBPROC_ARCH="arm64"; \
        else WEBPROC_ARCH="amd64"; \
    fi \
    && wget -O - "https://github.com/jpillora/webproc/releases/download/v${WEBPROC_VERSION}/webproc_${WEBPROC_VERSION}_linux_${WEBPROC_ARCH}.gz" \
      | gzip -d > /usr/bin/webproc \
	  && chmod 0755 /usr/bin/webproc \
    && apk add --update dnsmasq \
    && mkdir -p /etc/default \
    && echo -e "ENABLED=1\nIGNORE_RESOLVCONF=yes" > /etc/default/dnsmasq \

COPY dnsmasq.conf /etc/dnsmasq.conf

# The dhcp.leases files is put here, may want to mount as tmpfs
# XXX: should this be preserved?
VOLUME [ "/var/lib/misc" ]

# Ports:
#  80: Web interface
#  67: DHCP
#  53: DNS: normal on udp, transfers on tcp
EXPOSE 80/tcp 67/udp 53/tcp 53/udp

ENTRYPOINT ["webproc","-p","80","-c","/etc/dnsmasq.conf","--","dnsmasq","--no-daemon"]

HEALTHCHECK --interval=30s \
	--timeout=30s \
	--start-period=10s \
	--retries=3 \
	CMD [ "pidof", "dnsmasq" ]

# syntax=docker/dockerfile:1
FROM golang:{{ getenv "IMAGE_TAG" "alpine" }} as builder

WORKDIR /src/

RUN <<-EOF
    set -eux
    apk add --no-cache \
      git \
      gcc \
      rust
    git clone --single-branch --branch "{{ getenv "SUPERVISORD_VERSION" "v0.7.3" }}" https://github.com/ochinchina/supervisord.git .
    if [ "$(apk --print-arch)" = "aarch64" ]; \
      then BUILD_ARCH="arm64"; \
      else BUILD_ARCH="amd64"; \
    fi
    CGO_ENABLED=0 GOOS=linux GOARCH=$BUILD_ARCH go build -a -ldflags "-linkmode internal -extldflags -static" -o /usr/local/bin/supervisord github.com/ochinchina/supervisord
EOF

FROM scratch

COPY --from=builder /usr/local/bin/supervisord /usr/local/bin/supervisord

CMD ["/usr/local/bin/supervisord"]

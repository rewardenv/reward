# syntax=docker/dockerfile:1
FROM alpine:{{ getenv "IMAGE_TAG" "3.15" }}

COPY ./entry.sh /entry.sh

RUN <<-EOF
    set -eux
    apk add --no-cache \
      augeas \
      bash \
      git \
      openssh \
      rssh \
      rsync \
      shadow
    deluser $(getent passwd 33 | cut -d: -f1)
    delgroup $(getent group 33 | cut -d: -f1) 2>/dev/null || true
    mkdir -p ~root/.ssh /etc/authorized_keys && chmod 700 ~root/.ssh/
    augtool 'set /files/etc/ssh/sshd_config/AuthorizedKeysFile ".ssh/authorized_keys /etc/authorized_keys/%u"'
    echo -e "Port 22\n" >> /etc/ssh/sshd_config
    cp -a /etc/ssh /etc/ssh.cache
    chmod +x /entry.sh
EOF

EXPOSE 22

ENTRYPOINT ["/entry.sh"]

CMD ["/usr/sbin/sshd", "-D", "-e", "-f", "/etc/ssh/sshd_config"]

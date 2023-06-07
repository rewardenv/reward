## Reward Settings

During the installation of Reward a global configuration file will be created in the user's HOME directory,
called `~/.reward.yml`. It is possible to configure various settings of Reward in this configuration file, like what
container image should Reward use for global services, what image repo should it use, etc.

### Available settings

---

Set the logging level. Default is `info`.

- `log_level: info` - valid options: `trace`, `debug`, `info`, `warning`, `error`.

---

Enable debug logging (set log level to debug).

- `debug: false` - valid options: `false`, `true`

    Many options including this can be configured using an environment variable.

    ```bash
    DEBUG=true reward env up
    ```

---

Disable default common services. These services are enabled by default.

- `reward_portainer: true` - valid options: `false`, `true`
- `reward_dnsmasq: true` - valid options: `false`, `true`
- `reward_tunnel: true` - valid options: `false`, `true`
- `reward_mailhog: true` - valid options: `false`, `true`
- `reward_phpmyadmin: true` - valid options: `false`, `true`
- `reward_elastichq: true` - valid options: `false`, `true`

---

Enable additional common services. These services are disabled by default.

- `reward_adminer: false` - valid options: `false`, `true`

---

#### Service Container Settings

It's possible to change service container images using the following vars.

- `reward_traefik_image: "traefik"`
- `reward_portainer_image: "portainer/portainer-ce"`
- `reward_dnsmasq_image: "docker.io/rewardenv/dnsmasq"`
- `reward_mailhog_image: "docker.io/rewardenv/mailhog:1.0"`
- `reward_tunnel_image: "docker.io/rewardenv/sshd"`
- `reward_phpmyadmin_image: "phpmyadmin"`
- `reward_elastichq_image: "elastichq/elasticsearch-hq"`
- `reward_adminer_image: "dehy/adminer"`

---

By default, Traefik listens on `0.0.0.0`. To change this behaviour you can add the following line to the config
file and change the IP address to `127.0.0.1` to listen on localhost only.
It is also possible to change the listening ports.

- `reward_traefik_listen: "127.0.0.1"`
- `reward_traefik_http_port: "80"`
- `reward_traefik_https_port: "443"`

---

You can also add additional http and https ports on top of the defaults (80, 443). This is useful when you want to
expose a service on a different port than the default ones. See more info in
the [Open Ports](../configuration/open-additional-port.md) section.

- `reward_traefik_bind_additional_http_ports: []` - valid option example: `[8080, 8081]`
- `reward_traefik_bind_additional_https_ports: []` - valid option example: `[8443, 9443]`

---

By default, Reward makes it possible to resolve the environment's domain to the nginx container's IP address inside the
docker network. To disable this behaviour you add this line to the config file.

- `reward_resolve_domain_to_traefik: false`

---

By default, Reward redirects all http traffic to https. To disable this behaviour you add this line to the config file.

- `reward_traefik_allow_http: true`

---

It is possible to change DNSMasq listen address and ports. By default, DNSMasq listens on `127.0.0.1` and on port `53`.

- `reward_dnsmasq_listen: "0.0.0.0"`
- `reward_dnsmasq_tcp_port: "53"`
- `reward_dnsmasq_udp_port: "53"`

---

By default, only the UDP port 53 is exposed from the dnsmasq container. Sometimes it doesn't seem to be enough, and the
TCP port 53 has to be exposed as well. To do so enable the `reward_dnsmasq_bind_tcp` variable in the ~/.reward.yml file.

- `reward_dnsmasq_bind_tcp: false`
- `reward_dnsmasq_bind_udp: true`

---

It is possible to change Tunnel listen address and ports. By default, Tunnel listens on `0.0.0.0` and on port `2222`.

- `reward_tunnel_listen: "127.0.0.1"`
- `reward_tunnel_port: "2222"`

---

By default, Reward is not allowed to run commands as root. To disable this check you can add the following setting to
the
config.

- `reward_allow_superuser: true`

---

By default, Reward is going to use Mutagen sync for macOS and Windows. If you want to disable Mutagen you can set this
in Reward config.
Also, on Windows with WSL2 it's possible to use well performing direct mount from WSL2's drive. It is disabled by
default. To enable this functionality, disable syncing with the following line to the config.

- `reward_sync_enabled: false`

---

Previously Reward used CentOS 7 based images, now the defaults are debian based images.
Experimental images: `debian-bookworm`, `ubuntu-jammy`.

- `reward_docker_image_base: ubuntu-jammy`

---

By default Reward uses separated nginx + php-fpm containers. Enabling this setting will merge them to one "web"
container.

- `reward_single_web_container: true`

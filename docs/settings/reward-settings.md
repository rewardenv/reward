## Reward Settings

During the installation of Reward a global configuration file will be created in the user's HOME directory,
called `~/.reward.yml`. It is possible to configure various settings of Reward in this configuration file, like what
container image should Reward use for global services, what image repo should it use, etc.

### Available settings

Logging level, can be: trace, debug, info, warn, error

- `log_level: info` - valid options: `trace`, `debug`, `info`, `warning`, `error`.

Enable debugging (set log level to debug). Can be used as environment variable too.

eg: `DEBUG=true reward env up`

- `debug: false` - valid options: `false`, `true`

Disable default common services. These services are enabled by default.

- `reward_portainer: 1` - valid options: `0`, `1`
- `reward_dnsmasq: 1` - valid options: `0`, `1`
- `reward_tunnel: 1` - valid options: `0`, `1`
- `reward_mailhog: 1` - valid options: `0`, `1`
- `reward_phpmyadmin: 1` - valid options: `0`, `1`
- `reward_elastichq: 1` - valid options: `0`, `1`

Enable additional common services. These services are disabled by default.

- `reward_adminer: 0` - valid options: `0`, `1`

#### Service Container Settings

It's possible to change service container images using the following vars.

- `reward_traefik_image: "traefik"`
- `reward_portainer_image: "portainer/portainer-ce"`
- `reward_dnsmasq_image: "docker.io/rewardenv/dnsmasq"`
    - Reward < v0.2.33 uses "jpillora/dnsmasq" as the default dnsmasq image. Reward >= v0.2.34 uses the internally
      built "docker.io/rewardenv/dnsmasq"
- `reward_mailhog_image: "docker.io/rewardenv/mailhog:1.0"`
- `reward_tunnel_image: "docker.io/rewardenv/sshd"`
    - Reward < v0.2.33 uses "panubo/sshd:1.1.0" as the default dnsmasq image. Reward >= v0.2.34 uses the internally
      built "docker.io/rewardenv/sshd"
- `reward_phpmyadmin_image: "phpmyadmin"`
- `reward_elastichq_image: "elastichq/elasticsearch-hq"`
- `reward_adminer_image: "dehy/adminer"`

You can configure Traefik to bind additional http ports on top of the default port (80).

- `reward_traefik_bind_additional_http_ports: []` - valid option example: `[8080, 8081]`

You can configure Traefik to bind additional https ports on top of the default port (443).

- `reward_traefik_bind_additional_https_ports: []` - valid option example: `[8443, 9443]`

By default Reward makes it possible to resolve the environment's domain to the nginx container's IP address inside the
docker network. To disable this behaviour you add this line to the config file.

- `reward_resolve_domain_to_traefik: 0`

By default Reward is not allowed to run commands as root. To disable this check you can add the following setting to the
config.

- `reward_allow_superuser: 1`

By default Reward is going to use sync session for Windows. With WSL2 it's possible to use well performing direct mount
from WSL2's drive. It is disabled by default. To enable this functionality, add the following line to the config.

- `reward_wsl2_direct_mount: 1`

Previously Reward used CentOS 7 based images, now the defaults are debian based images and currently it only supports
debian. It's possible it will change in the future. Using this setting it's possible to change the Docker Image's base
image. Currently not working.

- `reward_docker_image_base: debian`

By default Reward uses separated nginx + php-fpm containers. Enabling this setting will merge them to one "web"
container.

- `reward_single_web_container: 1`

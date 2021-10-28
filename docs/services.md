## Global Services

After running `reward svc up` for the first time following installation, the following URLs can be used to interact with
the UIs for services Reward runs globally:

* [https://traefik.reward.test/](https://traefik.reward.test/)
* [https://portainer.reward.test/](https://portainer.reward.test/)
* [https://dnsmasq.reward.test/](https://dnsmasq.reward.test/)
* [https://mailhog.reward.test/](https://mailhog.reward.test/) or [https://mh.reward.test/](https://mh.reward.test/)
* [https://phpmyadmin.reward.test/](https://phpmyadmin.reward.test/)
  or [https://pma.reward.test/](https://pma.reward.test/)
* [https://elastichq.reward.test/](https://elastichq.reward.test/)
* optional: [https://adminer.reward.test/](https://adminer.reward.test/)

### Customizable Settings

When spinning up global services via `docker-compose` Reward uses `~/.reward` as the project directory
and `~/.reward.yml` or `~/.reward/.env` to function for overriding variables in the `docker-compose` configuration used
to deploy these services.

The following options are available (with default values indicated):

* `TRAEFIK_LISTEN=127.0.0.1` may be set to `0.0.0.0` for example to have Traefik accept connections from other devices
  on the local network.
* `REWARD_RESTART_POLICY=always` may be set to `no` to prevent Docker from restarting these service containers or any
  other
  valid [restart policy](https://docs.docker.com/config/containers/start-containers-automatically/#use-a-restart-policy)
  value.
* `REWARD_SERVICE_DOMAIN=reward.test` may be set to a domain of your choosing if so desired. Please note that this will
  not currently change network settings or alter `dnsmasq` configuration. Any TLD other than `test` will require DNS
  resolution be manually configured.

``` warning::
    Setting ``TRAEFIK_LISTEN=0.0.0.0`` can be quite useful in some cases, but be aware that causing Traefik to listen for requests publicly poses a security risk when on public WiFi or networks otherwise outside of your control.
```

After changing settings in `~/.reward.yml` or `~/.reward/.env`, please run `reward svc up` to apply.

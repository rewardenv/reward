## Expose Reward HTTP/HTTPS ports to the local network

It's possible to expose Reward HTTP/HTTPS ports to the local network. This is useful when you want to access the
environment from the local network.

To expose the HTTP and HTTPS ports of traefik to the local network, add the following line to `~/.reward.yml`:

```yaml
reward_traefik_listen: 0.0.0.0
```

When it's done, restart Traefik.

```shell
reward svc down
reward svc up
```

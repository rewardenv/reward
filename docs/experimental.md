## Experimental Features

### Debian-slim based Docker images

We have added experimental support for debian-slim based php images.
These are way smaller than the original (CentOS based) ones.

To enable using the debian-slim based php docker images append the following line to the `~/.reward.yml` config file:

```
reward_docker_image_base: debian
```

### WSL2 for Windows

[Using WSL2 for Windows with Reward](configuration/wsl2.md)

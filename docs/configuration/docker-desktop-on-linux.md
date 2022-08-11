## Docker Desktop on Linux

### Background

If you are using Docker Desktop on Linux your user is not going to utilize docker-engine on your system. Instead, in the
background Docker Desktop creates a Virtual Machine with Docker installed on it. After that Docker Desktop forwards the
socket of this machine's docker to your machine in your home (eg: `~/.docker/desktop/docker.sock`) and sets your
`docker` command (using docker contexts) to use this socket.

You can check these contexts using `docker context ls`.

But Reward still tries to use the default docker socket (`/run/docker.sock`).

### Solution

To change this behaviour **permanently**, you can open `~/.reward.yml` and add the following line:

```
docker_host: /home/_YOUR_USER_/.docker/desktop/docker.sock
```

TIP:
If you want to change the docker socket temporary you can set an environment variable before calling `reward`.

```
DOCKER_HOST=/home/_YOUR_USER_/.docker/desktop/docker.sock reward env ps
```

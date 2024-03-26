### Initializing a Custom Node Environment in a Subdomain

First, create our node "application":

`vim app.js`

```javascript
const http = require('http');

const hostname = '0.0.0.0';
const port = 3000;

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Hello World');
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
});
```

Next create Reward's `.env` file in the same directory and modify it as you wish.

```
reward env-init webapp --environment-type=local
```

With the first 3 variables you can configure the target
container, the shell and the user of `reward shell`.

With the next 3 variables you can configure the target of the synchronization (to which container, to which folder) and
you can disable sync if you want.

`vim .env`

```
REWARD_SHELL_CONTAINER=custom-container
REWARD_SHELL_COMMAND=sh
REWARD_SHELL_USER=node

REWARD_SYNC_ENABLED=true
REWARD_SYNC_CONTAINER=custom-container
REWARD_SYNC_PATH=/app

REWARD_ENV_NAME=webapp
REWARD_ENV_TYPE=local
REWARD_WEB_ROOT=/

TRAEFIK_DOMAIN=webapp.test
TRAEFIK_SUBDOMAIN=custom
TRAEFIK_EXTRA_HOSTS=

REWARD_DB=false
REWARD_ELASTICSEARCH=false
REWARD_OPENSEARCH=false
REWARD_VARNISH=false
REWARD_RABBITMQ=false
REWARD_REDIS=false
REWARD_MERCURE=false
```

Now we define the new environment using a custom go template.

Note: if you modify the container name in the service you
should modify it in the labels as well!

`vim .reward/reward-env.yml`

```
version: "3.5"
services:
  custom-container:
    hostname: "{{ .reward_env_name }}-node"
    build:
      context: .
      dockerfile: .reward/Dockerfile
    volumes:
      - appdata:/app
    extra_hosts:
      - {{ .traefik_domain }}:{{ default "0.0.0.0" .traefik_address }}
      - {{ default "app" .traefik_subdomain }}.{{ .traefik_domain }}:{{ default "0.0.0.0" .traefik_address}}
    labels:
      - traefik.enable=true
      - traefik.http.routers.custom.tls=true
      - traefik.http.routers.custom.rule=Host(`{{ default "app" .traefik_subdomain }}.{{ default "custom.test" .traefik_domain }}`)
      - traefik.http.services.custom.loadbalancer.server.port=3000
      - traefik.docker.network={{ .reward_env_name }}
      - dev.reward.container.name=custom-container
      - dev.reward.environment.name={{ .reward_env_name }}

volumes:
  appdata: {}

```

And finally create a custom Dockerfile

`vim .reward/Dockerfile`

```
FROM node:lts
WORKDIR /app
COPY . /app

ARG DOCKER_START_COMMAND="node app.js"
ENV DOCKER_START_COMMAND=$DOCKER_START_COMMAND

CMD ["sh", "-c", "while true; do ${DOCKER_START_COMMAND}; sleep 10; done"]
```

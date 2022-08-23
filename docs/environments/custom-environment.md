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

Next create Reward's `.env` file in the same directory. With the first 3 variables you can configure the target
container, the shell and the user of `reward shell`.

With the next 3 variables you can configure the target of the synchronization (to which container, to which folder) and
you can disable sync if you want.

`vim .env`

```
REWARD_SHELL_CONTAINER=custom-container
REWARD_SHELL_COMMAND=sh
REWARD_SHELL_USER=node
REWARD_SYNC_CONTAINER=custom-container
REWARD_SYNC_PATH=/app
REWARD_MUTAGEN_ENABLED=true

REWARD_ENV_NAME=webapp
REWARD_ENV_TYPE=local
REWARD_WEB_ROOT=/

TRAEFIK_DOMAIN=webapp.test
TRAEFIK_SUBDOMAIN=custom
TRAEFIK_EXTRA_HOSTS=

REWARD_DB=0
REWARD_ELASTICSEARCH=0
REWARD_OPENSEARCH=0
REWARD_VARNISH=0
REWARD_RABBITMQ=0
REWARD_REDIS=0
REWARD_MERCURE=0

ELASTICSEARCH_VERSION=7.12
MARIADB_VERSION=10.4
NODE_VERSION=10
PHP_VERSION=7.4
RABBITMQ_VERSION=3.8
REDIS_VERSION=6.0
VARNISH_VERSION=6.5
COMPOSER_VERSION=2

REWARD_SYNC_IGNORE=
REWARD_ALLURE=0
REWARD_SELENIUM=0
REWARD_SELENIUM_DEBUG=0
REWARD_BLACKFIRE=0
REWARD_SPLIT_SALES=0
REWARD_SPLIT_CHECKOUT=0
REWARD_TEST_DB=0
REWARD_MAGEPACK=0
BLACKFIRE_CLIENT_ID=
BLACKFIRE_CLIENT_TOKEN=
BLACKFIRE_SERVER_ID=
BLACKFIRE_SERVER_TOKEN=
XDEBUG_VERSION=

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
      - traefik.docker.network={{ .reward_env_name }}_default
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
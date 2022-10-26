## Expose ports from environment to the host machine

It's possible to expose ports from the environment to the host machine. This is useful when you want to access the
environment from the host machine
(eg. you run an application directly on the host machine and you want to access the Magento database.)

To expose ports from the environment to the host machine, add the following line to the project's `.env` file.

```shell
MYSQL_EXPOSE=true
REDIS_EXPOSE=true
OPENSEARCH_EXPOSE=true
ELASTICSEARCH_EXPOSE=true
RABBITMQ_EXPOSE=true
```

When it's done, restart the environment.

```shell
reward env down
reward env up
```

Please note that you cannot expose the same service twice. Eg. if you have `MYSQL_EXPOSE=true` in one environment and
you want to expose mysql from another environment that would cause a port conflict. You have to select a different port
using `MYSQL_EXPOSE_TARGET`.

The same applies for the rest of the services.

The default ports:

```bash
MYSQL_EXPOSE_TARGET=3306
REDIS_EXPOSE_TARGET=6379
OPENSEARCH_EXPOSE_TARGET=9200
ELASTICSEARCH_EXPOSE_TARGET=9200
RABBITMQ_EXPOSE_TARGET=5672
```
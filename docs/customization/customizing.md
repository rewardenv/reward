## Customizing An Environment

Further information on customizing or extending an environment is forthcoming. For now, this section is limited to very
simple and somewhat common customizations.

To configure your project with a non-default PHP version, add the following to the project's `.env` file and
run `reward env up` to re-create the affected containers:

    PHP_VERSION=7.4

The versions of MariaDB, Elasticsearch, Varnish, Redis, and NodeJS may also be similarly configured using variables in
the `.env` file:

* `MARIADB_VERSION`
* `ELASTICSEARCH_VERSION`
* `REDIS_VERSION`
* `VARNISH_VERSION`
* `RABBITMQ_VERSION`
* `NODE_VERSION`

The components in an environment can be skipped by disabling these variables in `.env` file:

* `REWARD_DB=0`
* `REWARD_REDIS=0`

### Customize a Reward environment to be able to reach another Reward environment

To make it possible to reach another Reward environment, the container DNS have to resolve the other project's domain
(eg.: `otherproject.test`) to Reward's Traefik container.

To do so add a space separated list of domains to the TRAEFIK_EXTRA_HOSTS variable in the .env file.

* `TRAEFIK_EXTRA_HOSTS="otherproject.test thirdproject.test"`


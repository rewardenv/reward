## Additional Packages

You can override the service containers and install additional packages to them using the following method.

### Override service

First, create an **override file** for the service you want to modify. For example, if you want to modify the `php-fpm`
service, create a file named `reward-env.yml` in the `.reward` directory.

Then, add the following snippet to the override file:

`vim .reward/reward-env.yml`

```yaml
version: "3.5"
services:
  php-fpm:
    build:
      context: .
      dockerfile: .reward/Dockerfile
```

### Dockerfile

Next, create the **custom Dockerfile**. The one below uses the default php-fpm 7.4 container for magento2 environments.

``` note::
    If you want to see what image is used by your current environment, run the following command:

        reward env config
```

`vim .reward/Dockerfile`

```Dockerfile
FROM docker.io/rewardenv/php-fpm:7.4-magento2

USER root

RUN apt-get update && apt-get install -y --no-install-recommends \
    telnet \
    && rm -rf /var/lib/apt/lists/*

USER www-data
```

Now we are ready, let's rebuild the environment.

```bash
reward env up --build
```

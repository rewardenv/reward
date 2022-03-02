## Composer

### Composer home

Composer home `~/.composer` directory is mounted and shared from your host system to the `php-fpm` container. This makes
possible to share Composer's cache, and your Composer auth configuration between environments.

### Change Composer version by environment

From Reward >0.2.0 it is possible to configure `COMPOSER_VERSION` in the .env file like this:

```
COMPOSER_VERSION=2
```

Default Composer versioning matrix by environment type:

| Environment Type | Composer Version |
|------------------|------------------|
| Generic PHP      | 2                |
| Magento 1        | 1                |
| Magento 2        | 2                |
| Laravel          | 2                |
| Shopware         | 2                |
| Symfony          | 2                |
| WordPress        | 2                |

### Change Composer version interactively inside the Reward Shell

```
$ reward shell
$ sudo alternatives --config composer
```

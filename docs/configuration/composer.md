## Composer

### Composer home

Composer home `~/.composer` directory is mounted and shared from your host system to the `php-fpm` container.
This makes possible to share Composer's cache, and your Composer auth configuration between environments.

### Change Composer version from 1.x to 2.x

```
$ reward shell
$ sudo alternatives --config composer
```

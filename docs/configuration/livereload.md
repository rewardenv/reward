## LiveReload Setup

LiveReload routing is currently supported on the `magento2`, `pwa-studio` and `shopware` environment types. Other
environment types may utilize LiveReload via per-project compose configurations to set up the routing for LiveReload JS
and WebSocket endpoints.

### Configuration for Magento 2

Magento 2 bundles an example grunt based server-side compilation workflow which includes LiveReload, and it works within
the Reward shell environment. In order to use this:

1. Rename or copy `Gruntfile.js.sample` file to `Gruntfile.js` in your project root.

2. Rename or copy `package.json.sample` file to `package.json` in your project root.

3. Run `npm install` to install the required NodeJS packages as defined in `package.json`.

4. Merge the following into your project's `app/etc/env.php` configuration file:

``` php
<?php
return [
    'system' => [
        'default' => [
            'design' => [
                'footer' => [
                    'absolute_footer' => '<script defer src="/livereload.js?port=443"></script>'
                ]
            ]
        ]
    ]
];
```

``` note::
    This can be accomplished via alternative means, the important part is the browser requesting ``/livereload.js?port=443`` when running the site on your local development environment.
```

5. Run `bin/magento app:config:import` to load merged configuration into the application and flush the
   cache `bin/magento cache:flush`.

**With the above configuration in place**, you'll first enter the FPM container via `reward shell` and then setup as
follows:

1. Clean and build the project theme using grunt:

```shell
$ grunt clean
$ grunt exec:blank
$ grunt less:blank
```

2. Thereafter, only a single command should be needed for daily development:

```shell
$ grunt watch
```

``` note::
    Grunt should be used within the php-fpm container entered via ``reward shell``
```

This setup will also be used to persist changes to your compiled CSS. When you run `grunt watch`, a LiveReload server
will be started on ports 35729 within the php-fpm container and Traefik will take care of proxying the JavaScript tag
and WebSocket requests to this listener.

On a working setup with `grunt watch` running within `reward shell` you should see something like the following in the
network inspector after reloading the project in a web browser.

![LiveReload Network Requests](screenshots/livereload.png)

### Configuration for Shopware Development Template

Shopware Hot Reload functionality requires to open additional ports on the same hostname and disable https redirects.

To configure the additional ports, you will have to update Traefik configuration.

Open Reward Global Configuration (default: `~/.reward.yml`) and add the following line:

```yaml
reward_traefik_bind_additional_https_ports: [ 9998, 9999 ]
reward_traefik_allow_http: true
```

When it's done, restart Traefik.

```shell
reward svc down
reward svc up
```

If you open [Traefik dashboard](https://traefik.reward.test), you should see the new ports in the entrypoints section.

![Traefik Additional HTTPS Port](screenshots/traefik-additional-port.png)

When it's done you have to configure this additional port on the PHP container as well.
Also, make sure to enable adding the required headers by Traefik.

Add the following line to the environment's `.env` file:

```
REWARD_HTTPS_PROXY_PORTS=9998,9999
REWARD_TRAEFIK_CUSTOM_HEADERS=hot-reload-mode=1,hot-reload-port=9999
```

And restart the environment.

```shell
reward env down
reward env up
```

Now if you start the Live Reload server, the requests will be proxied to it.

```
reward shell
./psh.phar storefront:hot-proxy
```

### Configuration for Shopware Production Template

Similarly to the development template, you have to configure Traefik to allow additional ports and disable https
redirects.

Open Reward Global Configuration (default: `~/.reward.yml`) and add the following line:

```yaml
reward_traefik_bind_additional_https_ports: [ 9998, 9999 ]
reward_traefik_allow_http: true
```

When it's done, restart Traefik.

```shell
reward svc down
reward svc up
```

Open the environment's `.env` file and add the following line:

```shell
REWARD_HTTPS_PROXY_PORTS=9998,9999
REWARD_TRAEFIK_CUSTOM_HEADERS=hot-reload-mode=1,hot-reload-port=9999
```

And restart the environment.

```shell
reward env down
reward env up
```

Now if you start the Live Reload server, the requests will be proxied to it.

```shell
reward shell
bin/build-storefront.sh
bin/watch-storefront.sh
```

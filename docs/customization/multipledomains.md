## Multiple Domains

If you need multiple domains configured for your project, Reward will now automatically route all sub-domains of the
configured `TRAEFIK_DOMAIN` (as given when running `env-init`) to the Varnish/Nginx containers provided there is not a
more specific rule such as for example `rabbitmq.exampleproject.com` which routes to the `rabbitmq` service for the
project.

Multiple top-level domains may also be setup by following the instructions below:

1. Sign certificates for your new domains:

       reward sign-certificate alternate1.test
       reward sign-certificate alternate2.test

2. **OPTIONAL**: The hosts from the `TRAEFIK_EXTRA_HOSTS` will be automatically configured and mapped to the webservers.

   It's possible to add additional host routing rules. Create a `.reward/reward-env.yml` file with the contents below (
   this will be additive to the docker-compose config Reward uses for the env, anything added here will be merged in,
   and you can see the complete config using `reward env config`):

    ```yaml
    version: "3.5"
    services:
      varnish:
        labels:
          - traefik.http.routers.{{.reward_env_name}}-varnish.rule=
              HostRegexp(`{subdomain:.+}.{{.traefik_domain}}`)
              || Host(`{{.traefik_domain}}`)
              || HostRegexp(`{subdomain:.+}.alternate1.test`)
              || Host(`alternate1.test`)
              || HostRegexp(`{subdomain:.+}.alternate2.test`)
              || Host(`alternate2.test`)
      nginx:
        labels:
          - traefik.http.routers.{{.reward_env_name}}-nginx.rule=
              HostRegexp(`{subdomain:.+}.{{.traefik_domain}}`)
              || Host(`{{.traefik_domain}}`)
              || HostRegexp(`{subdomain:.+}.alternate1.test`)
              || Host(`alternate1.test`)
              || HostRegexp(`{subdomain:.+}.alternate2.test`)
              || Host(`alternate2.test`)
    ```

3. Configure the application to handle traffic coming from each of these domains appropriately. An example on this for
   Magento 2 environments may be found below.

4. Run `reward env up -d` to update the containers, after which each of the URLs should work as expected.

    ``` note::
        If these alternate domains must be resolvable from within the FPM containers, you must also leverage ``extra_hosts`` to add each specific sub-domain to the ``/etc/hosts`` file of the container as dnsmasq is used only on the host machine, not inside the containers. This should look something like the following excerpt.

    ```

   From Reward >0.2.0 it is possible to add additional domains using `TRAEFIK_EXTRA_HOSTS` variable.

    ```
    TRAEFIK_EXTRA_HOSTS="alternate1.test sub1.alternate1.test alternate2.test and.test so.test on.test"
    ```

   Before Reward 0.2.0 you have to add these lines to the `.reward/reward-env.yml` file as you did in step 2.

    ```yaml
    version: "3.5"
    services:
      php-fpm:
       extra_hosts:
         - alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - sub1.alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - sub2.alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - alternate2.test:{{default "0.0.0.0" .traefik_address}}
         - sub1.alternate2.test:{{default "0.0.0.0" .traefik_address}}
         - sub2.alternate2.test:{{default "0.0.0.0" .traefik_address}}

      php-debug:
       extra_hosts:
         - alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - sub1.alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - sub2.alternate1.test:{{default "0.0.0.0" .traefik_address}}
         - alternate2.test:{{default "0.0.0.0" .traefik_address}}
         - sub1.alternate2.test:{{default "0.0.0.0" .traefik_address}}
         - sub2.alternate2.test:{{default "0.0.0.0" .traefik_address}}
    ```

### Magento Run Params (eg. Magento Multi Store)

There are two (and many more) ways to configure Magento run params (`MAGE_RUN_TYPE`, `MAGE_RUN_CODE`).

* Nginx mappings
* Composer autoload

#### Nginx mappings

Nginx makes it possible to map values to variables based on other variable's values.

Example:
Add the following file to you project folder `./.reward/nginx/http-maps.conf` with the content below. Don't forget to
restart your nginx container. `reward env restart -- nginx`

* if the `$http_host` value is `sub.example.test`, nginx will map value `store_code_1` to `$MAGE_RUN_CODE`.
* if the `$http_host` value is `sub.example.test`, nginx will map value `store` to `$MAGE_RUN_TYPE`.

```
map $http_host $MAGE_RUN_CODE {
    example.test            default;
    sub.example.test        store_code_1;
    website.example.test    another_run_code;
    default                 default;
}
map $http_host $MAGE_RUN_TYPE {
    example.test            store;
    sub.example.test        store;
    website.example.test    website;
    default                 store;
}
```

#### Composer autoload php file

When multiple domains are being used to load different stores or websites on Magento 2, the following configuration
should be defined in order to set run codes and types as needed.

1. Add a file at `app/etc/stores.php` with the following contents:

    ```php
    <?php

    use \Magento\Store\Model\StoreManager;
    $serverName = isset($_SERVER['HTTP_HOST']) ? $_SERVER['HTTP_HOST'] : null;

    switch ($serverName) {
        case 'domain1.exampleproject.test':
            $runCode = 'examplecode1';
            $runType = 'website';
            break;
        case 'domain2.exampleproject.test':
            $runCode = 'examplecode2';
            $runType = 'website';
            break;
        default:
            return;
    }

    if ((!isset($_SERVER[StoreManager::PARAM_RUN_TYPE])
            || !$_SERVER[StoreManager::PARAM_RUN_TYPE])
        && (!isset($_SERVER[StoreManager::PARAM_RUN_CODE])
            || !$_SERVER[StoreManager::PARAM_RUN_CODE])
    ) {
        $_SERVER[StoreManager::PARAM_RUN_CODE] = $runCode;
        $_SERVER[StoreManager::PARAM_RUN_TYPE] = $runType;
    }
    ```

    ``` note::
        The above example will not alter production site behavior given the default is to return should the ``HTTP_HOST`` value not match one of the defined ``case`` statements. This is desired as some hosting environments define run codes and types in an Nginx mapping. One may add production host names to the switch block should it be desired to use the same site switching mechanism across all environments.
    ```

2. Then in `composer.json` add the file created in the previous step to the list of files which are automatically loaded
   by composer on each web request:

    ```json
    {
        "autoload": {
            "files": [
                "app/etc/stores.php"
            ]
        }
    }
    ```

    ``` note::
        This is similar to using `magento-vars.php` on Magento Commerce Cloud, but using composer to load the file rather than relying on Commerce Cloud magic: https://devdocs.magento.com/guides/v2.3/cloud/project/project-multi-sites.html
    ```

3. After editing the `composer.json` regenerate the autoload configuration:

    ```bash
    composer dump-autoload
    ```

### Initializing PWA Studio

#### Prerequisites in Magento

From PWA Studio 12.1 you have to install PWA Studio metapackage in your Magento instance.

If you run Magento as a Reward Environment, run the following commands in the PHP container:

```shell
composer require magento/pwa

# If you want to install sample data
composer require magento/venia-sample-data

php bin/magento setup:upgrade
php bin/magento setup:di:compile
php bin/magento setup:static-content:deploy -f
php bin/magento indexer:reindex
php bin/magento cache:clean
```

#### Running Default PWA Studio

1. Clone your project and initialize Reward.

    ``` shell
    $ git clone https://github.com/magento/pwa-studio.git ~/Sites/pwa-studio
    $ cd ~/Sites/pwa-studio
    $ reward env-init pwa-studio --environment-type=pwa-studio
    ```

2. Sign a certificate for your project

    ```
    $ reward sign-certificate pwa-studio.test
    ```


3. Fill up the `.env` file with samples and change some settings

    ``` shell
    $ cat docker/.env.docker.dev >> .env
    ```

    ``` shell
    NODE_VERSION=14
    DEV_SERVER_HOST=0.0.0.0
    DEV_SERVER_PORT=8000
    MAGENTO_BACKEND_EDITION=MOS
    ```

4. Update `package.json` and add these to the `scripts` part.

   Note: it seems like PWA Studio 12.x ignores the `DEV_SERVER_PORT` variable, so we override it from the command line.

    ```
    "watch": "yarn watch:venia --disable-host-check --public pwa-studio.test --port 8000",
    "start": "yarn stage:venia --disable-host-check --public pwa-studio.test --port 8000"
    
    # you can use this script if you are familiar with jq
    DOMAIN="pwa-studio.test"
    cat package.json | jq --arg domain "$DOMAIN" -Mr '. * {scripts:{watch: ("yarn watch:venia --public " + $domain + " --disable-host-check --port 8000"), start: ("yarn stage:venia --public " + $domain + " --disable-host-check --port 8000")}}' | tee package.json
    ```

    ``` note::
        We have to add both --disable-host-check (to skip host header verification) and --public (to let webpack dev 
        server know it is behind a proxy and it shouldn't add custom port to it's callback URLs).
   
        https://webpack.js.org/configuration/dev-server/#devserverpublic
    ```

5. Bring up the environment

    ```
    $ reward env up
    ```

6. Install its dependencies

    ```
    $ reward shell
    $ yarn install
    ```

7. Restart the PWA container

    ```
    $ reward env restart
    ```

8. Optional: if you'd like to run the project in Developer/Production mode, add the following line to your `.env` file

    ```
    # Developer Mode (default)
    DOCKER_START_COMMAND="yarn watch"

    # Production Mode
    DOCKER_START_COMMAND="yarn start"
    ```

##### Reach a Reward Magento backend environment

If your PWA's Magento backend is also running on your computer as a Reward environment, you will have to configure the
PWA container to resolve the Magento DNS to the Reward Traefik container.

To do so add a space separated list of domains to the `TRAEFIK_EXTRA_HOSTS` variable in the .env file.

* `TRAEFIK_EXTRA_HOSTS="otherproject.test thirdproject.test"`

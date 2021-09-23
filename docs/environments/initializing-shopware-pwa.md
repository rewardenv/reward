### Initializing Shopware PWA

#### Running Shopware PWA

This guide is based on the [official documentation](https://shopware-pwa-docs.vuestorefront.io/landing/getting-started/local-environment.html#let-s-do-it) of Shopware PWA.

(0.) Install SwagShopwarePwa Plugin

   1. Download a plugin packed in a zip file from github: [master version](https://github.com/elkmod/SwagShopwarePwa/archive/master.zip)
   
   2. Log in to the admin panel at [https://your-awesome-shopware-project.test/admin](https://your-awesome-shopware-project.test/admin)
   
   3. Go to Setting > System > [Plugins](https://your-awesome-shopware-project.test/admin#/sw/plugin/index/list) and click Upload plugin button.
   
   4. When the plugin is uploaded - just install and activate it. That's all. Shopware 6 is shopware-pwa ready now!

   Or to do it programmatically run the following commands inside your Shopware Reward Shell:
   ```
      wget https://github.com/elkmod/SwagShopwarePwa/archive/master.zip -O ~/swpwa.zip
      php bin/console plugin:zip-import ~/swpwa.zip
   ```

---

1. Initialize Reward.

    ``` shell
    $ mkdir ~/Sites/shopware-pwa
    $ cd ~/Sites/shopware-pwa
    $ reward env-init shopware-pwa --environment-type=shopware-pwa
    ```

2. Sign a certificate for your project

    ```
    $ reward sign-certificate shopware-pwa.test
    ```

3. Bring up the environment

    ```
    $ reward env up
    ```

4. Create a regular PWA Project

    ```
    $ reward shell
    $ npx @shopware-pwa/cli@canary init
    ```

5. Configure backend in `shopware-pwa.config.js`

6. Update `package.json` and add these to the `scripts` part

    ```
    "start": "yarn && yarn build --ci && node scripts/init.js --disable-host-check --public shopware-pwa.test"
    
    # you can use this script if you are familiar with jq
    DOMAIN="shopware-pwa.test"
    cat package.json | jq --arg domain "$DOMAIN" -Mr '. * {scripts:{start: ("yarn && yarn build --ci && node scripts/init.js --public " + $domain + " --disable-host-check")}}' | tee package.json
    ```
   
    ``` ...note::
        We have to add both --disable-host-check (to skip host header verification) and --public (to let webpack dev 
        server know it is behind a proxy and it shouldn't add custom port to it's callback URLs).
   
        https://webpack.js.org/configuration/dev-server/#devserverpublic
    ```

7. Install dependencies

    ```
    $ reward shell
    $ yarn
    ```

8. Restart the PWA container

    ```
    $ reward env restart
    ```

9. Optional: if you'd like to run the project in Developer/Production mode, add the following line to your `.env` file

    ```
    # Developer Mode (default)
    DOCKER_START_COMMAND="yarn dev"

    # Production Mode
    DOCKER_START_COMMAND="yarn start"
    ```

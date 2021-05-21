### Initializing PWA Studio

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
    NODE_VERSION=12
    DEV_SERVER_HOST=0.0.0.0
    DEV_SERVER_PORT=8000
    ```

4. Update `package.json` and add these to the scripts part

    ```
    "watch": "yarn watch:venia --disable-host-check",
    "start": " yarn stage:venia --disable-host-check"
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

If your PWA's Magento backend is also running on your computer as a Reward environment, you will have to configure
the PWA container to resolve the Magento DNS to the Reward Traefik container.

To do so add a space separated list of domains to the `TRAEFIK_EXTRA_HOSTS` variable in the .env file.
* `TRAEFIK_EXTRA_HOSTS="otherproject.test thirdproject.test"`

### Initializing Shopware

#### Initializing an Empty Shopware Project

1. Clone the code and initialize a Laravel Reward environment

    ``` shell
    $ git clone https://github.com/shopware/development.git ~/Sites/your-awesome-shopware-project/webroot
    $ ~/Sites/your-awesome-shopware-project
    $ reward env-init your-awesome-shopware-project --environment-type=shopware
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    $ reward sign-certificate your-awesome-shopware-project.test
    ```

3. Change Reward WEBROOT in the `.env` file and bring up the Reward environment

    ``` shell
    $ sed -i.old -e 's#^REWARD_WEB_ROOT.*#REWARD_WEB_ROOT=/webroot#' .env

    $ reward env up
    ```

4. As composer2 is required for Shopware make it default and install shopware

    ``` shell
    $ reward shell

    $ sudo mv -f /usr/bin/composer2 /usr/bin/composer

    $ echo $'const:\n  APP_ENV: "dev"\n  APP_URL: "https://your-awesome-shopware-project.test"\n  DB_HOST: "db"\n  DB_NAME: "shopware"\n  DB_USER: "app"\n  DB_PASSWORD: "app"' > .psh.yaml.override

    $ ./psh.phar install
    ```

    ``` ...note::
        Now you can reach the project on the following url:

        https://your-awesome-shopware-project.test
    ```

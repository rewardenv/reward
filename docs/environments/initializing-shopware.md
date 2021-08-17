### Initializing Shopware

#### Initializing an Empty Shopware Project

1. Clone the code and initialize a Reward Shopware environment

    ``` shell
    $ git clone https://github.com/shopware/development.git -b v6.4.3.0 ~/Sites/your-awesome-shopware-project/webroot
    $ cd ~/Sites/your-awesome-shopware-project
    $ reward env-init your-awesome-shopware-project --environment-type=shopware
    ```

    ``` ...note::
        In this example the shopware code will live in the $PROJECT/webroot directory.
        If you'd like to use a different directory, change `REWARD_WEB_ROOT` environment variable in `.env` file.
    ```

3. Sign a new certificate for your dev domain

    ``` shell
    $ reward sign-certificate your-awesome-shopware-project.test
    ```

4. Bring up the Reward environment

    ``` shell
    $ reward env up
    ```

5. Install Shopware

    ``` shell
    $ reward shell

    $ echo $'const:\n  APP_ENV: "dev"\n  APP_URL: "https://your-awesome-shopware-project.test"\n  DB_HOST: "db"\n  DB_NAME: "shopware"\n  DB_USER: "app"\n  DB_PASSWORD: "app"' > .psh.yaml.override

    $ ./psh.phar install
    ```

    ``` ...note::
        Now you can reach the project on the following url:

        https://your-awesome-shopware-project.test
   
        Or the admin dashboard on
        https://your-awesome-shopware-project.test/admin
   
        user: admin
        password: shopware
    ```

### Initializing Shopware

#### Initializing an Empty Shopware Development Project

1. Clone the code and initialize a Reward Shopware environment

    ``` shell
    git clone https://github.com/shopware/development.git -b v6.4.3.0 ~/Sites/your-awesome-shopware-project/webroot
    cd ~/Sites/your-awesome-shopware-project
    reward env-init your-awesome-shopware-project --environment-type=shopware
    ```

    ``` note::
        In this example the shopware code will live in the $PROJECT/webroot directory.
        If you'd like to use a different directory, change `REWARD_WEB_ROOT` environment variable in `.env` file.
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    reward sign-certificate your-awesome-shopware-project.test
    ```

3. Bring up the Reward environment

    ``` shell
    reward env up
    ```

4. Install Shopware

   If you use different domain, make sure to **update the `APP_URL`** in the `.psh.yaml.override` file.

    ``` shell
    reward shell

    echo $'const:\n  APP_ENV: "dev"\n  APP_URL: "https://your-awesome-shopware-project.test"\n  DB_HOST: "db"\n  DB_NAME: "shopware"\n  DB_USER: "app"\n  DB_PASSWORD: "app"' > .psh.yaml.override

    # Windows only: give run permissions for the necessary files
    chmod +x psh.phar bin/console bin/setup

    ./psh.phar install
    ```

``` note::
    Now you can reach the project on the following url:

    https://your-awesome-shopware-project.test
   
    Or the admin dashboard on
    https://your-awesome-shopware-project.test/admin
   
    user: admin
    password: shopware
```

#### Initializing an Empty Shopware Production Project

1. Clone the code and initialize a Reward Shopware Production environment

    ``` shell
    git clone https://github.com/shopware/production.git -b v6.4.18.0 ~/Sites/your-awesome-shopware-project/webroot
    cd ~/Sites/your-awesome-shopware-project
    reward env-init your-awesome-shopware-project --environment-type=shopware
    ```

    ``` note::
        In this example the shopware code will live in the $PROJECT/webroot directory.
        If you'd like to use a different directory, change `REWARD_WEB_ROOT` environment variable in `.env` file.
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    reward sign-certificate your-awesome-shopware-project.test
    ```

3. Make some fixes in the .env file

    ``` shell
    # Enable opensearch
    # REWARD_OPENSEARCH=true
    sed -i.bak "s/^REWARD_OPENSEARCH=.*/REWARD_OPENSEARCH=true/g" .env

    # Change node version
    # NODE_VERSION=16
    sed -i.bak "s/^NODE_VERSION=.*/NODE_VERSION=16/g" .env

    ```

4. Bring up the Reward environment

    ``` shell
    reward env up
    ```

5. Install Shopware

    ``` shell
    reward shell

    # Issue with composer version 2.5
    # https://github.com/shopware/production/issues/168
    # Rollback to a previous composer version
    sudo composer self-update 2.4.4

    composer install --no-interaction

    # Configure Shopware
    bin/console system:setup --no-interaction --app-env dev --app-url https://your-awesome-shopware-project.test --database-url mysql://app:app@db:3306/shopware --es-enabled=1 --es-hosts=opensearch:9200 --es-indexing-enabled=1 --cdn-strategy=physical_filename --mailer-url=native://default

    # Install Shopware
    bin/console system:install --no-interaction --drop-database --create-database --basic-setup

    # Dump plugin and theme settings
    export CI=1
    bin/console bundle:dump
    bin/console theme:dump

    # Build storefront
    export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1
    bin/build.sh

    # Do the migrations and install assets
    bin/console system:update:finish --no-interaction
    ```

``` note::
    Now you can reach the project on the following url:

    https://your-awesome-shopware-project.test
   
    Or the admin dashboard on
    https://your-awesome-shopware-project.test/admin
   
    user: admin
    password: shopware
```

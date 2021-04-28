### Initializing Laravel

#### Initializing an Empty Laravel Project

1. Create an empty directory and a Reward Laravel environment

    ``` shell
    $ mkdir ~/Sites/your-awesome-laravel-project
    $ reward env-init your-awesome-laravel-project --environment-type=laravel
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    $ reward sign-certificate your-awesome-laravel-project.test
    ```

3. Bring up the Reward environment

    ``` shell
    $ reward env up
    ```

4. Create the laravel project in the php container

    ``` shell
    $ reward shell

    $ composer create-project --no-install --no-scripts --prefer-dist \
        laravel/laravel /tmp/laravel-tmp
    $ rsync -au --remove-source-files /tmp/laravel-tmp/ /var/www/html/
    ```

5. Install the composer packages and create an app key

    ``` shell
    $ reward shell

    $ composer install

    $ php artisan key:generate --show

    # Add the previously generated key to your .env file with your favourite editor
    # It should be added using the following format
    APP_KEY=base64:yourkey

    # Import the new .env content to the runtime environment variables
    $ source .env

    # Generate your config cache
    $ php artisan config:cache
    ```

    ``` ...note::
        Now you can reach the project on the following url:

        https://your-awesome-laravel-project.test
    ```

#### Initializing a Laravel Backpack Demo Project

1. Clone the code and initialize a Laravel Reward environment

    ``` shell
    $ git clone https://github.com/Laravel-Backpack/demo.git ~/Sites/demo
    $ cd ~/Sites/demo
    $ reward env-init demo --environment-type=laravel
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    $ reward sign-certificate demo.test
    ```

3. Bring up the Reward environment

    ``` shell
    $ reward env up
    ```

4. Install the composer packages and intialize the database

    ``` shell
    $ reward shell

    $ composer install
    $ php artisan key:generate --show

    # Add the previously generated key to your .env file with your favourite editor
    # It should be added using the following format
    APP_KEY=base64:yourkey

    # Import the new .env content to the runtime environment variables
    $ source .env

    $ php artisan migrate
    $ php artisan db:seed
    ```

    ``` ...note::
        Now you can reach the project on the following url:

        https://demo.test

        The default admin credentials are the following:

        .. code::

            user: admin@example.com
            pass: admin
    ```

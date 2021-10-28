### Initializing Symfony

#### Initializing an Empty Symfony Project

1. Create an empty directory and a Reward Symfony environment

    ``` shell
    $ mkdir ~/Sites/your-awesome-symfony-project
    $ reward env-init your-awesome-symfony-project --environment-type=symfony
    ```

2. Sign a new certificate for your dev domain

    ``` shell
    $ reward sign-certificate your-awesome-symfony-project.test
    ```

3. Bring up the Reward environment

    ``` shell
    $ reward env up
    ```

4. Create the symfony project in the php container

    ``` shell
    $ reward shell

    $ composer create-project --no-install --no-interaction \
        symfony/website-skeleton /tmp/symfony-tmp
    $ rsync -au --remove-source-files /tmp/symfony-tmp/ /var/www/html/
    ```

5. Install the composer packages

    ``` shell
    $ reward shell

    $ composer install
    ```

    ``` note::
        Now you can reach the project on the following url:

        https://your-awesome-symfony-project.test
    ```


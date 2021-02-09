### Initializing Wordpress

#### Empty Wordpress Project with bootstrap command

It's pretty easy to bootstrap a Wordpress project using Reward.

1. Create a new environment in an empty directory:

    ``` shell
    $ mkdir ~/Sites/your-awesome-wordpress-project
    $ cd ~/Sites/your-awesome-wordpress-project
    $ reward env-init your-awesome-wordpress-project
    ```

2. Provision the environment using Reward's bootstrap command:
    ``` shell
    $ reward bootstrap
    ```

    This is going to create a new wordpress installation by downloading wordpress and configuring wp-config.php.

#### Importing a Wordpress Project and initializing with bootstrap command

1. Clone your project and initialize Reward.

    ``` shell
    $ git clone git://github.com/your-user/your-awesome-wordpress-project.git ~/Sites/your-awesome-wordpress-project
    $ cd ~/Sites/your-awesome-wordpress-project
    $ reward env-init your-awesome-wordpress-project
    ```

2. Before running the bootstrap command, you should import the Magento database to the DB Container. To do so, first start the DB container:

    ``` shell
    $ reward env up -- db
    ```

3. Import the database.

    ``` shell
    $ reward db import < /path/to/db-dump-wordpress.sql
    ```

4. When the import is done, you can run the bootstrap.

    ```
    $ reward bootstrap
    ```


### Initializing Magento 1

#### Importing a Magento 1 Project and initializing with bootstrap command

1. Clone your project and initialize Reward.

    ``` shell
    $ git clone git://github.com/your-user/your-awesome-m1-project.git ~/Sites/your-awesome-m1-project
    $ reward env-init your-awesome-m2-project --environment-type=magento1
    ```

2. Before running the bootstrap command, you should import the Magento database to the DB Container. To do so, first start the DB container:

    ``` shell
    $ reward env up -- db
    ```

3. Import the database.

    ``` shell
    $ reward db import < /path/to/db-dump-for-magento1.sql
    ```

4. When the import is done, you can run the bootstrap.

    ```
    $ reward bootstrap
    ```

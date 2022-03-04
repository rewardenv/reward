### Initializing Magento 2

#### Empty Magento 2 Project with bootstrap command

It's pretty easy to bootstrap a Magento 2 project using Reward.

1. Create a new environment in an empty directory:

    ``` shell
    $ mkdir ~/Sites/your-awesome-m2-project
    $ cd ~/Sites/your-awesome-m2-project
    $ reward env-init your-awesome-m2-project
    ```

2. If this is the first time you install a magento project, you have to configure your composer keys:

    ``` shell
    $ reward env up -- php-fpm
    $ reward shell
    $ composer config -a -g http-basic.repo.magento.com MAGENTO_PUBLIC_KEY MAGENTO_PRIVATE_KEY
    ```

3. Provision the environment using Reward's bootstrap command:
    ``` shell
    $ reward bootstrap
    ```

   This is going to create a new Magento 2 installation using Composer (if your Composer Magento Repo Key is not set it
   will ask for it). After the vendor installation it's going to install Magento (configure the env.php), configure the
   local domains and configure an admin user.

   It is also possible to install the project and dump Magento sample data with one step.

    ``` shell
    $ reward bootstrap --with-sampledata
    ```

   It is possible to change the DB prefix with the following command.

   ``` shell
   $ reward bootstrap --db-prefix=<somestring>
   ```


#### Importing a Magento 2 Project and initializing with bootstrap command

1. Clone your project and initialize Reward.

    ``` shell
    $ git clone git://github.com/your-user/your-awesome-m2-project.git ~/Sites/your-awesome-m2-project
    $ cd ~/Sites/your-awesome-m2-project
    $ reward env-init your-awesome-m2-project
    ```

2. Before running the bootstrap command, you should import the Magento database to the DB Container. To do so, first
   start the DB container:

    ``` shell
    $ reward env up -- db
    ```

3. Import the database.

    ``` shell
    $ reward db import < /path/to/db-dump.sql
    ```

4. If this is the first time you install a magento project, you have to configure your composer keys:

    ``` shell
    $ reward env up -- php-fpm
    $ reward shell
    $ composer config -a -g http-basic.repo.magento.com MAGENTO_PUBLIC_KEY MAGENTO_PRIVATE_KEY
    ```

5. When the import is done, you can run the bootstrap.

    ```
    $ reward bootstrap
    ```

    ``` note::
        If you already installed your composer component, it's also possible to skip the composer install steps:

        .. code::

            $ reward bootstrap --skip-composer-install
    ```

#### Initializing A Magento 2 Environment Manually

The below example demonstrates the from-scratch setup of the Magento 2 application for local development. A similar
process can easily be used to configure an environment of any other type. This assumes that Reward has been previously
started via `reward svc up` as part of the installation procedure.

1. Create a new directory on your host machine at the location of your choice and then jump into the new directory to
   get started:

    ``` shell
    $ mkdir -p ~/Sites/your-awesome-m2-project
    $ cd ~/Sites/your-awesome-m2-project
    ```

2. From the root of your new project directory, run `env-init` to create the `.env` file with configuration needed for
   Reward and Docker to work with the project.

    ``` shell
    $ reward env-init your-awesome-m2-project --environment-type magento2
    ```

   The result of this command is a `.env` file in the project root (tip: commit this to your VCS to share the
   configuration with other team members) having the following contents:

    ``` shell
    REWARD_ENV_NAME=your-awesome-m2-project
    REWARD_ENV_TYPE=magento2
    REWARD_WEB_ROOT=/

    TRAEFIK_DOMAIN=your-awesome-m2-project.test
    TRAEFIK_SUBDOMAIN=
    TRAEFIK_EXTRA_HOSTS=

    REWARD_DB=1
    REWARD_ELASTICSEARCH=1
    REWARD_VARNISH=1
    REWARD_RABBITMQ=1
    REWARD_REDIS=1

    REWARD_SYNC_IGNORE=

    ELASTICSEARCH_VERSION=7.12
    MARIADB_VERSION=10.4
    NODE_VERSION=10
    PHP_VERSION=7.3
    RABBITMQ_VERSION=3.8
    REDIS_VERSION=6.0
    VARNISH_VERSION=6.5

    REWARD_ALLURE=0
    REWARD_SELENIUM=0
    REWARD_SELENIUM_DEBUG=0
    REWARD_BLACKFIRE=0
    REWARD_SPLIT_SALES=0
    REWARD_SPLIT_CHECKOUT=0
    REWARD_TEST_DB=0
    REWARD_MAGEPACK=0

    BLACKFIRE_CLIENT_ID=
    BLACKFIRE_CLIENT_TOKEN=
    BLACKFIRE_SERVER_ID=
    BLACKFIRE_SERVER_TOKEN=
    ```

3. Sign an SSL certificate for use with the project (the input here should match the value of `TRAEFIK_DOMAIN` in the
   above `.env` example file):

    ``` bash
    $ reward sign-certificate your-awesome-m2-project.test
    ```

4. Next you'll want to start the project environment:

    ``` shell
    $ reward env up
    ```

    ``` warning::
        If you encounter an error about ``Mounts denied``, follow the instructions in the error message and run ``reward env up -d`` again.
    ```

5. Drop into a shell within the project environment. Commands following this step in the setup procedure will be run
   from within the `php-fpm` docker container this launches you into:

    ``` shell
    $ reward shell
    ```

6. Configure global Magento Marketplace credentials

    ``` shell
    $ composer global config http-basic.repo.magento.com <username> <password>
    ```

    ``` note::
        To locate your authentication keys for Magento 2 repository, `reference DevDocs <https://devdocs.magento.com/guides/v2.    3/install-gde/prereq/connect-auth.html>`_.

        If you have previously configured global credentials, you may skip this step, as ``~/.composer/`` is mounted into the container from the host machine in order to share composer cache between projects, and also shares the global ``auth.json`` from the host machine.
    ```

7. Initialize project source files using composer create-project and then move them into place:

    ``` shell
    META_PACKAGE=magento/project-community-edition META_VERSION=2.4.x

    $ composer create-project --no-install --repository-url=https://repo.magento.com/ \
        "${META_PACKAGE}" /tmp/magento-tmp "${META_VERSION}"

    $ rsync -au --remove-source-files /tmp/magento-tmp/ /var/www/html/

    $ composer install
    ```

8. Install the application, and you should be all set:

    ```
    # Install Application
    bin/magento setup:install \
        --backend-frontname=backend \
        --amqp-host=rabbitmq \
        --amqp-port=5672 \
        --amqp-user=guest \
        --amqp-password=guest \
        --db-host=db \
        --db-name=magento \
        --db-user=magento \
        --db-password=magento \
        --search-engine=elasticsearch7 \
        --elasticsearch-host=elasticsearch \
        --elasticsearch-port=9200 \
        --elasticsearch-index-prefix=magento2 \
        --elasticsearch-enable-auth=0 \
        --elasticsearch-timeout=15 \
        --http-cache-hosts=varnish:80 \
        --session-save=redis \
        --session-save-redis-host=redis \
        --session-save-redis-port=6379 \
        --session-save-redis-db=2 \
        --session-save-redis-max-concurrency=20 \
        --cache-backend=redis \
        --cache-backend-redis-server=redis \
        --cache-backend-redis-db=0 \
        --cache-backend-redis-port=6379 \
        --page-cache=redis \
        --page-cache-redis-server=redis \
        --page-cache-redis-db=1 \
        --page-cache-redis-port=6379

    # Configure Application
    bin/magento config:set --lock-env web/unsecure/base_url \
        "https://${TRAEFIK_SUBDOMAIN:+$TRAEFIK_SUBDOMAIN.}${TRAEFIK_DOMAIN}/"

    bin/magento config:set --lock-env web/secure/base_url \
        "https://${TRAEFIK_SUBDOMAIN:+$TRAEFIK_SUBDOMAIN.}${TRAEFIK_DOMAIN}/"

    bin/magento config:set --lock-env web/secure/offloader_header X-Forwarded-Proto

    bin/magento config:set --lock-env web/secure/use_in_frontend 1
    bin/magento config:set --lock-env web/secure/use_in_adminhtml 1
    bin/magento config:set --lock-env web/seo/use_rewrites 1

    bin/magento config:set --lock-env system/full_page_cache/caching_application 2
    bin/magento config:set --lock-env system/full_page_cache/ttl 604800

    bin/magento config:set --lock-env catalog/search/enable_eav_indexer 1

    bin/magento config:set --lock-env dev/static/sign 0

    bin/magento deploy:mode:set -s developer
    bin/magento cache:disable block_html full_page

    bin/magento indexer:reindex
    bin/magento cache:flush
    ```

    ``` note::
            Prior to Magento ``2.4.x`` it was not required to enter search-engine and elasticsearch configuration during installation and these params     to ``setup:install`` are not supported by Magento ``2.3.x``. These should be omitted on older versions where not supported and Elasticsearch     configured via ``config:set`` instead:

            .. code::

                bin/magento config:set --lock-env catalog/search/engine elasticsearch7
                bin/magento config:set --lock-env catalog/search/elasticsearch7_server_hostname elasticsearch
                bin/magento config:set --lock-env catalog/search/elasticsearch7_server_port 9200
                bin/magento config:set --lock-env catalog/search/elasticsearch7_index_prefix magento2
                bin/magento config:set --lock-env catalog/search/elasticsearch7_enable_auth 0
            bin/magento config:set --lock-env catalog/search/elasticsearch7_server_timeout 15
   ```

9. Generate an admin user and configure 2FA for OTP

    ``` shell
    # Generate localadmin user
    ADMIN_PASS="$(pwgen -n1 16)"
    ADMIN_USER=localadmin

    bin/magento admin:user:create \
        --admin-password="${ADMIN_PASS}" \
        --admin-user="${ADMIN_USER}" \
        --admin-firstname="Local" \
        --admin-lastname="Admin" \
        --admin-email="${ADMIN_USER}@example.com"

    printf "u: %s\np: %s\n" "${ADMIN_USER}" "${ADMIN_PASS}"

    # Configure 2FA provider
    OTPAUTH_QRI=
    TFA_SECRET=$(python -c "import base64; print base64.b32encode('$(pwgen -A1 128)')" | sed 's/=*$//')
    OTPAUTH_URL=$(printf "otpauth://totp/%s%%3Alocaladmin%%40example.com?issuer=%s&secret=%s" \
        "${TRAEFIK_SUBDOMAIN}.${TRAEFIK_DOMAIN}" "${TRAEFIK_SUBDOMAIN}.${TRAEFIK_DOMAIN}" "${TFA_SECRET}"
    )

    bin/magento config:set --lock-env twofactorauth/general/force_providers google
    bin/magento security:tfa:google:set-secret "${ADMIN_USER}" "${TFA_SECRET}"

    printf "%s\n\n" "${OTPAUTH_URL}"
    printf "2FA Authenticator Codes:\n%s\n" "$(oathtool -s 30 -w 10 --totp --base32 "${TFA_SECRET}")"

    segno "${OTPAUTH_URL}" -s 4 -o "pub/media/${ADMIN_USER}-totp-qr.png"
    printf "%s\n\n" "https://${TRAEFIK_SUBDOMAIN}.${TRAEFIK_DOMAIN}/media/${ADMIN_USER}-totp-qr.png?t=$(date +%s)"
    ```

    ``` note::
        Use of 2FA is mandatory on Magento ``2.4.x`` and setup of 2FA should be skipped when installing ``2.3.x`` or earlier. Where 2FA is setup     manually via UI upon login rather than using the CLI commands above, the 2FA configuration email may be retrieved from `the Mailhog service <https://mailhog.reward.test/>`_.
    ```

10. Launch the application in your browser:

    * [https://your-awesome-m2-project.test/](https://your-awesome-m2-project.test/)
    * [https://your-awesome-m2-project.test/backend/](https://your-awesome-m2-project.test/backend/)
    * [https://rabbitmq.your-awesome-m2-project.test/](https://rabbitmq.your-awesome-m2-project.test/)
    * [https://elasticsearch.your-awesome-m2-project.test/](https://elasticsearch.your-awesome-m2-project.test/)

    ``` note::
        To completely destroy the ``your-awesome-m2-project`` environment we just created, run ``reward env down -v`` to tear down the project's Docker     containers, volumes, etc.
    ```

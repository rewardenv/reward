ARG PHP_VERSION
FROM docker.io/rewardenv/php-fpm:${PHP_VERSION}
USER root

RUN npm install -g grunt-cli gulp yarn

RUN curl -s https://files.magerun.net/n98-magerun.phar > /usr/bin/n98-magerun \
    && chmod +x /usr/bin/n98-magerun

RUN curl -o /etc/bash_completion.d/n98-magerun.phar.bash \
        https://raw.githubusercontent.com/netz98/n98-magerun/master/res/autocompletion/bash/n98-magerun.phar.bash

# Create mr alias for n98-magerun
RUN ln -s /usr/bin/n98-magerun /usr/bin/mr

USER www-data

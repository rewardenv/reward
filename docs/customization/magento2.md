## Magento 2

The following variables can be added to the project's `.env` file to enable additional database containers for use with
the Magento 2 (Commerce
Only) [split-database solution](https://devdocs.magento.com/guides/v2.3/config-guide/multi-master/multi-master.html).

* `REWARD_SPLIT_SALES=1`
* `REWARD_SPLIT_CHECKOUT=1`

Some unnecessary Magento 2 specific components can be disabled by using these environment variables in `.env` file:

* `REWARD_ELASTICSEARCH=0`
* `REWARD_VARNISH=0`
* `REWARD_RABBITMQ=0`


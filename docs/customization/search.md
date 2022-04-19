## Search Engine

### OpenSearch or Elasticsearch

Reward currently supports both Elasticsearch and OpenSearch as Search Engines.

To use one of them, you'll need to enable one of them the `.env` file:

* `REWARD_ELASTICSEARCH=0`
* `REWARD_OPENSEARCH=1`

*If you enable both, Reward will install Magento using OpenSearch.*

### OpenSearch Dashboards

Reward also supports OpenSearch Dashboards. Enable it in the `.env` file:

* `REWARD_OPENSEARCH_DASHBOARDS=1`

It is not a global service. You can reach it as a subdomain of the development url:

`https://opensearch-dashboards.projectname.test`

### OpenSearch Configuration

You can change the OpenSearch version by changing it in the `.env` file. The available version can be
found [here](https://github.com/rewardenv/reward/tree/main/images/opensearch).

* `OPENSEARCH_VERSION=1.2`

You can also configure the memory limitations for OpenSearch.

* `OPENSEARCH_XMS=64m`
* `OPENSEARCH_XMX=512m`

### Elasticsearch Configuration

You can change the Elasticsearch version by changing it in the `.env` file. The available version can be
found [here](https://github.com/rewardenv/reward/tree/main/images/elasticsearch).

* `ELASTICSEARCH_VERSION=7.16`

You can also configure the memory limitations for Elasticsearch.

* `ELASTICSEARCH_XMS=64m`
* `ELASTICSEARCH_XMX=512m`

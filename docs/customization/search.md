## Search Engine

Reward currently supports both Elasticsearch and OpenSearch as Search Engines.

To use one of them, you'll need to enable one of them the `.env` file:

* `REWARD_ELASTICSEARCH=1`
* `REWARD_OPENSEARCH=1`

If you enable both, Reward will install Magento using OpenSearch.

You can also configure the memory limitations for both of them.

* `ELASTICSEARCH_XMS=64m`
* `ELASTICSEARCH_XMX=512m`
* `OPENSEARCH_XMS=64m`
* `OPENSEARCH_XMX=512m`

### OpenSearch Dashboards

Reward also supports OpenSearch Dashboards. Enable it in the `.env` file:

```
REWARD_OPENSEARCH_DASHBOARDS=1
```

It is not a global service. You can reach it as a subdomain of the development url:

`https://opensearch-dashboards.projectname.test`

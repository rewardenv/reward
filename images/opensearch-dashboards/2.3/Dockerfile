FROM opensearchproject/opensearch-dashboards:2.3.0

RUN set -eux \
    && /usr/share/opensearch-dashboards/bin/opensearch-dashboards-plugin remove securityDashboards \
    && printf '%s\n%s\n' "server.host: '0'" "opensearch.hosts: [https://localhost:9200]" > /usr/share/opensearch-dashboards/config/opensearch_dashboards.yml

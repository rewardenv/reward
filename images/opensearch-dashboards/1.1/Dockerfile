FROM opensearchproject/opensearch-dashboards:1.1.0

RUN set -eux \
    && /usr/share/opensearch-dashboards/bin/opensearch-dashboards-plugin remove securityDashboards \
    && echo -e "server.host: '0'\nopensearch.hosts: [https://localhost:9200]" > /usr/share/opensearch-dashboards/config/opensearch_dashboards.yml

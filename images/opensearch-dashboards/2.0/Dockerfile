FROM opensearchproject/opensearch-dashboards:2.0.1

RUN set -eux \
    && /usr/share/opensearch-dashboards/bin/opensearch-dashboards-plugin remove securityDashboards \
    && echo -e "server.host: '0'\nopensearch.hosts: [https://localhost:9200]" > /usr/share/opensearch-dashboards/config/opensearch_dashboards.yml

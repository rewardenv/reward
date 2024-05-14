# syntax=docker/dockerfile:1
FROM opensearchproject/opensearch-dashboards:2.5.0

RUN <<-EOF
    set -eux
    /usr/share/opensearch-dashboards/bin/opensearch-dashboards-plugin remove securityDashboards
    printf '%s\n%s\n' "server.host: '0.0.0.0'" "opensearch.hosts: [https://localhost:9200]" > /usr/share/opensearch-dashboards/config/opensearch_dashboards.yml
EOF

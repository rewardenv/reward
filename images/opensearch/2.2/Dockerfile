FROM opensearchproject/opensearch:2.2.1

RUN set -eux \
    && bin/opensearch-plugin install analysis-phonetic \
    && bin/opensearch-plugin install analysis-icu

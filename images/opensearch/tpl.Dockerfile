FROM opensearchproject/opensearch:{{ getenv "IMAGE_TAG" "2.5.0" }}

RUN set -eux \
    && bin/opensearch-plugin install analysis-phonetic \
    && bin/opensearch-plugin install analysis-icu

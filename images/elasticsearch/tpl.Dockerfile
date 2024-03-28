FROM docker.elastic.co/elasticsearch/elasticsearch:{{ getenv "IMAGE_TAG" "latest" }}

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

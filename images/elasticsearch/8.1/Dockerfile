FROM docker.elastic.co/elasticsearch/elasticsearch:8.1.2

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

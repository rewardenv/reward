FROM docker.elastic.co/elasticsearch/elasticsearch:7.11.2

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

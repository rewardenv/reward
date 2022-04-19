FROM docker.elastic.co/elasticsearch/elasticsearch:7.17.2

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

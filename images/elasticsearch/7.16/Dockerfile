FROM docker.elastic.co/elasticsearch/elasticsearch:7.16.3

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

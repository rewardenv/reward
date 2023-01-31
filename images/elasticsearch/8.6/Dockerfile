FROM docker.elastic.co/elasticsearch/elasticsearch:8.6.1

RUN set -eux \
    && bin/elasticsearch-plugin install analysis-phonetic \
    && bin/elasticsearch-plugin install analysis-icu

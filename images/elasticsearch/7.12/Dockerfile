FROM docker.elastic.co/elasticsearch/elasticsearch:7.12.0

RUN set -eux \
  && bin/elasticsearch-plugin install analysis-phonetic \
  && bin/elasticsearch-plugin install analysis-icu

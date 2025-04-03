# syntax=docker/dockerfile:1
FROM docker.elastic.co/elasticsearch/elasticsearch:{{ getenv "IMAGE_TAG" "latest" }}

# https://experienceleague.adobe.com/en/docs/commerce-operations/upgrade-guide/prepare/prerequisites#upgrade-elasticsearchupgrade-elasticsearch
ENV ES_SETTING_INDICES_ID__FIELD__DATA_ENABLED=true

RUN <<-EOF
    set -eux
    bin/elasticsearch-plugin install analysis-phonetic
    bin/elasticsearch-plugin install analysis-icu
EOF

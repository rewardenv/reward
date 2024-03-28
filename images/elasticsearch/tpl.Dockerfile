# syntax=docker/dockerfile:1
FROM docker.elastic.co/elasticsearch/elasticsearch:{{ getenv "IMAGE_TAG" "latest" }}

RUN <<-EOF
    set -eux
    bin/elasticsearch-plugin install analysis-phonetic
    bin/elasticsearch-plugin install analysis-icu
EOF

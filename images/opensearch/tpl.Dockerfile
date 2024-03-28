# syntax=docker/dockerfile:1
FROM opensearchproject/opensearch:{{ getenv "IMAGE_TAG" "2.5.0" }}

RUN <<-EOF
    set -eux
    bin/opensearch-plugin install analysis-phonetic
    bin/opensearch-plugin install analysis-icu
EOF

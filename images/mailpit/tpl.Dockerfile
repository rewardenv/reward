FROM {{ getenv "IMAGE_REPOSITORY" "axllent/mailpit" }}:{{ getenv "IMAGE_TAG" "v1.15" }}

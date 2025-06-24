# syntax=docker/dockerfile:1
FROM valkey/valkey:{{ getenv "IMAGE_TAG" "8.1" }}-alpine

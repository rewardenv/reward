# syntax=docker/dockerfile:1
FROM mysql:{{ getenv "IMAGE_TAG" "8.0" }}

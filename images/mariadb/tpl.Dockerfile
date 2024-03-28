# syntax=docker/dockerfile:1
FROM mariadb:{{ getenv "IMAGE_TAG" "10.9" }}

# syntax=docker/dockerfile:1
FROM rabbitmq:{{ getenv "IMAGE_TAG" "3.13" }}-management

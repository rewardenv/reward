# syntax=docker/dockerfile:1
FROM redis:{{ getenv "IMAGE_TAG" "7.2" }}-alpine

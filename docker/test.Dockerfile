# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder
LABEL stage=gobuilder

WORKDIR /app

RUN apk add -u --no-cache \
    gcc \
    musl-dev \
    gcompat

COPY ../go.mod go.sum ./

RUN go mod download -x

COPY . .

CMD go test -race -v ./...

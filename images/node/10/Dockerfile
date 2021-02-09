FROM node:10-alpine

ARG DOCKER_START_COMMAND="yarn watch"

RUN apk add --no-cache --virtual .build-deps \
  python \
  make \
  gcc \
  g++ \
  && apk add --no-cache \
  git \
  openssh-client \
  yarn \
  && mkdir -p /usr/src/app \
  && chown node:node /usr/src/app

ENV GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
ENV DOCKER_START_COMMAND=${DOCKER_START_COMMAND}

WORKDIR /usr/src/app
USER node

CMD ["sh", "-c", "while true; do ${DOCKER_START_COMMAND}; sleep 10; done"]

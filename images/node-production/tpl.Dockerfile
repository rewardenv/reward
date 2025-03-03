# syntax=docker/dockerfile:1
FROM rewardenv/node:{{ getenv "IMAGE_TAG" "20" }}

ARG DOCKER_START_COMMAND="npm start"
ENV DOCKER_START_COMMAND=${DOCKER_START_COMMAND}

WORKDIR /usr/src/app
USER node

CMD ["sh", "-c", "while true; do ${DOCKER_START_COMMAND}; sleep 10; done"]

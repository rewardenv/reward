# syntax=docker/dockerfile:1.7-labs
FROM rewardenv/node:{{ getenv "IMAGE_TAG" "20" }}

COPY --from=scripts-lib --exclude=*_test.sh --chown=node:node --chmod=755 / /home/node/.local/lib/
COPY --from=scripts-bin --exclude=*_test.sh --chown=node:node --chmod=755 / /home/node/.local/bin/

USER root

RUN apk add --no-cache \
    bash \
    coreutils

USER node
WORKDIR /usr/src/app

ARG DOCKER_START_COMMAND="npm start"
ENV DOCKER_START_COMMAND=${DOCKER_START_COMMAND}
ENV PATH="/usr/src/app/node_modules/.bin/:/home/node/node_modules/.bin:/home/node/bin:/home/node/.local/bin:${PATH}"

CMD ["sh", "-c", "while true; do ${DOCKER_START_COMMAND}; sleep 10; done"]

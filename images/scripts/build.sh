#!/bin/bash
[ "$DEBUG" = "true" ] && set -x
set -e
trap '>&2 printf "\n\e[01;31mError: Command \`%s\` on line $LINENO failed with exit code $?\033[0m\n" "$BASH_COMMAND"' ERR

## find directory where this script is located following symlinks if necessary
readonly BASE_DIR="$(
  cd "$(
    dirname "$(
      (readlink "${BASH_SOURCE[0]}" || echo "${BASH_SOURCE[0]}") |
        sed -e "s#^../#$(dirname "$(dirname "${BASH_SOURCE[0]}")")/#"
    )"
  )" >/dev/null &&
    pwd
)/.."
pushd "${BASE_DIR}" >/dev/null

DOCKER_REGISTRY=${DOCKER_REGISTRY:-docker.io}
DOCKER_REPO=${DOCKER_REPO:-rewardenv}
IMAGE_BASE="${DOCKER_REGISTRY}/${DOCKER_REPO}"

TAG_AS_DEFAULT_BASE_IMAGE=${TAG_AS_DEFAULT_BASE_IMAGE:-debian-bullseye}
# If TAG_AS_DEFAULT environment variable is true, the image is going to be tagged without the base image postfix too.
# Eg.: rewardenv/php-fpm:7.4, rewardenv/php-fpm:7.4-fpm-debian
# Otherwise it will be tagged only with the base image postfix.
# Eg: rewardenv/php-fpm:7.4-fpm-debian
TAG_AS_DEFAULT=${TAG_AS_DEFAULT:-false}

printf >&2 "\n\e[01;31mUsing Docker Registry: %s and Docker Repo: %s.\033[0m\n" "${DOCKER_REGISTRY}" "${DOCKER_REPO//reward/repo-reward}"

function print_usage() {
  echo "build.sh [--push] [--dry-run] <IMAGE_TYPE>"
  echo
  echo "example:"
  echo "build.sh --push php-fpm"
}

# Parse long args and translate them to short ones.
for arg in "$@"; do
  shift
  case "$arg" in
  "--push") set -- "$@" "-p" ;;
  "--dry-run") set -- "$@" "-n" ;;
  "--help") set -- "$@" "-h" ;;
  *) set -- "$@" "$arg" ;;
  esac
done

PUSH=${PUSH:-''}
DRY_RUN=${DRY_RUN:-''}

# Parse short args.
OPTIND=1
while getopts "pnh" opt; do
  case "$opt" in
  "p") PUSH=true ;;
  "n") DRY_RUN=true ;;
  "?" | "h")
    print_usage >&2
    exit 1
    ;;
  esac
done
shift "$((OPTIND - 1))"

SEARCH_PATH="${1}"

if [[ ${DRY_RUN} ]]; then
  DOCKER_COMMAND="echo docker"
else
  DOCKER_COMMAND="docker"
fi

if [ "${DOCKER_USE_BUILDX}" = "true" ]; then
  DOCKER_BUILD_COMMAND=${DOCKER_BUILD_COMMAND:-buildx build}
else
  DOCKER_BUILD_COMMAND=${DOCKER_BUILD_COMMAND:-build}
fi
DOCKER_BUILD_PLATFORM=${DOCKER_BUILD_PLATFORM:-} # "linux/amd64,linux/arm/v7,linux/arm64"

## since fpm images no longer can be traversed, this script should require a search path vs defaulting to build all
if [[ -z ${SEARCH_PATH} ]]; then
  printf >&2 "\n\e[01;31mError: Missing search path. Please try again passing an image type as an argument!\033[0m\n"
  print_usage
  exit 1
fi

function version_gt() { test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"; }

function docker_login() {
  if [ "${PUSH}" = "true" ]; then
    if [[ ${DOCKER_USERNAME:-} ]]; then
      echo "Attempting non-interactive docker login (via provided credentials)"
      echo "${DOCKER_PASSWORD:-}" | ${DOCKER_COMMAND} login -u "${DOCKER_USERNAME:-}" --password-stdin "${DOCKER_REGISTRY}"
    elif [[ -t 1 ]]; then
      echo "Attempting interactive docker login (tty)"
      ${DOCKER_COMMAND} login "${DOCKER_REGISTRY}"
    fi
  fi
}

function docker_build() {
  if [ -n "${DOCKER_BUILD_PLATFORM}" ]; then
    DOCKER_BUILD_PLATFORM_ARG="--platform ${DOCKER_BUILD_PLATFORM}"
  fi

  printf "\e[01;31m==>\nBuilding %s \n\tFrom: %s/Dockerfile \n\tContext: %s \n\tPlatforms: %s\n\tTags: %s\n==>\033[0m\n" "${IMAGE_TAG}" "${BUILD_DIR}" "${BUILD_CONTEXT}" "${DOCKER_BUILD_PLATFORM}" "${BUILD_TAGS}"

  if [ "${PUSH}" = "true" ] && [ "${DOCKER_USE_BUILDX}" = "true" ]; then
    DOCKER_PUSH_ARG="--push"
    TAGS_ARG=$(printf -- "%s " "${BUILD_TAGS[@]/#/--tag }")
  else
    TAGS_ARG="-t ${IMAGE_TAG}"
  fi

  # shellcheck disable=SC2046
  # shellcheck disable=SC2086
  ${DOCKER_COMMAND} ${DOCKER_BUILD_COMMAND} \
    ${TAGS_ARG} \
    -f "${BUILD_DIR}/Dockerfile" \
    ${DOCKER_BUILD_PLATFORM_ARG} \
    ${DOCKER_PUSH_ARG} \
    $(printf -- "%s " "${BUILD_ARGS[@]/#/--build-arg }") \
    "${BUILD_CONTEXT}"

  # We have to manually push the images if not using docker buildx
  if [ "${DOCKER_USE_BUILDX}" != "true" ]; then
    for tag in "${BUILD_TAGS[@]}"; do
      ${DOCKER_COMMAND} tag "${IMAGE_TAG}" "${tag}"

      if [ "${PUSH}" = "true" ]; then ${DOCKER_COMMAND} push "${tag}"; fi
    done
  fi
}

function build_context() {
  # Check if the context directory exist in the build directory.
  # Eg.: Priorities
  #   1. php-fpm/centos7/magento2/blackfire/context
  #   2. php-fpm/centos7/magento2/context
  #   3. php-fpm/centos7/context
  #   4. php-fpm/context
  if [ "${DEBUG}" = "true" ]; then
    echo "Looking for context directory option 1: $(echo "${BUILD_DIR}" | rev | cut -d/ -f1- | rev)/context"
    echo "Looking for context directory option 2: $(echo "${BUILD_DIR}" | rev | cut -d/ -f2- | rev)/context"
    echo "Looking for context directory option 3: $(echo "${BUILD_DIR}" | rev | cut -d/ -f3- | rev)/context"
    echo "Looking for context directory option 4: $(echo "${BUILD_DIR}" | rev | cut -d/ -f4- | rev)/context"
  fi

  if [[ -d "$(echo "${BUILD_DIR}" | rev | cut -d/ -f1- | rev)/context" ]]; then
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | rev | cut -d/ -f1- | rev)/context"
    printf "Using context 1: %s\n" "${BUILD_CONTEXT}"
  elif [[ -d "$(echo "${BUILD_DIR}" | rev | cut -d/ -f2- | rev)/context" ]]; then
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | rev | cut -d/ -f2- | rev)/context"
    printf "Using context 2: %s\n" "${BUILD_CONTEXT}"
  elif [[ -d "$(echo "${BUILD_DIR}" | rev | cut -d/ -f3- | rev)/context" ]]; then
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | rev | cut -d/ -f3- | rev)/context"
    printf "Using context 3: %s\n" "${BUILD_CONTEXT}"
  elif [[ -d "$(echo "${BUILD_DIR}" | rev | cut -d/ -f4- | rev)/context" ]]; then
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | rev | cut -d/ -f4- | rev)/context"
    printf "Using context 4: %s\n" "${BUILD_CONTEXT}"
  else
    BUILD_CONTEXT="${BUILD_DIR}"
    printf "Using default working directory as context: %s\n" "${BUILD_CONTEXT}"
  fi
}

function build_image() {
  BUILD_DIR="$(dirname "${file}")"
  IMAGE_NAME=$(echo "${BUILD_DIR}" | cut -d/ -f1)
  IMAGE_TAG="${IMAGE_BASE}/${IMAGE_NAME}"
  # Base Image: centos7, centos8, debian
  BASE_IMAGE="$(echo "${BUILD_DIR}" | cut -d/ -f2- -s | cut -d/ -f1 -s)"

  if [ "${BASE_IMAGE}" = "${TAG_AS_DEFAULT_BASE_IMAGE}" ]; then
    TAG_AS_DEFAULT="true"
  fi

  if [ "$BASE_IMAGE" ]; then
    # Tag Suffix: magento2-centos7, magento2-debug-centos7
    TAG_SUFFIX="$(echo "${BUILD_DIR}" | cut -d/ -f3- -s | tr / - | sed 's/^-//')-${BASE_IMAGE}"
  else
    # Tag Suffix: 7.12
    TAG_SUFFIX="$(echo "${BUILD_DIR}" | cut -d/ -f2- -s | tr / - | sed 's/^-//')"
  fi
  # If the TAG_SUFFIX contains "_base", we will remove it
  if [[ ${TAG_SUFFIX} == *"_base"* ]]; then
    TAG_SUFFIX=$(echo "$TAG_SUFFIX" | sed -r 's/_base-//')
  fi

  if [[ ${BUILD_VERSION:-} ]]; then
    export PHP_VERSION="${MAJOR_VERSION}"
    BUILD_ARGS=(PHP_VERSION)
  else
    BUILD_ARGS=()
  fi

  echo "=========================="
  echo "$PHP_VERSION"
  echo "$BUILD_DIR"
  echo "=========================="
  # Xdebug2 doesn't exist for php < 7.1 or php >= 8.0 or later. We should skip this step.
  if [[ ${BUILD_DIR} =~ xdebug2 ]]; then
    if version_gt "${PHP_VERSION}" "7.99.99"; then
      echo "Skipping build."
      return
    fi
  fi

  # PHP Images built with different method than others.
  #   So at the end of this if we build and push the images and return.
  if [[ "${SEARCH_PATH}" =~ php$|php/.+ ]]; then
    build_context

    # Strip the term 'cli' from tag suffix as this is the default variant
    TAG_SUFFIX="$(echo "${TAG_SUFFIX}" | sed -E 's/^(cli$|cli-)//')"
    [[ ${TAG_SUFFIX} ]] && TAG_SUFFIX="-${TAG_SUFFIX}"

    BUILD_TAGS=(
      "${IMAGE_TAG}:${MAJOR_VERSION}${TAG_SUFFIX}"
    )

    for TAG in "${BUILD_TAGS[@]}"; do
      if [ "${TAG_AS_DEFAULT}" = "true" ]; then
        SHORT_TAG=$(echo "${TAG}" | sed -r "s/-?${BASE_IMAGE}//")
        BUILD_TAGS+=("${SHORT_TAG}")
      fi
    done

    docker_build

    return 0

  # PHP-FPM images will not have each version in a directory tree; require version be passed
  #   in as env variable for use as a build argument.
  elif [[ ${SEARCH_PATH} == *fpm* ]]; then
    if [[ -z ${PHP_VERSION} ]]; then
      printf >&2 "\n\e[01;31mError: Building %s images requires PHP_VERSION env variable be set!\033[0m\n" "${SEARCH_PATH}"
      exit 1
    fi

    export PHP_VERSION

    IMAGE_TAG+=":${PHP_VERSION}"
    if [[ ${TAG_SUFFIX} ]]; then
      IMAGE_TAG+="-${TAG_SUFFIX}"
    fi
    BUILD_ARGS+=("PHP_VERSION")

    # Support for PHP 8 images which require (temporarily at least) use of non-loader variant of base image
    if [[ ${PHP_VARIANT:-} ]]; then
      export PHP_VARIANT
      BUILD_ARGS+=("PHP_VARIANT")
    fi
  else
    IMAGE_TAG+=":${TAG_SUFFIX}"
  fi

  build_context

  BUILD_TAGS=("${IMAGE_TAG}")

  if [ "${TAG_AS_DEFAULT}" = "true" ]; then
    SHORT_TAG=$(echo "${IMAGE_TAG}" | sed -r "s/-?${BASE_IMAGE}//")
    BUILD_TAGS+=("${SHORT_TAG}")
  fi

  if [[ -n "${LATEST_TAG:+x}" && ${IMAGE_TAG} == *"${LATEST_TAG}"* ]]; then
    LATEST_TAG=$(echo "${IMAGE_TAG}" | sed -r "s/([^:]*:).*/\1latest/")
    BUILD_TAGS+=("${SHORT_TAG}")
  fi

  docker_build

  return 0
}

## Login to docker hub as needed
docker_login

# For PHP Build we have to use a specific order and version list
if [[ "${SEARCH_PATH}" =~ php$|php/(.+) ]]; then
  if [ "${BASH_REMATCH[1]}" ]; then
    SEARCH_PATH="php"
    VARIANT_LIST="${BASH_REMATCH[1]}"
  fi

  DEFAULT_IMAGES=("debian-bullseye")
  DEFAULT_VERSIONS=("5.6" "7.0" "7.1" "7.2" "7.3" "7.4" "8.0")
  DEFAULT_VARIANTS=("cli" "fpm" "cli-loaders" "fpm-loaders")

  if [[ -z ${DOCKER_BASE_IMAGES:-} ]]; then DOCKER_BASE_IMAGES=("${DEFAULT_IMAGES[*]}"); fi
  if [[ -z ${VERSION_LIST:-} ]]; then VERSION_LIST=("${DEFAULT_VERSIONS[*]}"); fi
  if [[ -z ${VARIANT_LIST:-} ]]; then VARIANT_LIST=("${DEFAULT_VARIANTS[*]}"); fi

  for IMG in ${DOCKER_BASE_IMAGES[@]}; do
    for BUILD_VERSION in ${VERSION_LIST[*]}; do
      MAJOR_VERSION="$(echo "${BUILD_VERSION}" | sed -E 's/([0-9])([0-9])/\1.\2/')"
      for BUILD_VARIANT in ${VARIANT_LIST[*]}; do
        for file in $(find "${SEARCH_PATH}/${IMG}/${BUILD_VARIANT}" -type f -name Dockerfile | sort -t_ -k1,1 -d); do
          build_image
        done
      done
    done
  done
else
  # For the rest we iterate through the folders to create them by version folder
  for file in $(find "${SEARCH_PATH}" -type f -name Dockerfile | sort -t_ -k1,1 -d); do

    # Due to build matrix requirements, magento1, magento2 and wordpress specific variants are built in
    #   separate invocation so we skip this one.
    if [[ "${SEARCH_PATH}" == "php-fpm" ]] && [[ ${file} =~ php-fpm/[^\/]+/(magento[1-2]|shopware|wordpress) ]]; then
      continue
    fi

    build_image
  done
fi

exit 0

#!/usr/bin/env bash
[ "$DEBUG" == "true" ] && set -x
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

DOCKER_REGISTRY="docker.io"
IMAGE_BASE="${DOCKER_REGISTRY}/rewardenv"
DEFAULT_BASE=${DEFAULT_BASE:-'centos7'}

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
  DOCKER="echo docker"
else
  DOCKER="docker"
fi

## since fpm images no longer can be traversed, this script should require a search path vs defaulting to build all
if [[ -z ${SEARCH_PATH} ]]; then
  printf >&2 "\n\e[01;31mError: Missing search path. Please try again passing an image type as an argument!\033[0m\n"
  print_usage
  exit 1
fi

function docker_login() {
  if [[ ${PUSH} ]]; then
    if [[ ${DOCKER_USERNAME:-} ]]; then
      echo "Attempting non-interactive docker login (via provided credentials)"
      echo "${DOCKER_PASSWORD:-}" | ${DOCKER} login -u "${DOCKER_USERNAME:-}" --password-stdin "${DOCKER_REGISTRY}"
    elif [[ -t 1 ]]; then
      echo "Attempting interactive docker login (tty)"
      ${DOCKER} login "${DOCKER_REGISTRY}"
    fi
  fi
}

function build_context() {
  # Check if the context directory exist in the build directory.
  # Eg.: Priorities
  #   1. php-fpm/centos7/magento2/context
  #   2. php-fpm/centos7/context
  #   3. php-fpm/context
  if [[ -d "$(echo ${BUILD_DIR} | rev | cut -d/ -f1- | rev)/context" ]]; then
#    echo 1
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | rev | cut -d/ -f1- | rev)/context"
  elif [[ -d "$(echo ${BUILD_DIR} | rev | cut -d/ -f2- | rev)/context" ]]; then
#    echo 2
    BUILD_CONTEXT="$(echo ${BUILD_DIR} | rev | cut -d/ -f2-| rev)/context"
  elif [[ -d "$(echo ${BUILD_DIR} | rev | cut -d/ -f3- | rev)/context" ]]; then
#    echo 3
    BUILD_CONTEXT="$(echo ${BUILD_DIR} | rev | cut -d/ -f3-| rev)/context"
  else
#    echo 4
    BUILD_CONTEXT="${BUILD_DIR}"
  fi
#  echo 1 "$(echo ${BUILD_DIR} | rev | cut -d/ -f1- | rev)/context"
#  echo 2 "$(echo ${BUILD_DIR} | rev | cut -d/ -f2- | rev)/context"
#  echo 3 "$(echo ${BUILD_DIR} | rev | cut -d/ -f3- | rev)/context"
#  echo 4 "$BUILD_DIR"
#  echo $BUILD_CONTEXT
}

function build_image() {
  BUILD_DIR="$(dirname "${file}")"
  IMAGE_NAME=$(echo "${BUILD_DIR}" | cut -d/ -f1)
  IMAGE_TAG="${IMAGE_BASE}/${IMAGE_NAME}"
  # Base Image: centos7, centos8, debian
  BASE_IMAGE="$(echo "${BUILD_DIR}" | cut -d/ -f2- -s | cut -d/ -f1 -s)"
  if [ $BASE_IMAGE ]; then
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

  # PHP Images built with different method than others.
  #   So at the end of this if we build and push the images and return.
  if [[ "${SEARCH_PATH}" =~ php$|php/.+ ]]; then
    build_context

    # Strip the term 'cli' from tag suffix as this is the default variant
    TAG_SUFFIX="$(echo "${TAG_SUFFIX}" | sed -E 's/^(cli$|cli-)//')"
    [[ ${TAG_SUFFIX} ]] && TAG_SUFFIX="-${TAG_SUFFIX}"

    printf "\e[01;31m==> building %s from %s/Dockerfile with context %s\033[0m\n" "${IMAGE_NAME}" "${BUILD_DIR}" "${BUILD_CONTEXT}"
    # Build the default version of the image
    # shellcheck disable=SC2046
    ${DOCKER} build \
      -t "${IMAGE_NAME}:build" \
      -f "${BUILD_DIR}/Dockerfile" \
      "${BUILD_CONTEXT}" \
      $(printf -- "--build-arg %s " "${BUILD_ARGS[@]}")

    # Fetch the precise php version from the built image and tag it
    MINOR_VERSION="$(${DOCKER} run --rm -t --entrypoint php \
      "${IMAGE_NAME}:build" -r 'preg_match("#^\d+(\.\d+)*#", PHP_VERSION, $match); echo $match[0];')"

    # Generate array of tags for the image being built
    IMAGE_TAGS=(
      "${IMAGE_TAG}:${MAJOR_VERSION}${TAG_SUFFIX}"
      "${IMAGE_TAG}:${MINOR_VERSION}${TAG_SUFFIX}"
    )

    # Iterate and push image tags to remote registry
    for TAG in "${IMAGE_TAGS[@]}"; do
      ${DOCKER} tag "${IMAGE_NAME}:build" "${TAG}"
      printf "\e[01;31m==> Successfully tagged %s\033[0m\n" "${TAG}"

      if [[ ${TAG} == *"${DEFAULT_BASE}"* ]]; then
        SHORT_TAG=$(echo "${TAG}" | sed -r "s/-?${DEFAULT_BASE}//")
        ${DOCKER} tag "${IMAGE_NAME}:build" "${SHORT_TAG}"
        printf "\e[01;31m==> Successfully tagged %s\033[0m\n" "${SHORT_TAG}"
        [[ $PUSH ]] && PUSH_SHORT=true
      fi

      [[ $PUSH ]] && ${DOCKER} push "${TAG}"
      [[ $PUSH_SHORT ]] && ${DOCKER} push "${SHORT_TAG}"
    done
    ${DOCKER} image rm "${IMAGE_NAME}:build" &>/dev/null || true

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
    BUILD_ARGS+=("--build-arg")
    BUILD_ARGS+=("PHP_VERSION")

    # Support for PHP 8 images which require (temporarily at least) use of non-loader variant of base image
    if [[ ${PHP_VARIANT:-} ]]; then
      export PHP_VARIANT
      BUILD_ARGS+=("--build-arg")
      BUILD_ARGS+=("PHP_VARIANT")
    fi
  else
    IMAGE_TAG+=":${TAG_SUFFIX}"
  fi

  build_context

  printf "\e[01;31m==> building %s from %s/Dockerfile with context %s\033[0m\n" "${IMAGE_TAG}" "${BUILD_DIR}" "${BUILD_CONTEXT}"
  ${DOCKER} build \
    -t "${IMAGE_TAG}" \
    -f "${BUILD_DIR}/Dockerfile" \
    "${BUILD_ARGS[@]}" \
    "${BUILD_CONTEXT}"

  if [[ ${IMAGE_TAG} == *"${DEFAULT_BASE}"* ]]; then
    SHORT_TAG=$(echo "${IMAGE_TAG}" | sed -r "s/-?${DEFAULT_BASE}//")
    ${DOCKER} tag "${IMAGE_TAG}" "${SHORT_TAG}"
    printf "\e[01;31m==> Successfully tagged %s\033[0m\n" "${SHORT_TAG}"
    [[ $PUSH ]] && PUSH_SHORT=true
  fi

  if [[ -n "${LATEST_TAG:+x}" && ${IMAGE_TAG} == *"${LATEST_TAG}"* ]]; then
    LATEST_TAG=$(echo "${IMAGE_TAG}" | sed -r "s/([^:]*:).*/\1latest/")
    ${DOCKER} tag "${IMAGE_TAG}" "${LATEST_TAG}"
    printf "\e[01;31m==> Successfully tagged %s\033[0m\n" "${LATEST_TAG}"
    [[ $PUSH ]] && PUSH_LATEST=true
  fi

  [[ $PUSH ]] && ${DOCKER} push "${IMAGE_TAG}"
  [[ $PUSH_SHORT ]] && ${DOCKER} push "${SHORT_TAG}"
  [[ $PUSH_LATEST ]] && ${DOCKER} push "${LATEST_TAG}"

  unset PUSH_SHORT PUSH_LATEST

  return 0
}

## Login to docker hub as needed
docker_login

# For PHP Build we have to use a specific order and version list
if [[ "${SEARCH_PATH}" =~ php$|php/(.+) ]]; then
  if [ "${BASH_REMATCH[1]}" ]; then
    SEARCH_PATH="php"
    VARIANT_LIST="${BASH_REMATCH[1]}";
  fi

  IMAGES=("centos7" "centos8" "debian")
  VERSIONS=("5.6" "7.0" "7.1" "7.2" "7.3" "7.4" "8.0")
  VARIANTS=("cli" "fpm" "cli-loaders" "fpm-loaders")

  if [[ -z ${DOCKER_BASE_IMAGES:-} ]]; then DOCKER_BASE_IMAGES=("${IMAGES[*]}"); fi
  if [[ -z ${VERSION_LIST:-} ]]; then VERSION_LIST=("${VERSIONS[*]}"); fi
  if [[ -z ${VARIANT_LIST:-} ]]; then VARIANT_LIST=("${VARIANTS[*]}"); fi

  for IMG in ${DOCKER_BASE_IMAGES[*]}; do
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

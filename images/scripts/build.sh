#!/usr/bin/env bash
set -x
set -e
trap '>&2 printf "\n\e[01;31mError: Command \`%s\` on line $LINENO failed with exit code $?\033[0m\n" "$BASH_COMMAND"' ERR

## find directory where this script is located following symlinks if necessary
readonly BASE_DIR="$(
  cd "$(
    dirname "$(
      (readlink "${BASH_SOURCE[0]}" || echo "${BASH_SOURCE[0]}") \
        | sed -e "s#^../#$(dirname "$(dirname "${BASH_SOURCE[0]}")")/#"
    )"
  )" >/dev/null \
  && pwd
)/.."
pushd "${BASE_DIR}" >/dev/null

DOCKER_REGISTRY="docker.io"
IMAGE_BASE="docker.io/rewardenv"


# Parse long args and translate them to short ones.
for arg in "$@"; do
  shift
  case "$arg" in
    "--push")    set -- "$@" "-p" ;;
    "--dry-run") set -- "$@" "-n" ;;
    *)           set -- "$@" "$arg"
  esac
done

PUSH_FLAG=
DRY_RUN_FLAG=

# Parse short args.
OPTIND=1
while getopts "pn" opt
do
  case "$opt" in
    "p") PUSH_FLAG=true ;;
    "n") DRY_RUN_FLAG=true ;;
    "?") print_usage >&2; exit 1 ;;
  esac
done
shift "$((OPTIND - 1))"

SEARCH_PATH="${1}"

if [[ ${DRY_RUN_FLAG} ]]; then
  DOCKER="echo"
else
  DOCKER="docker"
fi

## since fpm images no longer can be traversed, this script should require a search path vs defaulting to build all
if [[ -z ${SEARCH_PATH} ]]; then
  >&2 printf "\n\e[01;31mError: Missing search path. Please try again passing an image type as an argument!\033[0m\n"
  exit 1
fi

function docker_login () {
  if [[ ${PUSH_FLAG} ]]; then
    if [[ ${DOCKER_USERNAME:-} ]]; then
      echo "Attempting non-interactive docker login (via provided credentials)"
      echo "${DOCKER_PASSWORD:-}" | ${DOCKER} login -u "${DOCKER_USERNAME:-}" --password-stdin "${DOCKER_REGISTRY}"
    elif [[ -t 1 ]]; then
      echo "Attempting interactive docker login (tty)"
      ${DOCKER} login "${DOCKER_REGISTRY}"
    fi
  fi
}

function build_image () {
  BUILD_DIR="$(dirname "${file}")"
  IMAGE_NAME=$(echo "${BUILD_DIR}" | cut -d/ -f1)
  IMAGE_TAG="${IMAGE_BASE}/${IMAGE_NAME}"
  TAG_SUFFIX="$(echo "${BUILD_DIR}" | cut -d/ -f2- -s | tr / - | sed 's/^-//')"

  if [[ ${BUILD_VERSION:-} ]]; then
    export PHP_VERSION="${MAJOR_VERSION}"
    BUILD_ARGS=(PHP_VERSION)
  else
    BUILD_ARGS=()
  fi

  # PHP Images built with different method than others.
  #   So at the end of this if we build and push the images and return.
  if [[ ${SEARCH_PATH} = "php" ]]; then
    if [[ -d "$(echo "${BUILD_DIR}" | cut -d/ -f1)/context" ]]; then
      BUILD_CONTEXT="$(echo "${BUILD_DIR}" | cut -d/ -f1)/context"
    else
      BUILD_CONTEXT="${BUILD_DIR}"
    fi

    # Strip the term 'cli' from tag suffix as this is the default variant
    TAG_SUFFIX="$(echo "${BUILD_VARIANT}" | sed -E 's/^(cli$|cli-)//')"
    [[ ${TAG_SUFFIX} ]] && TAG_SUFFIX="-${TAG_SUFFIX}"

    # Build the default version of the image
    # shellcheck disable=SC2046
    "${DOCKER}" build \
      -t "${IMAGE_NAME}:build" \
      -f "${BUILD_DIR}/Dockerfile" \
      "${BUILD_CONTEXT}" \
      $(printf -- "--build-arg %s " "${BUILD_ARGS[@]}")

    # Fetch the precise php version from the built image and tag it
    MINOR_VERSION="$(${DOCKER} run --rm -t --entrypoint php "${IMAGE_NAME}:build" -r 'echo phpversion();')"

    # Generate array of tags for the image being built
    IMAGE_TAGS=(
      "${IMAGE_TAG}:${MAJOR_VERSION}${TAG_SUFFIX}"
      "${IMAGE_TAG}:${MINOR_VERSION}${TAG_SUFFIX}"
    )

    # Iterate and push image tags to remote registry
    for TAG in "${IMAGE_TAGS[@]}"; do
      "${DOCKER}" tag "${IMAGE_NAME}:build" "${TAG}"
      echo "Successfully tagged ${TAG}"
      [[ $PUSH_FLAG ]] && "${DOCKER}" push "${TAG}"
    done
    "${DOCKER}" image rm "${IMAGE_NAME}:build" || true

    return

  # PHP-FPM images will not have each version in a directory tree; require version be passed
  #   in as env variable for use as a build argument.
  elif [[ ${SEARCH_PATH} = *fpm* ]]; then
    if [[ -z ${PHP_VERSION} ]]; then
      >&2 printf "\n\e[01;31mError: Building %s images requires PHP_VERSION env variable be set!\033[0m\n" "${SEARCH_PATH}"
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

  if [[ -d "$(echo "${BUILD_DIR}" | cut -d/ -f1)/context" ]]; then
    BUILD_CONTEXT="$(echo "${BUILD_DIR}" | cut -d/ -f1)/context"
  else
    BUILD_CONTEXT="${BUILD_DIR}"
  fi

  printf "\e[01;31m==> building %s from %s/Dockerfile with context %s\033[0m\n" "${IMAGE_TAG}" "${BUILD_DIR}" "${BUILD_CONTEXT}"
  ${DOCKER} build \
    -t "${IMAGE_TAG}" \
    -f "${BUILD_DIR}/Dockerfile" \
    "${BUILD_ARGS[@]}" \
    "${BUILD_CONTEXT}"

  [[ $PUSH_FLAG ]] && "${DOCKER}" push "${IMAGE_TAG}"

  return
}

## Login to docker hub as needed
docker_login

# For PHP Build we have to use a specific order and version list
if [[ ${SEARCH_PATH} = "php" ]]; then
  VERSIONS=("5.5" "5.6" "7.0" "7.1" "7.2" "7.3" "7.4" "8.0")
  VARIANTS=("cli" "fpm" "cli-loaders" "fpm-loaders")

  if [[ -z ${VERSION_LIST:-} ]]; then VERSION_LIST=( "${VERSIONS[*]}" ); fi
  if [[ -z ${VARIANT_LIST:-} ]]; then VARIANT_LIST=( "${VARIANTS[*]}" ); fi

  for BUILD_VERSION in ${VERSION_LIST[*]}; do
    MAJOR_VERSION="$(echo "${BUILD_VERSION}" | sed -E 's/([0-9])([0-9])/\1.\2/')"
    for BUILD_VARIANT in ${VARIANT_LIST[*]}; do
      for file in $(find "${SEARCH_PATH}/${BUILD_VARIANT}" -type f -name Dockerfile | sort -V); do
        build_image
      done
    done
  done
else
  # For the rest we iterate through the folders to create them by version folder
  for file in $(find "${SEARCH_PATH}" -type f -name Dockerfile | sort -V); do

    # Due to build matrix requirements, magento1 and magento2 specific varients are built in separate invocation
    #   so we skip this one.
    if [[ ${SEARCH_PATH} == "php-fpm" ]] && [[ ${file} =~ php-fpm/magento[1-2] ]]; then
      continue
    fi

    build_image
  done
fi




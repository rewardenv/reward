#!/usr/bin/env bash
#===============================================================================
#          FILE:  build.sh
#
#         USAGE:  ./build.sh
#
#        AUTHOR:  mixe3y (Janos Miko), janos.miko@itg.cloud
#       COMPANY:  ITG
#       VERSION:  1.0
#       CREATED:  01/04/2021 15:59:53 CET
#===============================================================================
[ "$DEBUG" == "true" ] && set -x
set -eo pipefail

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

goreleaser --rm-dist --snapshot

pushd docs
make html
popd

if [[ ${*} =~ ^test ]]; then
  shift
  declare -a TEST_ARRAY=(
    "install"
    "install --reinstall"
    "install --uninstall"
    "install --ca-cert"
    "install --ssh-key"
    "install --ssh-config"
    "install --dns"
  )

  echo "build centos"
  docker build -t test-centos -f Dockerfile-centos .
  echo "build fedora"
  docker build -t test-fedora -f Dockerfile-fedora .
  echo "build ubuntu"
  docker build -t test-ubuntu -f Dockerfile-ubuntu .

  for i in "${TEST_ARRAY[@]}"; do
    echo "run tests on centos for ${i}"
    docker run --rm -it test-centos "${i}"
    echo "run tests on fedora for ${i}"
    docker run --rm -it test-fedora "${i}"
    echo "run tests on ubuntu for ${i}"
    docker run --rm -it test-ubuntu "${i}"
  done
fi

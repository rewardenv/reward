# How to build specific images locally

## Dnsmasq

```bash
export BUILD_TAG=latest
export IMAGE_TAG=latest
gomplate -f images/dnsmasq/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      --build-arg IMAGE_TAG \
      -t rewardenv/dnsmasq:${BUILD_TAG} \
      images/dnsmasq/context
```

## PHP-FPM

```bash
export BASE_IMAGE_NAME="debian"
export BASE_IMAGE_TAG="bookworm"
export PHP_VERSION="8.1"
gomplate -f images/php-fpm/base/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/php-fpm:${PHP_VERSION} \
      --build-arg PHP_VERSION \
      --progress plain \
      images/php-fpm/base/context
```

## PHP-FPM Rootless

```bash
export BASE_IMAGE_NAME="debian"
export BASE_IMAGE_TAG="bookworm"
export PHP_VERSION="8.1"
gomplate -f images/php-fpm-rootless/base/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/php-fpm:${PHP_VERSION} \
      --build-arg PHP_VERSION \
      --progress plain \
      images/php-fpm-rootless/base/context
```

## SSHD

```bash
export BUILD_TAG="latest"
export IMAGE_TAG="3.19"
gomplate -f images/sshd/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/sshd:${BUILD_TAG} \
      --build-arg IMAGE_TAG \
      --progress plain \
      images/sshd/context
```

## Varnish

```bash
# 7.4
export BUILD_TAG="latest"
export VARNISH_VERSION="7.4.1-1"
export VARNISH_REPO_VERSION="74"
export VARNISH_MODULES_BRANCH="7.4"
export DISTRO="ubuntu"
export DISTRO_RELEASE="jammy"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      images/varnish/context

# 6.6
export BUILD_TAG="6.6"
export VARNISH_VERSION="6.6.2-1"
export VARNISH_REPO_VERSION="66"
export VARNISH_MODULES_BRANCH="6.6"
export DISTRO="ubuntu"
export DISTRO_RELEASE="focal"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      images/varnish/context

# 6.5
export BUILD_TAG="6.5"
export VARNISH_VERSION="6.5.2"
export VARNISH_REPO_VERSION="65"
export VARNISH_MODULES_BRANCH="6.5"
export DISTRO="ubuntu"
export DISTRO_RELEASE="focal-1"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      images/varnish/context

# 6.4
export BUILD_TAG="6.4"
export VARNISH_VERSION="6.4.0-1"
export VARNISH_REPO_VERSION="64"
export VARNISH_MODULES_BRANCH="6.4"
export DISTRO="debian"
export DISTRO_RELEASE="buster"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      images/varnish/context

# 6.0
export BUILD_TAG="6.0"
export VARNISH_VERSION="6.0.13-1"
export VARNISH_REPO_VERSION="60lts"
export VARNISH_MODULES_BRANCH="6.0-lts"
export DISTRO="debian"
export DISTRO_RELEASE="buster"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      images/varnish/context
```

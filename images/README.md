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
      --build-context scripts=images/_php-fpm-scripts/context/
      --progress plain \
      images/php-fpm-rootless/base/context

gomplate -f images/php-fpm-rootless/shopware-web/tpl.Dockerfile -o - \
  | docker build \
      -f - \
      -t rewardenv/php-fpm:${PHP_VERSION}-shopware-web \
      --build-arg PHP_VERSION \
      --build-context scripts=images/_php-fpm-common/shopware-web \
      --progress plain \
      images/php-fpm-rootless/shopware-web/context
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
# 7.7
export BUILD_TAG="latest"
export VARNISH_VERSION="7.7.2-1"
export VARNISH_REPO_VERSION="77"
export VARNISH_MODULES_BRANCH="7.7"
export DISTRO="ubuntu"
export DISTRO_RELEASE="jammy"
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
export DISTRO="ubuntu"
export DISTRO_RELEASE="noble"
gomplate -f images/varnish/tpl.Dockerfile -o - \
  | docker buildx build \
      -f - \
      -t rewardenv/varnish:${BUILD_TAG} \
      --platform linux/amd64,linux/arm64 \
      images/varnish/context
```

# Run automated tests

```bash
find images/_common/lib/ -name "*_test.sh" -type f -print0 | xargs -0 -t bashunit
find images/_common/bin/ -name "*_test.sh" -type f -print0 | xargs -0 -t bashunit
```

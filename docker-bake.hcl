# docker-bake.hcl — PROTOTYPE for the PHP image chain.
#
# Goal: replace the 6-stage workflow-dispatch chain + ~50 per-image workflow
# files with ONE dependency graph that BuildKit builds in a single invocation,
# with shared layer cache.
#
# gomplate is KEPT, not removed. Several templates (php/cli, php-fpm/base, the
# xdebug/spx/blackfire utils) use gomplate `{{ if eq $IMAGE_NAME "ubuntu" }}`
# conditionals for distro-specific package logic (e.g. the ondrej PPA only on
# Ubuntu) — that is real build logic, not FROM substitution. The workflow
# renders each tpl.Dockerfile -> Dockerfile with gomplate first, then bake
# builds the rendered Dockerfiles. Dropping gomplate later is possible but is a
# separate refactor (move the distro `if` blocks into shell `case` statements
# inside RUN steps); it is NOT a prerequisite for this prototype.
#
# Scope of this prototype: the magento2 leaf and everything it depends on:
#   cli -> fpm -> fpm-loaders -> php-fpm/base -> php-fpm/magento2
# plus the cli-loaders branch (cli -> cli-loaders) for completeness.
#
# The layer DAG is expressed once here via `contexts`. The version x platform
# fan-out is driven by the workflow (one bake call per PHP version on a native
# runner), so each concern lives in exactly one place.
#
# Local usage:
#   docker buildx bake -f docker-bake.hcl magento2-chain                 # build whole chain, one version
#   PHP_VERSION=8.4 docker buildx bake -f docker-bake.hcl magento2-chain
#   docker buildx bake -f docker-bake.hcl --print magento2-chain         # inspect resolved graph
#
# NOTE: targets build the rendered `Dockerfile` next to each `tpl.Dockerfile`.
# Render first (the workflow does this automatically):
#   gomplate -f images/php/cli/tpl.Dockerfile -o images/php/cli/Dockerfile   # etc.

variable "REGISTRY" {
  default = "docker.io/rewardenv"
}

# Single source of truth for the PHP version of a given bake invocation.
# The full version matrix lives in the workflow (and could be one JSON file).
variable "PHP_VERSION" {
  default = "8.5"
}

variable "BASE_IMAGE_NAME" {
  default = "ubuntu"
}

variable "BASE_IMAGE_TAG" {
  default = "jammy"
}

# Single arch by default for fast local builds; the CI workflow sets both and
# builds each arch on its own NATIVE runner (no QEMU).
variable "PLATFORMS" {
  default = "linux/amd64"
}

# Per-arch tag suffix for the native-ARM build pattern: each arch builds on its
# own native runner and pushes `<tag>-amd64` / `<tag>-arm64`; the merge job then
# combines them into the final multi-arch manifest `<tag>`. Empty for local
# single-arch builds. Does NOT affect the DAG: `contexts` wire target->target
# inside one bake run, independent of the published tag names.
variable "TAG_SUFFIX" {
  default = ""
}

# Shared registry cache. mode=max caches intermediate layers (the expensive
# apt/extension-compile steps), so unchanged layers are reused across runs.
function "cache_from" {
  params = [ref]
  result = ["type=registry,ref=${REGISTRY}/buildcache:${ref}-${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}"]
}

function "cache_to" {
  params = [ref]
  result = ["type=registry,ref=${REGISTRY}/buildcache:${ref}-${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG},mode=max"]
}

# Common settings every target inherits.
target "_common" {
  platforms = split(",", PLATFORMS)
  args = {
    PHP_VERSION     = PHP_VERSION
    BASE_IMAGE_NAME = BASE_IMAGE_NAME
    BASE_IMAGE_TAG  = BASE_IMAGE_TAG
    # Pin IMAGE_NAME to the full ref so each FROM resolves to a string that
    # matches the `contexts` keys below (which is how BuildKit substitutes the
    # in-graph dependency instead of pulling from the registry).
    IMAGE_NAME = "${REGISTRY}/php"
  }
}

# ---- layer 0: base PHP (cli) — FROM ubuntu:jammy --------------------------
target "cli" {
  inherits   = ["_common"]
  context    = "images/php/cli/context"
  dockerfile = "../Dockerfile"
  args = {
    IMAGE_NAME = "ubuntu" # cli's FROM is the OS image, not rewardenv/php
    IMAGE_TAG  = BASE_IMAGE_TAG
  }
  tags = [
    "${REGISTRY}/php:${PHP_VERSION}${TAG_SUFFIX}",
    "${REGISTRY}/php:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("cli")
  cache-to   = cache_to("cli")
}

# ---- layer 1a: cli + source-guardian/ioncube loaders ----------------------
target "cli-loaders" {
  inherits   = ["_common"]
  context    = "images/php/cli-loaders/context"
  dockerfile = "../Dockerfile"
  contexts = {
    "${REGISTRY}/php:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}" = "target:cli"
  }
  tags = [
    "${REGISTRY}/php:${PHP_VERSION}-loaders${TAG_SUFFIX}",
    "${REGISTRY}/php:${PHP_VERSION}-loaders-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("cli-loaders")
  cache-to   = cache_to("cli-loaders")
}

# ---- layer 1b: fpm — FROM cli ---------------------------------------------
target "fpm" {
  inherits   = ["_common"]
  context    = "images/php/fpm/context"
  dockerfile = "../Dockerfile"
  contexts = {
    "${REGISTRY}/php:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}" = "target:cli"
  }
  tags = [
    "${REGISTRY}/php:${PHP_VERSION}-fpm${TAG_SUFFIX}",
    "${REGISTRY}/php:${PHP_VERSION}-fpm-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("fpm")
  cache-to   = cache_to("fpm")
}

# ---- layer 2: fpm + loaders — FROM fpm ------------------------------------
target "fpm-loaders" {
  inherits   = ["_common"]
  context    = "images/php/fpm-loaders/context"
  dockerfile = "../Dockerfile"
  contexts = {
    "${REGISTRY}/php:${PHP_VERSION}-fpm-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}" = "target:fpm"
  }
  tags = [
    "${REGISTRY}/php:${PHP_VERSION}-fpm-loaders${TAG_SUFFIX}",
    "${REGISTRY}/php:${PHP_VERSION}-fpm-loaders-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("fpm-loaders")
  cache-to   = cache_to("fpm-loaders")
}

# ---- layer 3: php-fpm base — FROM fpm-loaders (PHP_VARIANT default) --------
target "php-fpm-base" {
  inherits   = ["_common"]
  context    = "images/php-fpm/base/context"
  dockerfile = "../Dockerfile"
  args = {
    PHP_VARIANT = "fpm-loaders"
  }
  contexts = {
    "${REGISTRY}/php:${PHP_VERSION}-fpm-loaders-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}" = "target:fpm-loaders"
  }
  tags = [
    "${REGISTRY}/php-fpm:${PHP_VERSION}${TAG_SUFFIX}",
    "${REGISTRY}/php-fpm:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("php-fpm-base")
  cache-to   = cache_to("php-fpm-base")
}

# ---- layer 4: magento2 app — FROM php-fpm base ----------------------------
target "magento2" {
  inherits   = ["_common"]
  context    = "images/php-fpm/magento2/context"
  dockerfile = "../Dockerfile"
  args = {
    IMAGE_NAME = "${REGISTRY}/php-fpm" # this layer's FROM is php-fpm, not php
  }
  contexts = {
    "${REGISTRY}/php-fpm:${PHP_VERSION}-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}" = "target:php-fpm-base"
  }
  tags = [
    "${REGISTRY}/php-fpm:${PHP_VERSION}-magento2${TAG_SUFFIX}",
    "${REGISTRY}/php-fpm:${PHP_VERSION}-magento2-${BASE_IMAGE_NAME}-${BASE_IMAGE_TAG}${TAG_SUFFIX}",
  ]
  cache-from = cache_from("magento2")
  cache-to   = cache_to("magento2")
}

# Build the whole magento2 dependency chain in one invocation. BuildKit
# resolves the DAG from the `contexts` above and builds shared layers once.
group "magento2-chain" {
  targets = ["cli", "fpm", "fpm-loaders", "php-fpm-base", "magento2"]
}

# The cli loaders branch (independent of the fpm chain).
group "loaders" {
  targets = ["cli-loaders", "fpm-loaders"]
}

group "default" {
  targets = ["magento2-chain"]
}

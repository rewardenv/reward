# PHP image builds with `docker buildx bake` (prototype)

Prototype of a simpler, more reliable way to build the PHP image matrix.
Scope: the `magento2` leaf and its dependency chain. Files: `docker-bake.hcl`,
`.github/workflows/php-bake.yml`. Nothing existing is modified.

## The dependency chain

```
php/cli ──┬─> php/cli-loaders
          └─> php/fpm ─> php/fpm-loaders ─> php-fpm/base ─> php-fpm/magento2
```

`docker-bake.hcl` expresses this as `target`s wired with `contexts`
(`"<image-ref>" = "target:<dep>"`). BuildKit resolves the DAG and builds shared
layers once, in order, in a single `bake` call.

## Run it locally

```bash
# whole chain, default PHP 8.5
docker buildx bake -f docker-bake.hcl magento2-chain

# a specific version
PHP_VERSION=8.4 docker buildx bake -f docker-bake.hcl magento2-chain

# inspect the resolved graph without building
docker buildx bake -f docker-bake.hcl --print magento2-chain
```

Render the Dockerfiles first (the CI workflow does this automatically):

```bash
for d in php/cli php/fpm php/fpm-loaders php-fpm/base php-fpm/magento2; do
  IMAGE_NAME=ubuntu IMAGE_TAG=jammy gomplate -f "images/$d/tpl.Dockerfile" -o "images/$d/Dockerfile"
done
```

## What it replaces

| Concern | Current | This prototype |
|---|---|---|
| Orchestration | `00..05-chain`, 6 sequential `workflow-dispatch` stages, 3h poll-waits, barrier jobs | one `bake` call; BuildKit orders the DAG |
| Files | ~50 PHP workflow files | `docker-bake.hcl` + 1 workflow |
| Version list | duplicated in 18 files | one input in the workflow |
| Layer cache | commented out / weak `type=inline` | `type=registry,mode=max` |
| Retry | copy-pasted duplicate step + `sleep 60`, 1 try | native bash retry loop, 3x backoff |
| One version failing | cancels siblings (`fail-fast` defaults true) | `fail-fast: false` |
| arm64 | QEMU emulation (slow, flaky) | **native `ubuntu-24.04-arm` runner** + manifest merge |

## gomplate is kept (on purpose)

gomplate is **not** removed. Templates like `php/cli` and `php-fpm/base` use
`{{ if eq $IMAGE_NAME "ubuntu" }} … {{ else if eq $IMAGE_NAME "debian" }}` for
distro-specific package logic (e.g. the `ondrej/php` PPA only on Ubuntu). That
is real build logic. The workflow renders `tpl.Dockerfile -> Dockerfile` with
gomplate, then bake builds the result.

Dropping gomplate is a *separate, optional* refactor: move each distro `if`
block into a shell `case "$BASE_IMAGE_NAME"` inside the `RUN` step, driven by a
build-arg. Not required for this approach.

## Native ARM builds (implemented)

`arm64` is built on a **native `ubuntu-24.04-arm` runner**, not QEMU. The
workflow has two jobs:

1. **build** — matrix `php_version × {amd64 on ubuntu-24.04, arm64 on
   ubuntu-24.04-arm}`. Each builds its arch with `PLATFORMS=linux/<arch>` and
   `TAG_SUFFIX=-<arch>`, pushing per-arch tags (`php-fpm:8.5-magento2-amd64`, …).
   The DAG `contexts` wire target→target *inside* each single-arch run, so the
   suffix does not affect FROM resolution.
2. **merge** — reads the canonical (unsuffixed) tags straight from the bake file
   (`bake --print | jq`, single source of truth) and runs
   `docker buildx imagetools create -t <tag> <tag>-amd64 <tag>-arm64` for every
   layer in the chain.

`imagetools create` works against the registry directly, so the merge job needs
no builder or QEMU.

## Not covered by the prototype (would be needed for full migration)

- The other apps (`shopware`, `wordpress`, `magento1`) and the `rootless`,
  `-web`, and utils (`xdebug`, `spx`, `blackfire`) variants — each becomes a
  `target`/`group` in the same bake file.
- The per-arch `<tag>-amd64`/`<tag>-arm64` tags remain in the registry after
  merge; a cleanup step (or push-by-digest instead of suffixed tags) would
  remove that noise.
- Triggers/paths and scheduled builds equivalent to the current chain.

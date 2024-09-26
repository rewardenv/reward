# Contributing Guidelines

## How to build Reward locally

Dependencies:

- golang 1.23
- goreleaser

```
goreleaser --rm-dist --snapshot --config .goreleaser-local.yml
```

This command will build the binaries inside `./dist` directory.

## Use the Reward Test Docker Images

The `develop` branch contains pre-release testing Docker Images.

These are pushed to a separated Docker Repository.

To use them with your Reward instance add the following line to `~/.reward.yml`.

```yaml
reward_docker_image_repo: "docker.io/rewardtest"
```

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1] - 2023-02-07

### Improvements

- Fix an issue with disabled syncing. (#34)

## [0.4.0] - 2023-02-06

We're super excited to announce `Reward` 0.4.0!

### Major changes

This release is a major milestone for the project, and we're proud to share it with you.
The project was rewritten from scratch so this release can contain unexpected bugs and breaking changes.

## [0.4.0-beta3] - 2023-02-03

### New Features

- Add `reward info` command to show information about the current Reward configuration
- Add plugin install support
- Add support for `GITHUB_TOKEN` environment variable used for GitHub API requests.

### Improvements

- Fix linting issues, minor bugs, and improve documentation
- Bootstrap functions are revamped to be more modular
- Use GitHub API to self-update and plugin install

## [0.4.0-beta2] - 2023-02-01

### Improvements

- Add new versions for Elasticsearch, OpenSearch, MariaDB, Redis
- Enable ARM build for Windows
- Rename module from reward to github.com/rewardenv/reward
- Move main.go to cmd/reward
- Enable a bunch of linters and fix all the issues.
- Extract util to a separate package.

## [0.4.0-beta1] - 2023-01-31

### New Features

- Introducing a new **plugin system**. You can now write your own plugins and use them with `Reward`. For more
  information, see the [sample plugin repository](https://github.com/rewardenv/reward-plugin-template).
- Introducing a new feature called **shortcuts**. You can define your own shortcuts to automate `Reward` commands.
- Add support for **bootstrapping Shopware**.
- Allow **self-updating to pre-released** versions. Use `reward self-update --prerelease` to update to the latest
  pre-release version.
- You can now specify the exact version of Composer using `COMPOSER_VERSION=2.4.4` in your `.env` file. Or you can use
  `COMPOSER_VERSION=2` to always use the latest version.

### Improvements

- Some varying **command line flags** are now **only shown for the relevant environment types**. For example for
  the `bootstrap` command the `--magento-type` flag is only available when the environment type is `magento`.
- Requirements (docker API access, version requirements, etc.) are now checked before running any command instead of
  only running before specific commands.
- Default Magento version is now v2.4.5-p1.
- Default Node version is now 16 globally. You can change it by setting `NODE_VERSION` in your `.env` file.
- During self-update now using GitHub API to fetch the latest release instead of GitHub releases page.

### Breaking changes

- `REWARD_MUTAGEN_ENABLED` option is removed. Use `REWARD_SYNC_ENABLED` instead.
- `REWARD_WSL2_DIRECT_MOUNT` option is removed. Use `REWARD_SYNC_ENABLED=false` instead.
- Instead of using `1` and `0` for enabling and disabling options, use `true` and `false` everywhere.

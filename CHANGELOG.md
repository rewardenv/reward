# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.2] - 2025-01-08
### Fixed

- Fix regex for Magento version checking

## [0.7.1] - 2024-12-10
### Fixed

- Use regexes to make sure docker filters only match exact names
- Add BACKEND_PORT to varnish if using single web container

## [0.7.0] - 2024-12-04
### Changed

- Use Ubuntu Jammy as the default base image for php images

## [0.6.14] - 2024-11-19
### Fixed

- Fix version checking for docker compose (#74)

## [0.6.13] - 2024-10-18

### Changed

- Add support for KDE Neon (#72)

## [0.6.12] - 2024-10-18

### Changed

- Add support for KDE Neon (#72)
- Update goreleaser version

### Fix

- Issue with single web containers

## [0.6.11] - 2024-09-26

### Changed

- Convert Magento version to semver before comparison (#71)

## [0.6.10] - 2024-09-26

### Changed

- Update to go version 1.23

## [0.6.9] - 2024-07-26

### Changed

- Update Shopware build scripts
- Fix Elasticsearch/Opensearch version handling for Magento 2.4.7+

## [0.6.8] - 2024-06-02

### Changed

- Update default Magento version to v2.4.7-p1.
- Update default Shopware version to v6.6.3.1.
- Revamp error handling to use errorspkg
- Use inline error-handling where possible
- Update go deps
- Use gcr mirror for image build
- Use prebuilt supervisord image instead of building it inline
- Change varnish and php-fpm health check methods
- Use a dedicated `OPENSEARCH_INITIAL_ADMIN_PASSWORD` for Opensearch >= 2.12.0
- Update default dependencies versions for Magento, Shopware, Wordpress, etc.
- Better support for Mariadb 11+

## [0.6.6] - 2024-05-14

### Changed

- Enable reward-cloud plugin
- Fix opensearch-dashboard host

### Fixed

- Minor bug fixes

## [0.6.5] - 2024-05-10

### Changed

- Simplify the logic of suppressing output/catching output of wrapped commands.

### Fixed

- Fix an issue where the common services were not added to already existing service networks after `reward svc restart`.

## [0.6.4] - 2024-05-10

### Added

- Add support to add custom headers to requests proxied by Traefik. You can now add custom headers to requests proxied
  by Traefik. Add `REWARD_TRAEFIK_CUSTOM_HEADERS` to your `.env` file to configure custom headers.
- Add support to customize Traefik log level. You can now customize Traefik log level. Add `reward_traefik_log_level` to
  the `reward.yml` file to configure the log level.

### Changed

- Update templating to use dollar prefixed contextual variables.

### Fixed

- Fix an issue with opening multiple custom http/s ports.

## [0.6.3] - 2024-04-08

### Changed

- Fix an issue with SPX_KEY variables, set default SPX_KEY to `dev`.

## [0.6.2] - 2024-04-07

### Added

- Use mailpit instead of mailhog ( https://mailbox.reward.test )

## [0.6.1] - 2024-03-27

### Added

- Add support of SPX

### Changed

- Update to go version 1.22

## [0.6.0] - 2024-03-26

### Changed

- Add support of docker compose v2
- Drop support of docker-compose legacy
- Preparations to support podman

## [0.5.3] - 2024-01-22

### Changed

- Update default Magento version to v2.4.6-p3.
- Specify Composer v2.2.22 for Magento initialization

## [0.5.2] - 2023-12-17

### Added

- Configure exact Composer version from `.env` file.
- Use ssmtp to send emails from PHP containers instead of mhsendmail.
- Add option to configure MAGE_RUN_CODE and MAGE_RUN_TYPE from `.env` file.

### Changed

- Update Magento install command to support customizing cache options.

### Fixed

- Read template overrides from reward home directory properly (thanks @moke13-dev)

## [0.5.1] - 2023-11-02

### Changed

- Add `TRAEFIK_EXTRA_HOSTS` from `.env` to traefik labels.

## [0.5.0] - 2023-09-25

### Major changes

- Add experimental support for rootless `php-fpm` containers.

## [0.4.10] - 2023-09-20

### Changed

- Default Magento version is now v2.4.6-p2.
- Change default PHP versions to 8.2.

### Fixed

- Fix an issue with patch version comparisons.

## [0.4.9] - 2023-09-19

### Fixed

- Fix an issue with disabling Magento 2 Two Factor Authentication for 2.4.6+ versions.

## [0.4.8] - 2023-04-29

### Changed

- Add `host.docker.internal` to php-fpm and php-debug `extra_hosts` to support this host name on linux native docker.

## [0.4.7] - 2023-04-18

### Changed

- Reverted default dnsmasq listen address back to `127.0.0.1` instead of `0.0.0.0` as it caused issues for some users.

## [0.4.6] - 2023-03-20

### Added

- Above Magento 2.4.6 when you run `reward bootstrap` with `--disable-tfa` flag Reward will also disable Adobe IMS.

## [0.4.5] - 2023-03-08

### Fixed

- Fix an issue with the default shell container of PWA-Studio environment.

## [0.4.4] - 2023-03-08

### Fixed

- Fix an issue with PWA-Studio environment.

## [0.4.3] - 2023-03-07

### Changed

- Change the default listen address for traefik, dnsmasq and tunnel to `0.0.0.0` instead of `127.0.0.1`. This
  fixes: https://github.com/docker/for-win/issues/13182

### Added

- Add option to configure traefik, dnsmasq and tunnel listen interfaces and ports.

## [0.4.2] - 2023-02-20

### Added

- Add option to disable HTTP -> HTTPS redirection globally. Add `reward_traefik_allow_http=true` to your `.env` file to
  disable.

## [0.4.1] - 2023-02-07

### Fixed

- Fix an issue with disabled syncing. (#34)
- Fix an issue with self-update.

## [0.4.0] - 2023-02-06

We're super excited to announce `Reward` 0.4.0!

### Major changes

This release is a major milestone for the project, and we're proud to share it with you.
The project was rewritten from scratch so this release can contain unexpected bugs and breaking changes.

## [0.4.0-beta3] - 2023-02-03

### Added

- Add `reward info` command to show information about the current Reward configuration
- Add plugin install support
- Add support for `GITHUB_TOKEN` environment variable used for GitHub API requests.

### Changed

- Fix linting issues, minor bugs, and improve documentation
- Bootstrap functions are revamped to be more modular
- Use GitHub API to self-update and plugin install

## [0.4.0-beta2] - 2023-02-01

### Changed

- Add new versions for Elasticsearch, OpenSearch, MariaDB, Redis
- Enable ARM build for Windows
- Rename module from reward to github.com/rewardenv/reward
- Move main.go to cmd/reward
- Enable a bunch of linters and fix all the issues.
- Extract util to a separate package.

## [0.4.0-beta1] - 2023-01-31

### Added

- Introducing a new **plugin system**. You can now write your own plugins and use them with `Reward`. For more
  information, see the [sample plugin repository](https://github.com/rewardenv/reward-plugin-template).
- Introducing a new feature called **shortcuts**. You can define your own shortcuts to automate `Reward` commands.
- Add support for **bootstrapping Shopware**.
- Allow **self-updating to pre-released** versions. Use `reward self-update --prerelease` to update to the latest
  pre-release version.
- You can now specify the exact version of Composer using `COMPOSER_VERSION=2.4.4` in your `.env` file. Or you can use
  `COMPOSER_VERSION=2` to always use the latest version.

### Changed

- Some varying **command line flags** are now **only shown for the relevant environment types**. For example for
  the `bootstrap` command the `--magento-type` flag is only available when the environment type is `magento`.
- Requirements (docker API access, version requirements, etc.) are now checked before running any command instead of
  only running before specific commands.
- Default Magento version is now v2.4.5-p1.
- Default Node version is now 16 globally. You can change it by setting `NODE_VERSION` in your `.env` file.
- During self-update now using GitHub API to fetch the latest release instead of GitHub releases page.

### Removed

- `REWARD_MUTAGEN_ENABLED` option is removed. Use `REWARD_SYNC_ENABLED` instead.
- `REWARD_WSL2_DIRECT_MOUNT` option is removed. Use `REWARD_SYNC_ENABLED=false` instead.
- Instead of using `1` and `0` for enabling and disabling options, use `true` and `false` everywhere.

package logic

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/pkg/util"
)

// bootstrapMagento2 runs a full Magento 2 bootstrap process.
func (c *bootstrapper) bootstrapMagento2() error {
	if !util.AskForConfirmation(
		fmt.Sprintf(
			"Would you like to bootstrap Magento v%s?",
			c.magento2Version().String(),
		),
	) {
		return nil
	}

	log.Printf("Bootstrapping Magento %s...", c.magento2Version().String())

	if err := c.prepare(); err != nil {
		return errors.Wrap(err, "running preparation")
	}

	if err := c.composerPreInstall(); err != nil {
		return errors.Wrap(err, "configuring composer")
	}

	freshInstall, err := c.download()
	if err != nil {
		return errors.Wrap(err, "downloading")
	}

	if err := c.composerInstall(); err != nil {
		return errors.Wrap(err, "running composer install")
	}

	if err := c.composerPostInstall(); err != nil {
		return errors.Wrap(err, "running composer post install configuration")
	}

	adminPassword, err := c.installMagento2(freshInstall)
	if err != nil {
		return errors.Wrap(err, "installing magento 2")
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) minimumMagento2VersionForSearch() *version.Version {
	return version.Must(version.NewVersion("2.3.99"))
}

func (c *bootstrapper) magento2Version() *version.Version {
	v, err := c.MagentoVersion()
	if err != nil {
		log.Panicln(err)
	}

	return v
}

func (c *bootstrapper) magento2VerbosityFlag() string {
	magentoVerbosityFlag := "-v"

	if c.IsDebug() {
		magentoVerbosityFlag += "vv"
	}

	return magentoVerbosityFlag
}

func (c *bootstrapper) installMagento2(freshInstall bool) (string, error) {
	log.Println("Installing Magento...")

	if err := c.installMagento2SetupInstall(); err != nil {
		return "", err
	}

	if err := c.installMagento2ConfigureBasic(); err != nil {
		return "", err
	}

	if err := c.installMagento2ConfigureVarnish(); err != nil {
		return "", err
	}

	if err := c.installMagento2ConfigureSearch(); err != nil {
		return "", err
	}

	if err := c.installMagento2ConfigureDeployMode(); err != nil {
		return "", err
	}

	if err := c.installMagento2ConfigureTFA(); err != nil {
		return "", err
	}

	adminPassword, err := c.installMagento2ConfigureAdminUser()
	if err != nil {
		return "", err
	}

	if err := c.installMagento2DeploySampleData(freshInstall); err != nil {
		return "", err
	}

	if err := c.installMagento2Reindex(); err != nil {
		return "", err
	}

	if err := c.installMagento2ResetAdminURL(); err != nil {
		return "", err
	}

	if err := c.installMagento2FlushCache(); err != nil {
		return "", err
	}

	log.Println("...Magento installed successfully.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento2SetupInstall() error {
	log.Println("Running Magento setup:install...")

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento setup:install %s",
			strings.Join(c.buildMagento2InstallCommand(), " "),
		),
	); err != nil {
		return errors.Wrap(err, "running bin/magento setup:install")
	}

	log.Println("...Magento setup:install finished.")

	return nil
}

func (c *bootstrapper) installMagento2ConfigureDeployMode() error {
	log.Println("Setting Magento deploy mode...")

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento deploy:mode:set -s %s",
			c.MagentoMode(),
		),
	); err != nil {
		return errors.Wrap(err, "running bin/magento deploy:mode:set")
	}

	log.Println("...Magento deploy:mode set.")

	return nil
}

func (c *bootstrapper) installMagento2FlushCache() error {
	log.Println("Flushing cache...")

	if err := c.RunCmdEnvExec("bin/magento cache:flush"); err != nil {
		return errors.Wrap(err, "running bin/magento cache:flush")
	}

	log.Println("...cache flushed.")

	return nil
}

func (c *bootstrapper) installMagento2ResetAdminURL() error {
	if c.ResetAdminURL() {
		log.Println("Resetting admin URL...")

		if err := c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom 0"); err != nil {
			return errors.Wrap(err, "resetting admin url")
		}

		if err := c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom_path 0"); err != nil {
			return errors.Wrap(err, "resetting admin url path")
		}

		log.Println("...admin URL reset.")
	}

	return nil
}

func (c *bootstrapper) installMagento2Reindex() error {
	if c.FullBootstrap() {
		log.Println("Reindexing...")

		if err := c.RunCmdEnvExec("bin/magento indexer:reindex"); err != nil {
			return errors.Wrap(err, "running bin/magento indexer:reindex")
		}

		log.Println("...reindexing complete.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", errors.Wrap(err, "generating admin password")
	}

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			`bin/magento admin:user:create --admin-password=%s `+
				`--admin-user=localadmin --admin-firstname=Local --admin-lastname=Admin --admin-email="admin@example.com"`,
			adminPassword,
		),
	); err != nil {
		return "", errors.Wrap(err, "creating admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento2ConfigureTFA() error {
	minimumMagentoVersionForMFA := version.Must(version.NewVersion("2.3.99"))

	// For Magento 2.4.6 and above, we need to disable the Adobe IMS module as well
	minimumMagentoVersionForMFAAdminAdobeImsTwoFactorAuth := version.Must(version.NewVersion("2.4.5.99"))

	if c.magento2Version().GreaterThan(minimumMagentoVersionForMFA) && c.MagentoDisableTFA() {
		log.Println("Disabling TFA for local development...")

		modules := "Magento_TwoFactorAuth"
		if c.magento2Version().GreaterThan(minimumMagentoVersionForMFAAdminAdobeImsTwoFactorAuth) {
			modules = "{Magento_AdminAdobeImsTwoFactorAuth,Magento_TwoFactorAuth}"
		}

		if err := c.RunCmdEnvExec("bin/magento module:disable " + modules); err != nil {
			return errors.Wrapf(err, "running bin/magento module:disable %v", modules)
		}

		log.Println("...TFA disabled.")
	}

	return nil
}

func (c *bootstrapper) installMagento2DeploySampleData(freshInstall bool) error {
	if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
		log.Println("Installing sample data...")

		if err := c.RunCmdEnvExec(
			"mkdir -p /var/www/html/var/composer_home/ && " +
				"cp -va ~/.composer/auth.json /var/www/html/var/composer_home/auth.json",
		); err != nil {
			return errors.Wrap(err, "copying auth.json")
		}

		if err := c.RunCmdEnvExec(
			fmt.Sprintf(
				`php bin/magento %s sampledata:deploy`,
				c.magento2VerbosityFlag(),
			),
		); err != nil {
			return errors.Wrap(err, "running bin/magento sampledata:deploy")
		}

		if err := c.RunCmdEnvExec(
			fmt.Sprintf(
				`bin/magento setup:upgrade %s`,
				c.magento2VerbosityFlag(),
			),
		); err != nil {
			return errors.Wrap(err, "running bin/magento setup:upgrade")
		}

		log.Println("...sample data installed successfully.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureSearch() error {
	if c.ServiceEnabled("elasticsearch") || c.ServiceEnabled("opensearch") {
		log.Println("Configuring Elasticsearch/OpenSearch...")

		if err := c.RunCmdEnvExec("bin/magento config:set --lock-env catalog/search/enable_eav_indexer 1"); err != nil {
			return errors.Wrap(err, "enabling eav indexer")
		}

		searchHost, searchEngine := c.buildMagentoSearchHost()

		if c.magento2Version().GreaterThan(c.minimumMagento2VersionForSearch()) {
			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/engine %s",
					searchEngine,
				),
			); err != nil {
				return errors.Wrap(err, "setting magento search engine")
			}

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_hostname %s",
					searchEngine,
					searchHost,
				),
			); err != nil {
				return errors.Wrap(err, "setting magento search engine server")
			}

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_port 9200",
					searchEngine,
				),
			); err != nil {
				return errors.Wrap(err, "setting magento search engine port")
			}

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_index_prefix magento2",
					searchEngine,
				),
			); err != nil {
				return errors.Wrap(err, "setting magento search engine index prefix")
			}

			// Enable auth if OpenSearch is enabled and version is 2.12.0 or above
			openSearchInitialAdminPassword := c.GetString("OPENSEARCH_INITIAL_ADMIN_PASSWORD")
			if openSearchInitialAdminPassword != "" {
				if err := c.RunCmdEnvExec(
					fmt.Sprintf(
						"bin/magento config:set --lock-env catalog/search/%s_enable_auth 1",
						searchEngine,
					),
				); err != nil {
					return errors.Wrap(err, "disabling magento search engine auth")
				}

				if err := c.RunCmdEnvExec(
					fmt.Sprintf(
						"bin/magento config:set --lock-env catalog/search/%s_username admin",
						searchEngine,
					),
				); err != nil {
					return errors.Wrap(err, "disabling magento search engine auth")
				}

				if err := c.RunCmdEnvExec(
					fmt.Sprintf(
						"bin/magento config:set --lock-env catalog/search/%s_password %s",
						searchEngine,
						openSearchInitialAdminPassword,
					),
				); err != nil {
					return errors.Wrap(err, "disabling magento search engine auth")
				}
			} else {
				if err := c.RunCmdEnvExec(
					fmt.Sprintf(
						"bin/magento config:set --lock-env catalog/search/%s_enable_auth 0",
						searchEngine,
					),
				); err != nil {
					return errors.Wrap(err, "disabling magento search engine auth")
				}
			}

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_timeout 15",
					searchEngine,
				),
			); err != nil {
				return errors.Wrap(err, "setting magento search engine timeout")
			}
		}

		log.Println("...Elasticsearch/Opensearch configured.")
	}

	return nil
}

func (c *Client) installMagento2ConfigureVarnish() error {
	if c.ServiceEnabled("varnish") {
		log.Println("Configuring Varnish...")

		if err := c.RunCmdEnvExec(
			"bin/magento config:set --lock-env system/full_page_cache/caching_application 2",
		); err != nil {
			return errors.Wrap(err, "configuring magento varnish")
		}

		if err := c.RunCmdEnvExec("bin/magento config:set --lock-env system/full_page_cache/ttl 604800"); err != nil {
			return errors.Wrap(err, "configuring magento varnish cache ttl")
		}

		log.Println("...Varnish configured.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureBasic() error {
	log.Println("Configuring Magento basic settings...")

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	); err != nil {
		return errors.Wrap(err, "setting magento base URL")
	}

	// Set secure base URL
	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	); err != nil {
		return errors.Wrap(err, "setting magento secure base URL")
	}

	// Set offload header
	if err := c.RunCmdEnvExec(
		"bin/magento config:set --lock-env web/secure/offloader_header X-Forwarded-Proto",
	); err != nil {
		return errors.Wrap(err, "setting magento offload header")
	}

	// Set use https in frontend
	if err := c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_frontend 1"); err != nil {
		return errors.Wrap(err, "setting magento use https on frontend")
	}

	// Set use https in admin
	if err := c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_adminhtml 1"); err != nil {
		return errors.Wrap(err, "setting magento use https on admin")
	}

	// Set seo rewrites
	if err := c.RunCmdEnvExec("bin/magento config:set web/seo/use_rewrites 1"); err != nil {
		return errors.Wrap(err, "setting magento seo rewrites")
	}

	log.Println("...Magento basic settings configured.")

	return nil
}

func (c *bootstrapper) buildMagento2InstallCommand() []string {
	magentoCmdParams := []string{
		"--backend-frontname=" + c.MagentoBackendFrontname(),
		"--db-host=db",
		"--db-name=magento",
		"--db-user=magento",
		"--db-password=magento",
	}

	if c.DBPrefix() != "" {
		magentoCmdParams = append(magentoCmdParams, fmt.Sprintf("--db-prefix=%s", c.DBPrefix()))
	}

	if c.CryptKey() != "" {
		magentoCmdParams = append(magentoCmdParams, fmt.Sprintf("--key=%s", c.CryptKey()))
	}

	// Redis configuration
	if c.ServiceEnabled("redis") {
		magentoCmdParams = append(
			magentoCmdParams,
			"--session-save=redis",
			"--session-save-redis-host=redis",
			"--session-save-redis-port=6379",
			"--session-save-redis-db=2",
			"--session-save-redis-max-concurrency=20",
			"--cache-backend=redis",
			"--cache-backend-redis-server=redis",
			"--cache-backend-redis-db=0",
			"--cache-backend-redis-port=6379",
			"--page-cache=redis",
			"--page-cache-redis-server=redis",
			"--page-cache-redis-db=1",
			"--page-cache-redis-port=6379",
		)
	} else {
		magentoCmdParams = append(magentoCmdParams, "--session-save=files")
	}

	// Varnish configuration
	if c.ServiceEnabled("varnish") {
		magentoCmdParams = append(magentoCmdParams, "--http-cache-hosts=varnish:80")
	}

	// RabbitMQ configuration
	if c.ServiceEnabled("rabbitmq") {
		magentoCmdParams = append(
			magentoCmdParams,
			"--amqp-host=rabbitmq",
			"--amqp-port=5672",
			"--amqp-user=guest",
			"--amqp-password=guest",
		)

		minimumVersionForRabbitMQWait := version.Must(version.NewVersion("2.3.99"))
		if c.magento2Version().GreaterThan(minimumVersionForRabbitMQWait) {
			magentoCmdParams = append(magentoCmdParams, "--consumers-wait-for-messages=0")
		}
	}

	searchHost, searchEngine := c.buildMagentoSearchHost()

	if c.ServiceEnabled("elasticsearch") ||
		c.ServiceEnabled("opensearch") &&
			c.magento2Version().GreaterThan(c.minimumMagento2VersionForSearch()) {
		magentoCmdParams = append(
			magentoCmdParams,
			fmt.Sprintf("--search-engine=%s", searchEngine),
			fmt.Sprintf("--elasticsearch-host=%s", searchHost),
			"--elasticsearch-port=9200",
			"--elasticsearch-index-prefix=magento2",
			"--elasticsearch-enable-auth=0",
			"--elasticsearch-timeout=15",
		)
	}

	return magentoCmdParams
}

func (c *bootstrapper) buildMagentoSearchHost() (string, string) {
	// Elasticsearch/OpenSearch configuration
	searchHost, searchEngine := "", ""

	switch {
	case c.ServiceEnabled("opensearch"):
		searchHost = "opensearch"
		// Need to specify elasticsearch7 for opensearch too
		// https://devdocs.magento.com/guides/v2.4/install-gde/install/cli/install-cli.html
		searchEngine = "elasticsearch7"

	case c.ServiceEnabled("elasticsearch"):
		searchHost = "elasticsearch"
		searchEngine = "elasticsearch7"
	}

	return searchHost, searchEngine
}

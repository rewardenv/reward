package logic

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/pkg/util"
)

// bootstrapMagento2 runs a full Magento 2 bootstrap process.
func (c *bootstrapper) bootstrapMagento2() error {
	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Magento v%s?",
		c.magento2Version().String())) {
		return nil
	}

	log.Printf("Bootstrapping Magento %s...", c.magento2Version().String())

	err := c.prepare()
	if err != nil {
		return fmt.Errorf("error during preparation: %w", err)
	}

	err = c.composerPreInstall()
	if err != nil {
		return fmt.Errorf("error during composer configuration: %w", err)
	}

	freshInstall, err := c.download()
	if err != nil {
		return fmt.Errorf("error during download: %w", err)
	}

	err = c.composerInstall()
	if err != nil {
		return fmt.Errorf("error during composer install: %w", err)
	}

	err = c.composerPostInstall()
	if err != nil {
		return fmt.Errorf("error during composer post install configuration: %w", err)
	}

	adminPassword, err := c.installMagento2(freshInstall)
	if err != nil {
		return fmt.Errorf("error during magento 2 installation: %w", err)
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) minimumMagento2VersionForSearch() *version.Version {
	return version.Must(version.NewVersion("2.4.0"))
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

	err := c.installMagento2SetupInstall()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ConfigureBasic()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ConfigureVarnish()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ConfigureSearch()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ConfigureDeployMode()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ConfigureTFA()
	if err != nil {
		return "", err
	}

	adminPassword, err := c.installMagento2ConfigureAdminUser()
	if err != nil {
		return "", err
	}

	err = c.installMagento2DeploySampleData(freshInstall)
	if err != nil {
		return "", err
	}

	err = c.installMagento2Reindex()
	if err != nil {
		return "", err
	}

	err = c.installMagento2ResetAdminURL()
	if err != nil {
		return "", err
	}

	err = c.installMagento2FlushCache()
	if err != nil {
		return "", err
	}

	log.Println("...Magento installed successfully.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento2SetupInstall() error {
	log.Println("Running Magento setup:install...")

	err := c.RunCmdEnvExec(fmt.Sprintf("bin/magento setup:install %s",
		strings.Join(c.buildMagento2InstallCommand(), " ")))
	if err != nil {
		return fmt.Errorf("cannot run bin/magento setup:install: %w", err)
	}

	log.Println("...Magento setup:install finished.")

	return nil
}

func (c *bootstrapper) installMagento2ConfigureDeployMode() error {
	log.Println("Setting Magento deploy mode...")

	err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento deploy:mode:set -s %s",
			c.MagentoMode(),
		),
	)
	if err != nil {
		return fmt.Errorf("cannot run bin/magento deploy:mode:set: %w", err)
	}

	log.Println("...Magento deploy:mode set.")

	return nil
}

func (c *bootstrapper) installMagento2FlushCache() error {
	log.Println("Flushing cache...")

	err := c.RunCmdEnvExec("bin/magento cache:flush")
	if err != nil {
		return fmt.Errorf("cannot run bin/magento cache:flush: %w", err)
	}

	log.Println("...cache flushed.")

	return nil
}

func (c *bootstrapper) installMagento2ResetAdminURL() error {
	if c.ResetAdminURL() {
		log.Println("Resetting admin URL...")

		err := c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom 0")
		if err != nil {
			return fmt.Errorf("cannot reset admin url: %w", err)
		}

		err = c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom_path 0")
		if err != nil {
			return fmt.Errorf("cannot reset admin url path: %w", err)
		}

		log.Println("...admin URL reset.")
	}

	return nil
}

func (c *bootstrapper) installMagento2Reindex() error {
	if c.FullBootstrap() {
		log.Println("Reindexing...")

		err := c.RunCmdEnvExec("bin/magento indexer:reindex")
		if err != nil {
			return fmt.Errorf("cannot run bin/magento indexer:reindex: %w", err)
		}

		log.Println("...reindexing complete.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", fmt.Errorf("cannot generate admin password: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			`bin/magento admin:user:create --admin-password=%s `+
				`--admin-user=localadmin --admin-firstname=Local --admin-lastname=Admin --admin-email="admin@example.com"`,
			adminPassword,
		),
	)
	if err != nil {
		return "", fmt.Errorf("cannot create admin user: %w", err)
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento2ConfigureTFA() error {
	minimumMagentoVersionForMFA := version.Must(version.NewVersion("2.4.0"))

	if c.magento2Version().GreaterThan(minimumMagentoVersionForMFA) && c.MagentoDisableTFA() {
		log.Println("Disabling TFA for local development...")

		err := c.RunCmdEnvExec("bin/magento module:disable Magento_TwoFactorAuth")
		if err != nil {
			return fmt.Errorf("cannot run bin/magento module:disable Magento_TwoFactorAuth: %w", err)
		}

		log.Println("...TFA disabled.")
	}

	return nil
}

func (c *bootstrapper) installMagento2DeploySampleData(freshInstall bool) error {
	if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
		log.Println("Installing sample data...")

		err := c.RunCmdEnvExec("mkdir -p /var/www/html/var/composer_home/ && " +
			"cp -va ~/.composer/auth.json /var/www/html/var/composer_home/auth.json",
		)
		if err != nil {
			return fmt.Errorf("cannot copy auth.json: %w", err)
		}

		err = c.RunCmdEnvExec(
			fmt.Sprintf(
				`php bin/magento %s sampledata:deploy`,
				c.magento2VerbosityFlag(),
			),
		)
		if err != nil {
			return fmt.Errorf("cannot run bin/magento sampledata:deploy: %w", err)
		}

		err = c.RunCmdEnvExec(
			fmt.Sprintf(
				`bin/magento setup:upgrade %s`,
				c.magento2VerbosityFlag(),
			),
		)
		if err != nil {
			return fmt.Errorf("cannot run bin/magento setup:upgrade: %w", err)
		}

		log.Println("...sample data installed successfully.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureSearch() error {
	if c.ServiceEnabled("elasticsearch") || c.ServiceEnabled("opensearch") {
		log.Println("Configuring Elasticsearch/OpenSearch...")

		err := c.RunCmdEnvExec("bin/magento config:set --lock-env catalog/search/enable_eav_indexer 1")
		if err != nil {
			return fmt.Errorf("cannot enable eav indexer: %w", err)
		}

		searchHost, searchEngine := c.buildMagentoSearchHost()

		if c.magento2Version().GreaterThan(c.minimumMagento2VersionForSearch()) {
			err = c.RunCmdEnvExec(
				fmt.Sprintf("bin/magento config:set --lock-env catalog/search/engine %s",
					searchEngine,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot set magento search engine: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_hostname %s",
					searchEngine,
					searchHost,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot set magento search engine server: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_port 9200",
					searchEngine,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot set magento search engine port: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_index_prefix magento2",
					searchEngine,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot set magento search engine index prefix: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_enable_auth 0",
					searchEngine,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot disable magento search engine auth: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_timeout 15",
					searchEngine,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot set magento search engine timeout: %w", err)
			}
		}

		log.Println("...Elasticsearch/Opensearch configured.")
	}

	return nil
}

func (c *Client) installMagento2ConfigureVarnish() error {
	if c.ServiceEnabled("varnish") {
		log.Println("Configuring Varnish...")

		err := c.RunCmdEnvExec("bin/magento config:set --lock-env system/full_page_cache/caching_application 2")
		if err != nil {
			return fmt.Errorf("cannot configure magento varnish: %w", err)
		}

		err = c.RunCmdEnvExec("bin/magento config:set --lock-env system/full_page_cache/ttl 604800")
		if err != nil {
			return fmt.Errorf("cannot configure magento varnish cache ttl: %w", err)
		}

		log.Println("...Varnish configured.")
	}

	return nil
}

func (c *bootstrapper) installMagento2ConfigureBasic() error {
	log.Println("Configuring Magento basic settings...")

	err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("cannot set magento base URL: %w", err)
	}

	// Set secure base URL
	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("cannot set magento secure base URL: %w", err)
	}

	// Set offload header
	err = c.RunCmdEnvExec("bin/magento config:set --lock-env web/secure/offloader_header X-Forwarded-Proto")
	if err != nil {
		return fmt.Errorf("cannot set magento offload header: %w", err)
	}

	// Set use https in frontend
	err = c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_frontend 1")
	if err != nil {
		return fmt.Errorf("cannot set magento use https in frontend: %w", err)
	}

	// Set use https in admin
	err = c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_adminhtml 1")
	if err != nil {
		return fmt.Errorf("cannot set magento use https in admin: %w", err)
	}

	// Set seo rewrites
	err = c.RunCmdEnvExec("bin/magento config:set web/seo/use_rewrites 1")
	if err != nil {
		return fmt.Errorf("cannot set magento seo rewrites: %w", err)
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
		magentoCmdParams = append(magentoCmdParams,
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
		magentoCmdParams = append(magentoCmdParams,
			"--amqp-host=rabbitmq",
			"--amqp-port=5672",
			"--amqp-user=guest",
			"--amqp-password=guest",
		)

		minimumVersionForRabbitMQWait := version.Must(version.NewVersion("2.4.0"))
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

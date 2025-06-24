package logic

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/pkg/util"
)

// bootstrap runs a full Magento 2 bootstrap process.
func (c *magento2) bootstrap() error {
	m2versionString := c.semver().String()
	if m2versionString == "0.0.0+undefined" {
		m2versionString = "from existing composer.json"
	}

	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Magento %s?", m2versionString)) {
		return nil
	}

	log.Printf("Bootstrapping Magento %s...", m2versionString)

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

	adminPassword, err := c.install(freshInstall)
	if err != nil {
		return errors.Wrap(err, "installing magento 2")
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())

	if adminPassword != "" {
		log.Println("Admin user: localadmin")
		log.Printf("Admin password: %s", adminPassword)
	}

	log.Println("...bootstrap process finished.")

	return nil
}

func (c *magento2) version() *version.Version {
	v, err := c.MagentoVersion()
	if err != nil {
		log.Panicln(err)
	}

	return v
}

func (c *magento2) semver() *version.Version {
	v := c.version()

	return util.ConvertVersionPrereleaseToMetadata(v)
}

func (c *magento2) download() (downloaded bool, err error) {
	if c.SkipComposerInstall() {
		return false, nil
	}

	if util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
		return false, nil
	}

	log.Println("Creating Magento 2 composer project...")

	magentoVersion, err := c.MagentoVersion()
	if err != nil {
		return false, errors.Wrap(err, "determining magento version")
	}

	command := fmt.Sprintf("%s create-project --profile --no-install "+
		"--repository-url=https://repo.magento.com/ "+
		"magento/project-%s-edition=%s /tmp/magento-tmp/",
		c.composerCommand(),
		c.MagentoType(),
		magentoVersion.String(),
	)
	if err = c.RunCmdEnvExec(command); err != nil {
		return false, errors.Wrap(err, "creating composer magento project")
	}

	command = fmt.Sprintf(
		`%s -au --remove-source-files --chmod=D2775,F644 /tmp/magento-tmp/ /var/www/html/`,
		c.rsyncCommand(),
	)
	if err = c.RunCmdEnvExec(command); err != nil {
		return false, errors.Wrap(err, "moving magento project install files")
	}

	log.Println("...Magento 2 composer project created.")

	return true, nil
}

func (c *magento2) install(freshInstall bool) (string, error) {
	log.Println("Installing Magento...")

	if err := c.setupInstall(); err != nil {
		return "", err
	}

	if err := c.configureBasicSettings(); err != nil {
		return "", err
	}

	if err := c.configureVarnish(); err != nil {
		return "", err
	}

	if err := c.configureSearch(); err != nil {
		return "", err
	}

	if err := c.configureDeployMode(); err != nil {
		return "", err
	}

	if err := c.disableTFA(); err != nil {
		return "", err
	}

	adminPassword, err := c.configureAdminUser()
	if err != nil {
		return "", err
	}

	if err := c.deploySampleData(freshInstall); err != nil {
		return "", err
	}

	if err := c.reindex(); err != nil {
		return "", err
	}

	if err := c.resetAdminURL(); err != nil {
		return "", err
	}

	if err := c.flushCache(); err != nil {
		return "", err
	}

	log.Println("...Magento installed successfully.")

	return adminPassword, nil
}

func (c *magento2) setupInstall() error {
	log.Println("Running Magento setup:install...")

	command := fmt.Sprintf("%s setup:install %s", c.magentoCommand(), c.setupInstallArgs())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento setup:install")
	}

	log.Println("...Magento setup:install finished.")

	return nil
}

func (c *magento2) configureDeployMode() error {
	log.Println("Setting Magento deploy mode...")

	command := fmt.Sprintf("%s deploy:mode:set -s %s", c.magentoCommand(), c.MagentoMode())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento deploy:mode:set")
	}

	log.Println("...Magento deploy:mode set.")

	return nil
}

func (c *magento2) flushCache() error {
	log.Println("Flushing cache...")

	command := fmt.Sprintf("%s cache:flush", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento cache:flush")
	}

	log.Println("...cache flushed.")

	return nil
}

func (c *magento2) resetAdminURL() error {
	if !c.ResetAdminURL() {
		return nil
	}

	log.Println("Resetting admin URL...")

	command := fmt.Sprintf("%s config:set admin/url/use_custom 0", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "resetting admin url")
	}

	command = fmt.Sprintf("%s config:set admin/url/use_custom_path 0", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "resetting admin url path")
	}

	log.Println("...admin URL reset.")

	return nil
}

func (c *magento2) reindex() error {
	if !c.FullBootstrap() {
		return nil
	}

	log.Println("Reindexing...")

	command := fmt.Sprintf("%s indexer:reindex", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento indexer:reindex")
	}

	log.Println("...reindexing complete.")

	return nil
}

func (c *magento2) configureAdminUser() (string, error) {
	if c.SkipAdminUser() {
		return "", nil
	}

	log.Println("Creating admin user...")

	adminPassword := c.generatePassword()

	command := fmt.Sprintf(`%s admin:user:create --admin-password=%s `+
		`--admin-user=localadmin --admin-firstname=Local `+
		`--admin-lastname=Admin --admin-email="admin@example.com"`,
		c.magentoCommand(),
		adminPassword,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return "", errors.Wrap(err, "creating admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *magento2) disableTFA() error {
	if !c.MagentoDisableTFA() {
		return nil
	}

	force := false

	if c.semver().String() == "0.0.0+undefined" {
		if !util.AskForConfirmation("Magento version cannot be determined. Would you like to disable TFA?") {
			return nil
		}
		force = true
	}

	mfaConstraints := version.MustConstraints(version.NewConstraint(">=2.4"))
	if !force && !mfaConstraints.Check(c.semver()) {
		return nil
	}

	log.Println("Disabling TFA for local development...")

	modules := c.tfaModules()

	command := fmt.Sprintf("%s module:disable %s", c.magentoCommand(), modules)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrapf(err, "running bin/magento module:disable %v", modules)
	}

	log.Println("...TFA disabled.")

	return nil
}

func (c *magento2) tfaModules() string {
	// For Magento 2.4.6 and above, we need to disable the Adobe IMS module as well
	if c.semver().String() == "0.0.0+undefined" {
		if util.AskForConfirmation("Magento version cannot be determined. " +
			"Would you like to disable Adobe IMS module (this is usually required for Magento >=2.4.6)?",
		) {
			return "{Magento_AdminAdobeImsTwoFactorAuth,Magento_TwoFactorAuth}"
		}
	}

	adobeImsConstraints := version.MustConstraints(version.NewConstraint(">=2.4.6"))
	if adobeImsConstraints.Check(c.semver()) {
		return "{Magento_AdminAdobeImsTwoFactorAuth,Magento_TwoFactorAuth}"
	}

	return "Magento_TwoFactorAuth"
}

func (c *magento2) deploySampleData(freshInstall bool) error {
	if !freshInstall || (!c.WithSampleData() && !c.FullBootstrap()) {
		return nil
	}

	log.Println("Installing sample data...")

	command := `mkdir -p /var/www/html/var/composer_home/ && ` +
		`cp -va ~/.composer/auth.json /var/www/html/var/composer_home/auth.json`

	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "copying auth.json")
	}

	command = fmt.Sprintf(`%s sampledata:deploy`, c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento sampledata:deploy")
	}

	command = fmt.Sprintf(`%s setup:upgrade`, c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running bin/magento setup:upgrade")
	}

	log.Println("...sample data installed successfully.")

	return nil
}

func (c *magento2) configureSearch() error {
	if !c.ServiceEnabled("elasticsearch") && !c.ServiceEnabled("opensearch") {
		return nil
	}

	log.Println("Configuring Elasticsearch/OpenSearch...")

	command := c.magentoCommand() + " config:set --lock-env catalog/search/enable_eav_indexer 1"
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "enabling eav indexer")
	}

	searchHost, searchEngine := c.magentoSearchHost()
	force := false

	if c.semver().String() == "0.0.0+undefined" {
		if util.AskForConfirmation("Magento version cannot be determined. " +
			"Would you like to disable Adobe IMS module (this is usually required for Magento >=2.4.6)?",
		) {
			force = true
		}
	}

	var enabled bool
	if c.ServiceEnabled("elasticsearch") || c.ServiceEnabled("opensearch") {
		enabled = true
	}

	// Above Magento 2.4 the search engine must be configured
	constraints := version.MustConstraints(version.NewConstraint(">=2.4"))
	if !force && !enabled && !constraints.Check(c.semver()) {
		return nil
	}

	command = fmt.Sprintf("%s config:set --lock-env catalog/search/engine %s", c.magentoCommand(), searchEngine)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento search engine")
	}

	command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_server_hostname %s",
		c.magentoCommand(),
		searchEngine,
		searchHost,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento search engine server")
	}

	command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_server_port 9200",
		c.magentoCommand(),
		searchEngine,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento search engine port")
	}

	command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_index_prefix magento2",
		c.magentoCommand(),
		searchEngine,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento search engine index prefix")
	}

	// Enable auth if OpenSearch is enabled and version is 2.12.0 or above
	openSearchInitialAdminPassword := c.GetString("OPENSEARCH_INITIAL_ADMIN_PASSWORD")
	if openSearchInitialAdminPassword != "" {
		command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_enable_auth 1",
			c.magentoCommand(),
			searchEngine,
		)
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "disabling magento search engine auth")
		}

		command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_username admin",
			c.magentoCommand(),
			searchEngine,
		)
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "disabling magento search engine auth")
		}

		command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_password %s",
			c.magentoCommand(),
			searchEngine,
			openSearchInitialAdminPassword,
		)
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "disabling magento search engine auth")
		}
	} else {
		command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_enable_auth 0",
			c.magentoCommand(),
			searchEngine,
		)
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "disabling magento search engine auth")
		}
	}

	command = fmt.Sprintf("%s config:set --lock-env catalog/search/%s_server_timeout 15",
		c.magentoCommand(),
		searchEngine,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento search engine timeout")
	}

	log.Println("...Elasticsearch/Opensearch configured.")

	return nil
}

func (c *magento2) configureVarnish() error {
	if !c.ServiceEnabled("varnish") {
		return nil
	}

	log.Println("Configuring Varnish...")

	command := fmt.Sprintf("%s config:set --lock-env system/full_page_cache/caching_application 2", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "configuring magento varnish")
	}

	command = fmt.Sprintf("%s config:set --lock-env system/full_page_cache/ttl 604800", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "configuring magento varnish cache ttl")
	}

	log.Println("...Varnish configured.")

	return nil
}

func (c *magento2) configureBasicSettings() error {
	log.Println("Configuring Magento basic settings...")

	command := fmt.Sprintf("%s config:set web/unsecure/base_url http://%s/", c.magentoCommand(), c.TraefikFullDomain())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento base URL")
	}

	// Set secure base URL
	command = fmt.Sprintf("%s config:set web/secure/base_url https://%s/", c.magentoCommand(), c.TraefikFullDomain())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento secure base URL")
	}

	// Set offload header
	command = fmt.Sprintf("%s config:set --lock-env web/secure/offloader_header X-Forwarded-Proto", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento offload header")
	}

	// Set use https in frontend
	command = fmt.Sprintf("%s config:set web/secure/use_in_frontend 1", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento use https on frontend")
	}

	// Set use https in admin
	command = fmt.Sprintf("%s config:set web/secure/use_in_adminhtml 1", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento use https on admin")
	}

	// Set seo rewrites
	command = fmt.Sprintf("%s config:set web/seo/use_rewrites 1", c.magentoCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento seo rewrites")
	}

	log.Println("...Magento basic settings configured.")

	return nil
}

func (c *magento2) setupInstallArgs() string {
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
	if c.ServiceEnabled("valkey") {
		magentoCmdParams = append(
			magentoCmdParams,
			"--session-save=redis",
			"--session-save-redis-host=valkey",
			"--session-save-redis-port=6379",
			"--session-save-redis-db=2",
			"--session-save-redis-max-concurrency=20",
			"--cache-backend=redis",
			"--cache-backend-redis-server=valkey",
			"--cache-backend-redis-db=0",
			"--cache-backend-redis-port=6379",
			"--page-cache=redis",
			"--page-cache-redis-server=valkey",
			"--page-cache-redis-db=1",
			"--page-cache-redis-port=6379",
		)
	} else if c.ServiceEnabled("redis") {
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

		// --consumers-wait-for-messages option is only available above Magento 2.4
		constraints := version.MustConstraints(version.NewConstraint(">=2.4"))
		if constraints.Check(c.semver()) {
			magentoCmdParams = append(magentoCmdParams, "--consumers-wait-for-messages=0")
		}
	}

	searchHost, searchEngine := c.magentoSearchHost()

	searchEngineFlag := "opensearch"
	if strings.HasPrefix(searchEngine, "elasticsearch") {
		searchEngineFlag = "elasticsearch"
	}

	constraints := version.MustConstraints(version.NewConstraint(">=2.4"))
	if c.ServiceEnabled("elasticsearch") || c.ServiceEnabled("opensearch") && constraints.Check(c.semver()) {
		magentoCmdParams = append(
			magentoCmdParams,
			fmt.Sprintf("--search-engine=%s", searchEngine),
			fmt.Sprintf("--%s-host=%s", searchEngineFlag, searchHost),
			fmt.Sprintf("--%s-port=9200", searchEngineFlag),
			fmt.Sprintf("--%s-index-prefix=magento2", searchEngineFlag),
			fmt.Sprintf("--%s-enable-auth=0", searchEngineFlag),
			fmt.Sprintf("--%s-timeout=15", searchEngineFlag),
		)
	}

	return strings.Join(magentoCmdParams, " ")
}

func (c *magento2) magentoSearchHost() (string, string) {
	// Elasticsearch/OpenSearch configuration
	searchHost, searchEngine := "", ""

	switch {
	case c.ServiceEnabled("opensearch"):
		searchHost = "opensearch"

		// Need to specify elasticsearch7 for opensearch as Magento 2.4.6 and below
		// https://devdocs.magento.com/guides/v2.4/install-gde/install/cli/install-cli.html
		constraints := version.MustConstraints(version.NewConstraint(">=2.4.7"))
		if constraints.Check(c.semver()) {
			log.Println("Setting search engine to openSearch")
			searchEngine = "opensearch"
		} else {
			log.Println("Setting search engine to elasticsearch7")
			searchEngine = "elasticsearch7"
		}

	case c.ServiceEnabled("elasticsearch"):
		searchHost = "elasticsearch"
		searchEngine = "elasticsearch7"
		//
		// For now it's not working with Magento 2.4.7 + Elasticsearch 8
		// constraints := version.MustConstraints(version.NewConstraint(">=8.0, <9.0"))
		// if constraints.Check(c.ElasticsearchVersion()) {
		// 	log.Println("Setting search engine to elasticsearch8")
		// 	searchEngine = "elasticsearch8"
		// }
	}

	return searchHost, searchEngine
}

func (c *magento2) magentoCommand() string {
	verbosity := "-v"

	if c.IsDebug() {
		verbosity += "vv"
	}

	return fmt.Sprintf("php bin/magento %s", verbosity)
}

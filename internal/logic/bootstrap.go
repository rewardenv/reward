package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-version"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"

	"reward/internal/templates"
	"reward/internal/util"
)

type bootstrapper struct {
	*Client
	composerVerbosityFlag string
	debug                 bool
}

func newBootstrapper(c *Client) *bootstrapper {
	var composerVerbosityFlag = "--verbose"
	if c.IsDebug() {
		composerVerbosityFlag = "-vvv"
	}

	return &bootstrapper{
		Client:                c,
		composerVerbosityFlag: composerVerbosityFlag,
		debug:                 c.IsDebug(),
	}
}

// RunCmdBootstrap represents the bootstrap command.
func (c *Client) RunCmdBootstrap() error {
	switch c.EnvType() {
	case "magento2":
		err := newBootstrapper(c).bootstrapMagento2()
		if err != nil {
			return fmt.Errorf("error bootstrapping magento2: %w", err)
		}
	case "magento1":
		err := newBootstrapper(c).bootstrapMagento1()
		if err != nil {
			return fmt.Errorf("error bootstrapping magento1: %w", err)
		}
	case "wordpress":
		err := newBootstrapper(c).bootstrapWordpress()
		if err != nil {
			return fmt.Errorf("error bootstrapping wordpress: %w", err)
		}
	case "shopware":
		err := newBootstrapper(c).bootstrapShopware()
		if err != nil {
			return fmt.Errorf("error bootstrapping shopware: %w", err)
		}
	default:
		return fmt.Errorf("currently not supported for bootstrapping")
	}

	return nil
}

func (c *bootstrapper) prepare() error {
	log.Println("Preparing common services...")

	err := c.RunCmdSvc([]string{"up"})
	if err != nil {
		return fmt.Errorf("cannot start services: %w", err)
	}

	log.Println("...common services started.")
	log.Println("Preparing certificate...")

	err = c.RunCmdSignCertificate([]string{c.TraefikDomain()}, true)
	if err != nil {
		return fmt.Errorf("cannot sign certificate: %w", err)
	}

	log.Println("...certificate ready.")

	if !c.NoPull() {
		log.Println("Pulling images...")

		err = c.RunCmdEnv([]string{"pull"})
		if err != nil {
			return fmt.Errorf("cannot pull env containers: %w", err)
		}

		log.Println("...images pulled.")
	}

	log.Println("Preparing environment...")

	err = c.RunCmdEnv([]string{"build"})
	if err != nil {
		return fmt.Errorf("cannot build env containers: %w", err)
	}

	err = c.RunCmdEnv([]string{"up"})
	if err != nil {
		return fmt.Errorf("cannot start env containers: %w", err)
	}

	log.Println("...environment ready.")

	return nil
}

// bootstrapMagento2 runs a full Magento 2 bootstrap process.
func (c *bootstrapper) bootstrapMagento2() error {
	magentoVersion, err := c.MagentoVersion()
	if err != nil {
		return fmt.Errorf("cannot determine magento version: %w", err)
	}

	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Magento v%s?", magentoVersion.String())) {
		return nil
	}

	log.Printf("Bootstrapping Magento %s...", magentoVersion.String())

	err = c.prepare()
	if err != nil {
		return fmt.Errorf("error during bootstrap preparation: %w", err)
	}

	// Composer configuration
	err = c.composerPreInstall()
	if err != nil {
		return err
	}

	// Composer Install
	freshInstall, err := c.download()
	if err != nil {
		return fmt.Errorf("cannot download magento: %w", err)
	}

	err = c.composerInstall()
	if err != nil {
		return fmt.Errorf("cannot install composer dependencies: %w", err)
	}

	err = c.composerPostInstall()
	if err != nil {
		return err
	}

	adminPassword, err := c.installMagento2(magentoVersion, freshInstall)
	if err != nil {
		return err
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *Client) installMagento2(magentoVersion *version.Version, freshInstall bool) (string, error) {
	log.Println("Installing Magento...")

	var (
		minimumMagentoVersionForMFA   = version.Must(version.NewVersion("2.4.0"))
		minimumVersionForSearch       = version.Must(version.NewVersion("2.4.0"))
		minimumVersionForRabbitMQWait = version.Must(version.NewVersion("2.4.0"))
		magentoVerbosityFlag          = "-v"
	)

	if c.IsDebug() {
		magentoVerbosityFlag = "-vvv"
	}

	// Magento Install
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

		if magentoVersion.GreaterThan(minimumVersionForRabbitMQWait) {
			magentoCmdParams = append(magentoCmdParams, "--consumers-wait-for-messages=0")
		}
	}

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

	if c.ServiceEnabled("elasticsearch") ||
		c.ServiceEnabled("opensearch") &&
			magentoVersion.GreaterThan(minimumVersionForSearch) {
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

	log.Println("Running Magento setup:install...")
	// Magento install
	err := c.RunCmdEnvExec(fmt.Sprintf("bin/magento setup:install %s", strings.Join(magentoCmdParams, " ")))
	if err != nil {
		return "", err
	}

	log.Println("...Magento setup:install finished.")

	log.Println("Configuring Magento basic settings...")
	// Set base URL
	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return "", fmt.Errorf("unable to set base URL: %w", err)
	}

	// Set secure base URL
	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return "", fmt.Errorf("unable to set secure base URL: %w", err)
	}

	// Set offload header
	err = c.RunCmdEnvExec("bin/magento config:set --lock-env web/secure/offloader_header X-Forwarded-Proto")
	if err != nil {
		return "", fmt.Errorf("unable to set offload header: %w", err)
	}

	// Set use https in frontend
	err = c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_frontend 1")
	if err != nil {
		return "", fmt.Errorf("unable to set use https in frontend: %w", err)
	}

	// Set use https in admin
	err = c.RunCmdEnvExec("bin/magento config:set web/secure/use_in_adminhtml 1")
	if err != nil {
		return "", fmt.Errorf("unable to set use https in admin: %w", err)
	}

	// Set seo rewrites
	err = c.RunCmdEnvExec("bin/magento config:set web/seo/use_rewrites 1")
	if err != nil {
		return "", fmt.Errorf("unable to set seo rewrites: %w", err)
	}

	log.Println("...Magento basic settings configured.")

	// Configure varnish
	if c.ServiceEnabled("varnish") {
		log.Println("Configuring Varnish...")

		err = c.RunCmdEnvExec("bin/magento config:set --lock-env system/full_page_cache/caching_application 2")
		if err != nil {
			return "", fmt.Errorf("unable to configure varnish: %w", err)
		}

		err = c.RunCmdEnvExec("bin/magento config:set --lock-env system/full_page_cache/ttl 604800")
		if err != nil {
			return "", fmt.Errorf("unable to configure varnish cache ttl: %w", err)
		}

		log.Println("...Varnish configured.")
	}

	if c.ServiceEnabled("elasticsearch") || c.ServiceEnabled("opensearch") {
		log.Println("Configuring Elasticsearch/OpenSearch...")

		err = c.RunCmdEnvExec("bin/magento config:set --lock-env catalog/search/enable_eav_indexer 1")
		if err != nil {
			return "", fmt.Errorf("unable to enable eav indexer: %w", err)
		}

		if magentoVersion.GreaterThan(minimumVersionForSearch) {
			err = c.RunCmdEnvExec(
				fmt.Sprintf("bin/magento config:set --lock-env catalog/search/engine %s",
					searchEngine,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to set search engine: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_hostname %s",
					searchEngine,
					searchHost,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to set search engine server: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_port 9200",
					searchEngine,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to set search engine port: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_index_prefix magento2",
					searchEngine,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to set search engine index prefix: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_enable_auth 0",
					searchEngine,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to disable search engine auth: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"bin/magento config:set --lock-env catalog/search/%s_server_timeout 15",
					searchEngine,
				),
			)
			if err != nil {
				return "", fmt.Errorf("unable to set search engine timeout: %w", err)
			}
		}

		log.Println("...Elasticsearch configured.")
	}

	log.Println("Setting Magento deploy mode...")

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/magento deploy:mode:set -s %s",
			c.MagentoMode(),
		),
	)
	if err != nil {
		return "", fmt.Errorf("unable to set magento mode: %w", err)
	}

	log.Println("...Magento deploy:mode set.")

	// Disable TFA for local development.
	if magentoVersion.GreaterThan(minimumMagentoVersionForMFA) && c.MagentoDisableTFA() {
		log.Println("Disabling TFA for local development...")

		err = c.RunCmdEnvExec("bin/magento module:disable Magento_TwoFactorAuth")
		if err != nil {
			return "", fmt.Errorf("unable to disable TFA: %w", err)
		}

		log.Println("...TFA disabled.")
	}

	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", fmt.Errorf("unable to generate admin password: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			`bin/magento admin:user:create --admin-password=%s `+
				`--admin-user=localadmin --admin-firstname=Local --admin-lastname=Admin --admin-email="admin@example.com"`,
			adminPassword,
		),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create admin user: %w", err)
	}

	log.Println("...admin user created.")

	// sample data
	if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
		log.Println("Installing sample data...")

		err = c.RunCmdEnvExec("mkdir -p /var/www/html/var/composer_home/ && cp -va ~/.composer/auth.json /var/www/html/var/composer_home/auth.json")
		if err != nil {
			return "", fmt.Errorf("unable to copy auth.json: %w", err)
		}

		err = c.RunCmdEnvExec(
			fmt.Sprintf(
				`php bin/magento %s sampledata:deploy`,
				magentoVerbosityFlag,
			),
		)
		if err != nil {
			return "", fmt.Errorf("unable to deploy sample data: %w", err)
		}

		err = c.RunCmdEnvExec(
			fmt.Sprintf(
				`bin/magento setup:upgrade %s`,
				magentoVerbosityFlag,
			),
		)
		if err != nil {
			return "", fmt.Errorf("unable to run magento setup:upgrade: %w", err)
		}

		log.Println("...sample data installed successfully.")
	}

	if c.FullBootstrap() {
		log.Println("Reindexing...")

		err := c.RunCmdEnvExec("bin/magento indexer:reindex")
		if err != nil {
			return "", fmt.Errorf("unable to run magento indexer:reindex: %w", err)
		}

		log.Println("...reindexing complete.")
	}

	if c.ResetAdminURL() {
		log.Println("Resetting admin URL...")

		err = c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom 0")
		if err != nil {
			return "", fmt.Errorf("unable to reset admin url: %w", err)
		}

		err = c.RunCmdEnvExec("bin/magento config:set admin/url/use_custom_path 0")
		if err != nil {
			return "", fmt.Errorf("unable to reset admin url path: %w", err)
		}

		log.Println("...admin URL reset.")
	}

	log.Println("Flushing cache...")

	err = c.RunCmdEnvExec("bin/magento cache:flush")
	if err != nil {
		return "", fmt.Errorf("unable to flush cache: %w", err)
	}

	log.Println("...cache flushed.")
	log.Println("...Magento installed successfully.")

	return adminPassword, nil
}

// bootstrapMagento1 runs a full Magento 1 bootstrap process.
// Note: it will not install Magento 1 from zero, but only configures Magento 1's local.xml.
func (c *bootstrapper) bootstrapMagento1() error {
	magentoVersion, err := c.MagentoVersion()
	if err != nil {
		return fmt.Errorf("unable to get magento version: %w", err)
	}

	log.Printf("Bootstrapping Magento %s...", magentoVersion.String())

	if !util.AskForConfirmation("Would you like to bootstrap Magento v" + magentoVersion.String() + "?") {
		return nil
	}

	err = c.prepare()
	if err != nil {
		return fmt.Errorf("error during bootstrap preparation: %w", err)
	}

	// Composer Install
	if util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
		err = c.composerPreInstall()
		if err != nil {
			return err
		}

		err = c.composerInstall()
		if err != nil {
			return fmt.Errorf("cannot install composer dependencies: %w", err)
		}

		err = c.composerPostInstall()
		if err != nil {
			return err
		}
	}

	var (
		bs               bytes.Buffer
		localXMLFilePath = filepath.Join(c.Cwd(), c.WebRoot(), "app", "etc", "local.xml")
		localXMLTemplate = new(template.Template)
		tmpList          = new(list.List)
	)

	if util.CheckFileExistsAndRecreate(localXMLFilePath) {
		return fmt.Errorf("unable to create local.xml file")
	}

	err = templates.New().AppendTemplatesFromPathsStatic(localXMLTemplate, tmpList, []string{
		filepath.Join("templates", "magento1", "local.xml"),
	})
	if err != nil {
		return fmt.Errorf("unable to load local.xml template: %w", err)
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = templates.New().ExecuteTemplate(localXMLTemplate.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("unable to execute local.xml template: %w", err)
		}

		err = util.CreateDirAndWriteToFile(bs.Bytes(), localXMLFilePath)
		if err != nil {
			return fmt.Errorf("unable to write local.xml file: %w", err)
		}
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"/usr/bin/n98-magerun config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to set base url: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"/usr/bin/n98-magerun config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to set secure base url: %w", err)
	}

	err = c.RunCmdEnvExec("/usr/bin/n98-magerun config:set web/secure/use_in_frontend 1")
	if err != nil {
		return fmt.Errorf("unable to set use https in frontend: %w", err)
	}

	err = c.RunCmdEnvExec("/usr/bin/n98-magerun config:set web/secure/use_in_adminhtml 1")
	if err != nil {
		return fmt.Errorf("unable to set use https in adminhtml: %w", err)
	}

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return err
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"/usr/bin/n98-magerun admin:user:create localadmin admin@example.com %s Local Admin",
			adminPassword,
		),
	)
	if err != nil {
		return fmt.Errorf("unable to create admin user %w", err)
	}

	err = c.RunCmdEnvExec("/usr/bin/n98-magerun cache:flush")
	if err != nil {
		return fmt.Errorf("unable to flush cache: %w", err)
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

// bootstrapWordpress runs a full WordPress bootstrap process.
func (c *bootstrapper) bootstrapWordpress() error {
	if !util.AskForConfirmation("Would you like to bootstrap Wordpress?") {
		return nil
	}

	log.Println("Bootstrapping Wordpress...")

	err := c.prepare()
	if err != nil {
		return fmt.Errorf("error during bootstrap preparation: %w", err)
	}

	// Install
	_, err = c.download()
	if err != nil {
		return fmt.Errorf("cannot download wordpress: %w", err)
	}

	var (
		bs             bytes.Buffer
		configFilePath = filepath.Join(c.Cwd(), c.WebRoot(), "wp-config.php")
		tpl            = new(template.Template)
		tmpList        = new(list.List)
		tplPath        = []string{
			filepath.Join("templates", "wordpress", "wp-config.php"),
		}
	)

	if util.CheckFileExistsAndRecreate(configFilePath) {
		return nil
	}

	err = templates.New().AppendTemplatesFromPathsStatic(tpl, tmpList, tplPath)
	if err != nil {
		return fmt.Errorf("unable to load wp-config.php template: %w", err)
	}

	if c.DBPrefix() != "" {
		c.Set("wordpress_table_prefix", c.DBPrefix())
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = templates.New().ExecuteTemplate(tpl.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("unable to execute wp-config.php template: %w", err)
		}

		err = util.CreateDirAndWriteToFile(bs.Bytes(), configFilePath)
		if err != nil {
			return fmt.Errorf("unable to write wp-config.php file: %w", err)
		}
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) bootstrapShopware() error {
	shopwareVersion, err := c.ShopwareVersion()
	if err != nil {
		return fmt.Errorf("cannot determine shopware version: %w", err)
	}

	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Shopware v%s?",
		shopwareVersion.String())) {
		return nil
	}

	log.Printf("Bootstrapping Shopware %s...", shopwareVersion.String())

	err = c.prepare()
	if err != nil {
		return fmt.Errorf("error during bootstrap preparation: %w", err)
	}

	var freshInstall bool

	// Composer configuration
	err = c.composerPreInstall()
	if err != nil {
		return err
	}

	// Composer Install
	freshInstall, err = c.download()
	if err != nil {
		return err
	}

	err = c.composerInstall()
	if err != nil {
		return fmt.Errorf("cannot install composer dependencies: %w", err)
	}

	err = c.composerPostInstall()
	if err != nil {
		return err
	}

	adminPassword, err := c.installShopware(freshInstall)
	if err != nil {
		return err
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.MagentoBackendFrontname())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) download() (bool, error) {
	if c.SkipComposerInstall() {
		return false, nil
	}

	var (
		freshInstall          = false
		composerVerbosityFlag = "--verbose"
		rsyncVerbosityFlag    = "-v"
	)

	if c.IsDebug() {
		composerVerbosityFlag = "-vvv"
		rsyncVerbosityFlag = "-v"
	}

	switch c.EnvType() {
	case "magento2":
		if !util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
			log.Println("Creating Magento 2 composer project...")

			magentoVersion, err := c.MagentoVersion()
			if err != nil {
				return false, fmt.Errorf("cannot determine magento version: %w", err)
			}

			freshInstall = true

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"composer create-project %s --profile --no-install "+
						"--repository-url=https://repo.magento.com/ "+
						"magento/project-%s-edition=%s /tmp/magento-tmp/",
					composerVerbosityFlag,
					c.MagentoType(),
					magentoVersion.String(),
				),
			)
			if err != nil {
				return false, fmt.Errorf("cannot create composer magento project: %w", err)
			}

			err = c.RunCmdEnvExec(
				fmt.Sprintf(
					`rsync %s -au --remove-source-files --chmod=D2775,F644 /tmp/magento-tmp/ /var/www/html/`,
					rsyncVerbosityFlag,
				),
			)
			if err != nil {
				return false, fmt.Errorf("cannot move magento project install files: %w", err)
			}

			log.Println("...Magento 2 composer project created.")
		}

	case "wordpress":
		if !util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "index.php")) {
			log.Println("Downloading and installing WordPress...")

			freshInstall = true

			err := c.RunCmdEnvExec("wget -qO /tmp/wordpress.tar.gz https://wordpress.org/latest.tar.gz")
			if err != nil {
				return false, fmt.Errorf("unable to download wordpress: %w", err)
			}

			err = c.RunCmdEnvExec("tar -zxf /tmp/wordpress.tar.gz --strip-components=1 -C /var/www/html")
			if err != nil {
				return false, fmt.Errorf("unable to extract wordpress: %w", err)
			}

			err = c.RunCmdEnvExec("rm -f /tmp/wordpress.tar.gz")
			if err != nil {
				return false, fmt.Errorf("unable to remove wordpress archive: %w", err)
			}

			log.Println("...WordPress downloaded.")
		}

	case "shopware":
		if !util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
			log.Println("Downloading and installing Shopware...")

			freshInstall = true

			path := "production"
			if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
				path = "development"
			}

			err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"wget -qO /tmp/shopware.tar.gz https://github.com/shopware/%s/archive/refs/tags/v%s.tar.gz",
					path,
					version.Must(c.ShopwareVersion()).String(),
				),
			)
			if err != nil {
				return false, fmt.Errorf("unable to download shopware: %w", err)
			}

			err = c.RunCmdEnvExec("tar -zxf /tmp/shopware.tar.gz --strip-components=1 -C /var/www/html")
			if err != nil {
				return false, fmt.Errorf("unable to extract shopware: %w", err)
			}

			err = c.RunCmdEnvExec("rm -f /tmp/shopware.tar.gz")
			if err != nil {
				return false, fmt.Errorf("unable to remove shopware archive: %w", err)
			}

			log.Println("...Shopware downloaded.")
		}
	}

	return freshInstall, nil
}

func (c *bootstrapper) installShopware(freshInstall bool) (string, error) {
	// Elasticsearch/OpenSearch configuration
	searchEnabled, searchHost := 0, ""
	{
		switch {
		case c.ServiceEnabled("opensearch"):
			searchHost = "opensearch"
			c.Set("SHOPWARE_SEARCH_ENABLED", 1)
			c.Set("SHOPWARE_SEARCH_INDEXING_ENABLED", 1)
			c.Set("SHOPWARE_SEARCH_HOST", "opensearch")

		case c.ServiceEnabled("elasticsearch"):
			searchHost = "elasticsearch"
			c.Set("SHOPWARE_SEARCH_ENABLED", 1)
			c.Set("SHOPWARE_SEARCH_INDEXING_ENABLED", 1)
			c.Set("SHOPWARE_SEARCH_HOST", "elasticsearch")
		}
	}

	if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
		err := func() error {
			var (
				bs             bytes.Buffer
				configFilePath = filepath.Join(c.Cwd(), c.WebRoot(), ".psh.yaml.override")
				tpl            = new(template.Template)
				tmpList        = new(list.List)
				tplPath        = []string{
					filepath.Join("templates", "shopware", ".psh.yaml.override"),
				}
			)

			if util.CheckFileExistsAndRecreate(configFilePath) {
				return nil
			}

			err := templates.New().AppendTemplatesFromPathsStatic(tpl, tmpList, tplPath)
			if err != nil {
				return fmt.Errorf("cannot create .psh.yaml.override template: %w", err)
			}

			for e := tmpList.Front(); e != nil; e = e.Next() {
				tplName := fmt.Sprint(e.Value)

				err = templates.New().ExecuteTemplate(tpl.Lookup(tplName), &bs)
				if err != nil {
					return fmt.Errorf("cannot execute .psh.yaml.override template: %w", err)
				}

				err = util.CreateDirAndWriteToFile(bs.Bytes(), configFilePath)
				if err != nil {
					return fmt.Errorf("cannot write .psh.yaml.override file %s: %w",
						configFilePath,
						err)
				}
			}

			return nil
		}()
		if err != nil {
			return "", err
		}

		err = c.RunCmdEnvExec("chmod +x psh.phar bin/console bin/setup")
		if err != nil {
			return "", fmt.Errorf("cannot set permissions: %w", err)
		}

		err = c.RunCmdEnvExec("export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && ./psh.phar install")
		if err != nil {
			return "", fmt.Errorf("cannot install shopware: %w", err)
		}
	} else {
		err := c.RunCmdEnvExec(
			fmt.Sprintf(
				"bin/console system:setup --no-interaction --force --app-env dev --app-url https://%s --database-url mysql://app:app@db:3306/shopware --es-enabled=%d --es-hosts=%s:9200 --es-indexing-enabled=%d --cdn-strategy=physical_filename --mailer-url=native://default",
				c.TraefikFullDomain(),
				searchEnabled,
				searchHost,
				searchEnabled,
			),
		)
		if err != nil {
			return "", fmt.Errorf("cannot setup shopware: %w", err)
		}

		params := ""
		if freshInstall {
			params = "--basic-setup"
		}

		err = c.RunCmdEnvExec(
			fmt.Sprintf(
				"bin/console system:install --no-interaction --force %s", params,
			),
		)
		if err != nil {
			return "", fmt.Errorf("cannot install shopware: %w", err)
		}

		err = c.RunCmdEnvExec("export CI=1 && bin/console bundle:dump")
		if err != nil {
			return "", fmt.Errorf("cannot dump shopware bundles: %w", err)
		}

		// Ignore if themes cannot be dumped.
		_ = c.RunCmdEnvExec("export CI=1 && bin/console theme:dump")

		err = c.RunCmdEnvExec("export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && bin/build.sh")
		if err != nil {
			return "", fmt.Errorf("cannot build storefront: %w", err)
		}

		err = c.RunCmdEnvExec("bin/console system:update:finish --no-interaction")
		if err != nil {
			return "", fmt.Errorf("cannot finish shopware setup: %w", err)
		}

		// The dev bootstrap script already installs the demo data
		if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
			err := c.RunCmdEnvExec("mkdir -p custom/plugins && php bin/console store:download -p SwagPlatformDemoData && php bin/console plugin:install SwagPlatformDemoData --activate")
			if err != nil {
				return "", fmt.Errorf("cannot install demo data: %w", err)
			}
		}
	}

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", fmt.Errorf("unable to generate admin password: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			`php bin/console user:create --admin --email="admin@example.com" --firstName="Admin" --lastName="Local" --password="%s" --no-interaction localadmin`,
			adminPassword,
		),
	)
	if err != nil {
		return "", fmt.Errorf("cannot create admin user: %w", err)
	}

	err = c.RunCmdEnvExec("php bin/console cache:clear")
	if err != nil {
		return "", fmt.Errorf("cannot clear cache: %w", err)
	}

	return adminPassword, nil
}

func (c *bootstrapper) composerInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	log.Println("Installing composer dependencies...")

	err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"composer install %s --profile",
			c.composerVerbosityFlag,
		),
	)
	if err != nil {
		return fmt.Errorf("cannot install composer dependencies: %w", err)
	}

	log.Println("...composer dependencies installed.")

	return nil
}

func (c *bootstrapper) composerPreInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	log.Println("Configuring composer...")

	composerVersion := 2
	if c.ComposerVersion().LessThan(version.Must(version.NewVersion("2.0.0"))) {
		composerVersion = 1
	}

	if composerVersion == 1 {
		log.Println("Setting default composer version to 1.x")

		// Change default Composer Version
		err := c.RunCmdEnvExec("sudo alternatives --set composer /usr/bin/composer1")
		if err != nil {
			return fmt.Errorf("cannot change default composer version: %w", err)
		}
	} else {
		log.Println("Setting default composer version to 2.x")

		// Change default Composer Version
		err := c.RunCmdEnvExec("sudo alternatives --set composer /usr/bin/composer2")
		if err != nil {
			return fmt.Errorf("cannot change default composer version: %w", err)
		}

		// Specific Composer Version
		if !c.ComposerVersion().Equal(version.Must(version.NewVersion("2.0.0"))) {
			err = c.RunCmdEnvExec("sudo composer self-update " + c.ComposerVersion().String())
			if err != nil {
				return fmt.Errorf("cannot change default composer version: %w", err)
			}
		}
	}

	// Composer Install
	if c.Parallel() && composerVersion < 2 {
		err := c.RunCmdEnvExec(
			fmt.Sprintf(
				"composer global require %s --profile hirak/prestissimo",
				c.composerVerbosityFlag,
			),
		)
		if err != nil {
			return fmt.Errorf("cannot install hirak/prestissimo composer module: %w", err)
		}
	}

	log.Println("...composer configured.")

	return nil
}

func (c *bootstrapper) composerPostInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	composerVersion := 2
	if c.ComposerVersion().LessThan(version.Must(version.NewVersion("2.0.0"))) {
		composerVersion = 1
	}

	if !c.SkipComposerInstall() {
		if c.Parallel() && composerVersion != 2 {
			log.Println("Removing hirak/prestissimo composer module...")

			err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"composer global remove %s --profile hirak/prestissimo",
					c.composerVerbosityFlag,
				),
			)
			if err != nil {
				return fmt.Errorf("cannot remove hirak/prestissimo module: %w", err)
			}

			log.Println("...hirak/prestissimo composer module removed.")
		}
	}

	return nil
}

func (c *Client) RunCmdEnvExec(args string) error {
	return c.RunCmdEnv(append([]string{"exec", "-T", c.DefaultSyncedContainer(c.EnvType()), "bash", "-c"}, args))
}

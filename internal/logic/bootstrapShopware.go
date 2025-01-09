package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"net/url"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

func (c *shopware) bootstrap() error {
	shopwareVersion, err := c.ShopwareVersion()
	if err != nil {
		return errors.Wrap(err, "determining shopware version")
	}

	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Shopware v%s?",
		shopwareVersion.String(),
	),
	) {
		return nil
	}

	log.Printf("Bootstrapping Shopware %s...", shopwareVersion.String())

	if err := c.prepare(); err != nil {
		return errors.Wrap(err, "preparing bootstrap")
	}

	var freshInstall bool

	// Composer configuration
	if err = c.composerPreInstall(); err != nil {
		return err
	}

	// Composer Install
	freshInstall, err = c.download()
	if err != nil {
		return err
	}

	if err := c.composerInstall(); err != nil {
		return err
	}

	if err := c.composerPostInstall(); err != nil {
		return err
	}

	adminPassword, err := c.install(freshInstall)
	if err != nil {
		return err
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.ShopwareAdminPath())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *shopware) download() (downloaded bool, err error) {
	if c.SkipComposerInstall() {
		return false, nil
	}

	if util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
		return false, nil
	}

	log.Println("Downloading and installing Shopware...")

	path := "production"
	if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
		path = "development"
	}

	command := fmt.Sprintf("wget -qO /tmp/shopware.tar.gz https://github.com/shopware/%s/archive/refs/tags/v%s.tar.gz",
		path,
		version.Must(c.ShopwareVersion()).String(),
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return false, errors.Wrap(err, "downloading shopware")
	}

	command = "tar -zxf /tmp/shopware.tar.gz --strip-components=1 -C /var/www/html"
	if err := c.RunCmdEnvExec(command); err != nil {
		return false, errors.Wrap(err, "extracting shopware")
	}

	command = "rm -f /tmp/shopware.tar.gz"
	if err := c.RunCmdEnvExec(command); err != nil {
		return false, errors.Wrap(err, "removing shopware archive")
	}

	log.Println("...Shopware downloaded.")

	return true, nil
}

func (c *shopware) install(freshInstall bool) (string, error) {
	if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
		if err := c.devConfig(); err != nil {
			return "", err
		}

		if err := c.devSetup(); err != nil {
			return "", err
		}
	} else {
		if err := c.prodSetup(freshInstall); err != nil {
			return "", err
		}

		if err := c.deployDemoData(freshInstall); err != nil {
			return "", err
		}
	}

	adminPassword, err := c.configureAdminUser()
	if err != nil {
		return "", err
	}

	if err := c.clearCache(); err != nil {
		return "", err
	}

	return adminPassword, nil
}

func (c *shopware) deployDemoData(freshInstall bool) error {
	if !(freshInstall && (c.WithSampleData() || c.FullBootstrap())) {
		return nil
	}

	log.Println("Deploying Shopware demo data...")

	command := fmt.Sprintf("mkdir -p custom/plugins && "+
		"%s store:download -p SwagPlatformDemoData && "+
		"%s plugin:install SwagPlatformDemoData --activate",
		c.consoleCommand(), c.consoleCommand(),
	)

	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "installing demo data")
	}

	log.Println("...demo data deployed.")

	return nil
}

func (c *shopware) prodSetup(freshInstall bool) error {
	log.Println("Setting up Shopware production template...")

	searchEnabled, searchHost := c.configureSearch()

	command := fmt.Sprintf(
		"%s system:setup "+
			"--no-interaction --force "+
			"--app-env dev --app-url https://%s "+
			"--database-url mysql://app:app@db:3306/shopware "+
			"--es-enabled=%d --es-hosts=%s --es-indexing-enabled=%d "+
			"--cdn-strategy=physical_filename "+
			"--mailer-url=native://default",
		c.consoleCommand(),
		c.TraefikFullDomain(),
		searchEnabled,
		searchHost,
		searchEnabled,
	)

	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware system:setup")
	}

	// Add LOCK_DSN to .env
	command = `echo 'LOCK_DSN="flock://var/lock"' >> .env`
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "adding LOCK_DSN to .env")
	}

	params := ""
	if freshInstall {
		params = "--basic-setup"
	}

	command = fmt.Sprintf("%s system:install --no-interaction --force %s", c.consoleCommand(), params)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware system:install")
	}

	command = fmt.Sprintf("export CI=1 && %s bundle:dump", c.consoleCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware bundle:dump")
	}

	// Ignore if themes cannot be dumped.
	command = fmt.Sprintf("export CI=1 && %s theme:dump", c.consoleCommand())
	_ = c.RunCmdEnvExec(command)

	command = "export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && if [ -f 'bin/build.sh' ]; then bin/build.sh; fi"
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware bin/build.sh")
	}

	command = "export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && if [ -f 'bin/build-js.sh']; then bin/build-js.sh; fi"
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware bin/build-js.sh")
	}

	command = fmt.Sprintf("%s system:update:finish --no-interaction", c.consoleCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware system:update:finish")
	}

	log.Println("...Shopware production template installed.")

	return nil
}

func (c *shopware) devSetup() error {
	log.Println("Setting up Shopware development template...")

	command := fmt.Sprintf("chmod +x psh.phar %s bin/setup", c.consoleCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting permissions")
	}

	command = "export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && ./psh.phar install"
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "running shopware ./psh.phar install")
	}

	log.Println("...Shopware development template setup finished.")

	return nil
}

func (c *shopware) configureSearch() (int, string) {
	// Elasticsearch/OpenSearch configuration
	searchEnabled, searchHost := 0, ""

	switch {
	case c.ServiceEnabled("opensearch"):
		searchEnabled = 1
		searchHost = "http://opensearch:9200"

		c.Set("SHOPWARE_SEARCH_ENABLED", 1)
		c.Set("SHOPWARE_SEARCH_INDEXING_ENABLED", 1)
		c.Set("SHOPWARE_SEARCH_HOST", "opensearch")

		openSearchInitialAdminPassword := c.GetString("OPENSEARCH_INITIAL_ADMIN_PASSWORD")
		if openSearchInitialAdminPassword != "" {
			c.Set("SHOPWARE_SEARCH_USERNAME", "admin")
			c.Set("SHOPWARE_SEARCH_PASSWORD", openSearchInitialAdminPassword)
			searchHost = "http://admin:" + url.PathEscape(openSearchInitialAdminPassword) + "@opensearch:9200"
		}

	case c.ServiceEnabled("elasticsearch"):
		searchEnabled = 1
		searchHost = "elasticsearch"

		c.Set("SHOPWARE_SEARCH_ENABLED", 1)
		c.Set("SHOPWARE_SEARCH_INDEXING_ENABLED", 1)
		c.Set("SHOPWARE_SEARCH_HOST", "elasticsearch")
	}

	return searchEnabled, searchHost
}

func (c *shopware) configureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword := c.generatePassword()

	command := fmt.Sprintf(
		`%s user:create --no-interaction `+
			`--admin `+
			`--email="admin@example.com" `+
			`--firstName="Admin" `+
			`--lastName="Local" `+
			`--password="%s" `+
			`localadmin`,
		c.consoleCommand(),
		adminPassword,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return "", errors.Wrap(err, "creating admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *shopware) clearCache() error {
	log.Println("Clearing cache...")

	command := fmt.Sprintf("%s cache:clear", c.consoleCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "clearing cache")
	}

	log.Println("...cache cleared.")

	return nil
}

func (c *shopware) devConfig() error {
	_, _ = c.configureSearch()

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

	if err := templates.New().AppendTemplatesFromPathsStatic(tpl, tmpList, tplPath); err != nil {
		return errors.Wrap(err, "creating .psh.yaml.override template")
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		if err := templates.New().ExecuteTemplate(tpl.Lookup(tplName), &bs); err != nil {
			return errors.Wrap(err, "executing .psh.yaml.override template")
		}

		if err := util.CreateDirAndWriteToFile(bs.Bytes(), configFilePath); err != nil {
			return errors.Wrapf(err, "writing .psh.yaml.override file %s", configFilePath)
		}
	}

	return nil
}

func (c *shopware) consoleCommand() string {
	verbosity := "-v"

	if c.IsDebug() {
		verbosity += "vv"
	}

	return fmt.Sprintf("php bin/console %s", verbosity)
}

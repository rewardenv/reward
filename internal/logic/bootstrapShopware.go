package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"net/url"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

func (c *bootstrapper) bootstrapShopware() error {
	shopwareVersion, err := c.ShopwareVersion()
	if err != nil {
		return errors.Wrap(err, "determining shopware version")
	}

	if !util.AskForConfirmation(fmt.Sprintf("Would you like to bootstrap Shopware v%s?",
		shopwareVersion.String())) {
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

	adminPassword, err := c.installShopware(freshInstall)
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

func (c *bootstrapper) installShopware(freshInstall bool) (string, error) {
	if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
		if err := c.installShopwareDevConfig(); err != nil {
			return "", err
		}

		if err := c.installShopwareDevSetup(); err != nil {
			return "", err
		}
	} else {
		if err := c.installShopwareProdSetup(freshInstall); err != nil {
			return "", err
		}

		if err := c.installShopwareDemoData(freshInstall); err != nil {
			return "", err
		}
	}

	adminPassword, err := c.installShopwareConfigureAdminUser()
	if err != nil {
		return "", err
	}

	if err := c.installShopwareClearCache(); err != nil {
		return "", err
	}

	return adminPassword, nil
}

func (c *bootstrapper) installShopwareDemoData(freshInstall bool) error {
	log.Println("Deploying Shopware demo data...")

	if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
		if err := c.RunCmdEnvExec("mkdir -p custom/plugins && " +
			"php bin/console store:download -p SwagPlatformDemoData && " +
			"php bin/console plugin:install SwagPlatformDemoData --activate",
		); err != nil {
			return errors.Wrap(err, "installing demo data")
		}
	}

	log.Println("...demo data deployed.")

	return nil
}

func (c *bootstrapper) installShopwareProdSetup(freshInstall bool) error {
	log.Println("Setting up Shopware production template...")

	searchEnabled, searchHost := c.installShopwareConfigureSearch()

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/console system:setup "+
				"--no-interaction --force "+
				"--app-env dev --app-url https://%s "+
				"--database-url mysql://app:app@db:3306/shopware "+
				"--es-enabled=%d --es-hosts=%s --es-indexing-enabled=%d "+
				"--cdn-strategy=physical_filename "+
				"--mailer-url=native://default",
			c.TraefikFullDomain(),
			searchEnabled,
			searchHost,
			searchEnabled,
		),
	); err != nil {
		return errors.Wrap(err, "running shopware system:setup")
	}

	// Add LOCK_DSN to .env
	if err := c.RunCmdEnvExec(
		`echo 'LOCK_DSN="flock://var/lock"' >> .env`,
	); err != nil {
		return errors.Wrap(err, "adding LOCK_DSN to .env")
	}

	params := ""
	if freshInstall {
		params = "--basic-setup"
	}

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/console system:install --no-interaction --force %s", params,
		),
	); err != nil {
		return errors.Wrap(err, "running shopware system:install")
	}

	if err := c.RunCmdEnvExec("export CI=1 && bin/console bundle:dump"); err != nil {
		return errors.Wrap(err, "running shopware bundle:dump")
	}

	// Ignore if themes cannot be dumped.
	_ = c.RunCmdEnvExec("export CI=1 && bin/console theme:dump")

	if err := c.RunCmdEnvExec(
		"export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && if [ -f 'bin/build.sh' ]; then bin/build.sh; fi",
	); err != nil {
		return errors.Wrap(err, "running shopware bin/build.sh")
	}

	if err := c.RunCmdEnvExec(
		"export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && if [ -f 'bin/build-storefront.sh']; then bin/build-storefront.sh; fi",
	); err != nil {
		return errors.Wrap(err, "running shopware bin/build-storefront.sh")
	}

	//nolint:lll
	if err := c.RunCmdEnvExec(
		"export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && if [ -f 'bin/build-administration.sh']; then bin/build-administration.sh; fi",
	); err != nil {
		return errors.Wrap(err, "running shopware bin/build-administration.sh")
	}

	if err := c.RunCmdEnvExec("bin/console system:update:finish --no-interaction"); err != nil {
		return errors.Wrap(err, "running shopware system:update:finish")
	}

	log.Println("...Shopware production template installed.")

	return nil
}

func (c *bootstrapper) installShopwareDevSetup() error {
	log.Println("Setting up Shopware development template...")

	if err := c.RunCmdEnvExec("chmod +x psh.phar bin/console bin/setup"); err != nil {
		return errors.Wrap(err, "setting permissions")
	}

	if err := c.RunCmdEnvExec("export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && ./psh.phar install"); err != nil {
		return errors.Wrap(err, "running shopware ./psh.phar install")
	}

	log.Println("...Shopware development template setup finished.")

	return nil
}

func (c *bootstrapper) installShopwareConfigureSearch() (int, string) {
	// Elasticsearch/OpenSearch configuration
	searchEnabled, searchHost := 0, ""
	{
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
	}

	return searchEnabled, searchHost
}

func (c *bootstrapper) installShopwareConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", errors.Wrap(err, "generating admin password")
	}

	if err = c.RunCmdEnvExec(
		fmt.Sprintf(
			`php bin/console user:create --no-interaction `+
				`--admin `+
				`--email="admin@example.com" `+
				`--firstName="Admin" `+
				`--lastName="Local" `+
				`--password="%s" `+
				`localadmin`,
			adminPassword,
		),
	); err != nil {
		return "", errors.Wrap(err, "creating admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installShopwareClearCache() error {
	log.Println("Clearing cache...")

	if err := c.RunCmdEnvExec("php bin/console cache:clear"); err != nil {
		return errors.Wrap(err, "clearing cache")
	}

	log.Println("...cache cleared.")

	return nil
}

func (c *bootstrapper) installShopwareDevConfig() error {
	_, _ = c.installShopwareConfigureSearch()

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

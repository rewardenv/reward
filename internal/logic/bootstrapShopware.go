package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

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
		return err
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
	log.Printf("Backend Url: https://%s/%s", c.TraefikFullDomain(), c.ShopwareAdminPath())
	log.Println("Admin user: localadmin")
	log.Printf("Admin password: %s", adminPassword)
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) installShopware(freshInstall bool) (string, error) {
	if c.ShopwareMode() == "dev" || c.ShopwareMode() == "development" {
		err := c.installShopwareDevConfig()
		if err != nil {
			return "", err
		}

		err = c.installShopwareDevSetup()
		if err != nil {
			return "", err
		}
	} else {
		err := c.installShopwareProdSetup(freshInstall)
		if err != nil {
			return "", err
		}

		err = c.installShopwareDemoData(freshInstall)
		if err != nil {
			return "", err
		}
	}

	adminPassword, err := c.installShopwareConfigureAdminUser()
	if err != nil {
		return "", err
	}

	err = c.installShopwareClearCache()
	if err != nil {
		return "", err
	}

	return adminPassword, nil
}

func (c *bootstrapper) installShopwareDemoData(freshInstall bool) error {
	log.Println("Deploying Shopware demo data...")

	if freshInstall && (c.WithSampleData() || c.FullBootstrap()) {
		err := c.RunCmdEnvExec("mkdir -p custom/plugins && " +
			"php bin/console store:download -p SwagPlatformDemoData && " +
			"php bin/console plugin:install SwagPlatformDemoData --activate",
		)
		if err != nil {
			return fmt.Errorf("cannot install demo data: %w", err)
		}
	}

	log.Println("...demo data deployed.")

	return nil
}

func (c *bootstrapper) installShopwareProdSetup(freshInstall bool) error {
	log.Println("Setting up Shopware production template...")

	searchEnabled, searchHost := c.installShopwareConfigureSearch()

	err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"bin/console system:setup "+
				"--no-interaction --force "+
				"--app-env dev --app-url https://%s "+
				"--database-url mysql://app:app@db:3306/shopware "+
				"--es-enabled=%d --es-hosts=%s:9200 --es-indexing-enabled=%d "+
				"--cdn-strategy=physical_filename "+
				"--mailer-url=native://default",
			c.TraefikFullDomain(),
			searchEnabled,
			searchHost,
			searchEnabled,
		),
	)
	if err != nil {
		return fmt.Errorf("cannot run shopware system:setup: %w", err)
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
		return fmt.Errorf("cannot run shopware system:install: %w", err)
	}

	err = c.RunCmdEnvExec("export CI=1 && bin/console bundle:dump")
	if err != nil {
		return fmt.Errorf("cannot run shopware bundle:dump: %w", err)
	}

	// Ignore if themes cannot be dumped.
	_ = c.RunCmdEnvExec("export CI=1 && bin/console theme:dump")

	err = c.RunCmdEnvExec("export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && bin/build.sh")
	if err != nil {
		return fmt.Errorf("cannot build storefront: %w", err)
	}

	err = c.RunCmdEnvExec("bin/console system:update:finish --no-interaction")
	if err != nil {
		return fmt.Errorf("cannot run shopware system:update:finish: %w", err)
	}

	log.Println("...Shopware production template installed.")

	return nil
}

func (c *bootstrapper) installShopwareDevSetup() error {
	log.Println("Setting up Shopware development template...")

	err := c.RunCmdEnvExec("chmod +x psh.phar bin/console bin/setup")
	if err != nil {
		return fmt.Errorf("cannot set permissions: %w", err)
	}

	err = c.RunCmdEnvExec("export PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=1 && ./psh.phar install")
	if err != nil {
		return fmt.Errorf("cannot run shopware ./psh.phar install: %w", err)
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

	return searchEnabled, searchHost
}

func (c *bootstrapper) installShopwareConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", fmt.Errorf("cannot generate admin password: %w", err)
	}

	err = c.RunCmdEnvExec(
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
	)
	if err != nil {
		return "", fmt.Errorf("cannot create admin user: %w", err)
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installShopwareClearCache() error {
	log.Println("Clearing cache...")

	err := c.RunCmdEnvExec("php bin/console cache:clear")
	if err != nil {
		return fmt.Errorf("cannot clear cache: %w", err)
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
}

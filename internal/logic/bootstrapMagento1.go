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

// bootstrapMagento1 runs a full Magento 1 bootstrap process.
// Note: it will not install Magento 1 from zero, but only configures Magento 1's local.xml.
func (c *bootstrapper) bootstrapMagento1() error {
	magentoVersion, err := c.MagentoVersion()
	if err != nil {
		return fmt.Errorf("cannot to get magento version: %w", err)
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
			return fmt.Errorf("error during composer configuration: %w", err)
		}

		err = c.composerInstall()
		if err != nil {
			return fmt.Errorf("error during composer install: %w", err)
		}

		err = c.composerPostInstall()
		if err != nil {
			return fmt.Errorf("error during composer post install configuration: %w", err)
		}
	}

	err = c.installMagento1GenerateLocalXML()
	if err != nil {
		return err
	}

	err = c.installMagento1ConfigureBasic()
	if err != nil {
		return err
	}

	adminPassword, err := c.installMagento1ConfigureAdminUser()
	if err != nil {
		return err
	}

	err = c.installMagento1FlushCache()
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

func (c *bootstrapper) installMagento1GenerateLocalXML() error {
	log.Println("Generating local.xml...")

	var (
		bs               bytes.Buffer
		localXMLFilePath = filepath.Join(c.Cwd(), c.WebRoot(), "app", "etc", "local.xml")
		localXMLTemplate = new(template.Template)
		tmpList          = new(list.List)
	)

	if util.CheckFileExistsAndRecreate(localXMLFilePath) {
		return fmt.Errorf("cannot create magento local.xml file")
	}

	err := templates.New().AppendTemplatesFromPathsStatic(localXMLTemplate, tmpList, []string{
		filepath.Join("templates", "magento1", "local.xml"),
	})
	if err != nil {
		return fmt.Errorf("cannot load magento local.xml template: %w", err)
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = templates.New().ExecuteTemplate(localXMLTemplate.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("cannot execute magento local.xml template: %w", err)
		}

		err = util.CreateDirAndWriteToFile(bs.Bytes(), localXMLFilePath)
		if err != nil {
			return fmt.Errorf("cannot write magento local.xml file: %w", err)
		}
	}

	log.Println("...local.xml generated.")

	return nil
}

func (c *bootstrapper) installMagento1ConfigureBasic() error {
	log.Println("Configuring Magento basic settings...")

	err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("cannot set magento base url: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	)
	if err != nil {
		return fmt.Errorf("cannot set magento secure base url: %w", err)
	}

	err = c.RunCmdEnvExec("n98-magerun config:set web/secure/use_in_frontend 1")
	if err != nil {
		return fmt.Errorf("cannot set magento to use https in frontend: %w", err)
	}

	err = c.RunCmdEnvExec("n98-magerun config:set web/secure/use_in_adminhtml 1")
	if err != nil {
		return fmt.Errorf("cannot set magento to use https in adminhtml: %w", err)
	}

	log.Println("...Magento basic settings configured.")

	return nil
}

func (c *bootstrapper) installMagento1ConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", fmt.Errorf("cannot generate magento admin password: %w", err)
	}

	err = c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun admin:user:create localadmin admin@example.com %s Local Admin",
			adminPassword,
		),
	)
	if err != nil {
		return "", fmt.Errorf("cannot create magento admin user: %w", err)
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento1FlushCache() error {
	log.Println("Flushing cache...")

	err := c.RunCmdEnvExec("n98-magerun cache:flush")
	if err != nil {
		return fmt.Errorf("cannot run flush magento cache: %w", err)
	}

	log.Println("...cache flushed.")

	return nil
}

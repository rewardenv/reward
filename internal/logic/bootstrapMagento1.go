package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "getting magento version")
	}

	log.Printf("Bootstrapping Magento %s...", magentoVersion.String())

	if !util.AskForConfirmation("Would you like to bootstrap Magento v" + magentoVersion.String() + "?") {
		return nil
	}

	if err := c.prepare(); err != nil {
		return errors.Wrap(err, "error during bootstrap preparation")
	}

	// Composer Install
	if util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
		if err := c.composerPreInstall(); err != nil {
			return errors.Wrap(err, "configuring composer")
		}

		if err := c.composerInstall(); err != nil {
			return errors.Wrap(err, "running composer install")
		}

		if err := c.composerPostInstall(); err != nil {
			return errors.Wrap(err, "running composer post install configuration")
		}
	}

	if err := c.installMagento1GenerateLocalXML(); err != nil {
		return err
	}

	if err := c.installMagento1ConfigureBasic(); err != nil {
		return err
	}

	adminPassword, err := c.installMagento1ConfigureAdminUser()
	if err != nil {
		return err
	}

	if err := c.installMagento1FlushCache(); err != nil {
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
		return errors.New("cannot create magento local.xml file")
	}

	if err := templates.New().AppendTemplatesFromPathsStatic(localXMLTemplate, tmpList, []string{
		filepath.Join("templates", "magento1", "local.xml"),
	}); err != nil {
		return errors.Wrap(err, "loading magento local.xml template")
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		if err := templates.New().ExecuteTemplate(localXMLTemplate.Lookup(tplName), &bs); err != nil {
			return errors.Wrap(err, "executing magento local.xml template")
		}

		if err := util.CreateDirAndWriteToFile(bs.Bytes(), localXMLFilePath); err != nil {
			return errors.Wrap(err, "writing magento local.xml file")
		}
	}

	log.Println("...local.xml generated.")

	return nil
}

func (c *bootstrapper) installMagento1ConfigureBasic() error {
	log.Println("Configuring Magento basic settings...")

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun config:set web/unsecure/base_url http://%s/",
			c.TraefikFullDomain(),
		),
	); err != nil {
		return errors.Wrap(err, "setting magento base url")
	}

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun config:set web/secure/base_url https://%s/",
			c.TraefikFullDomain(),
		),
	); err != nil {
		return errors.Wrap(err, "setting magento secure base url")
	}

	if err := c.RunCmdEnvExec("n98-magerun config:set web/secure/use_in_frontend 1"); err != nil {
		return errors.Wrap(err, "setting magento to use https on frontend")
	}

	if err := c.RunCmdEnvExec("n98-magerun config:set web/secure/use_in_adminhtml 1"); err != nil {
		return errors.Wrap(err, "setting magento to use https on admin")
	}

	log.Println("...Magento basic settings configured.")

	return nil
}

func (c *bootstrapper) installMagento1ConfigureAdminUser() (string, error) {
	log.Println("Creating admin user...")

	adminPassword, err := password.Generate(16, 2, 0, false, false)
	if err != nil {
		return "", errors.Wrap(err, "generating magento admin password")
	}

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"n98-magerun admin:user:create localadmin admin@example.com %s Local Admin",
			adminPassword,
		),
	); err != nil {
		return "", errors.Wrap(err, "creating magento admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *bootstrapper) installMagento1FlushCache() error {
	log.Println("Flushing cache...")

	if err := c.RunCmdEnvExec("n98-magerun cache:flush"); err != nil {
		return errors.Wrap(err, "flushing magento cache")
	}

	log.Println("...cache flushed.")

	return nil
}

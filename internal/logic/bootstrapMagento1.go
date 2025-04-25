package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

// bootstrap runs a full Magento 1 bootstrap process.
// Note: it will not install Magento 1 from zero, but only configures Magento 1's local.xml.
func (c *magento1) bootstrap() error {
	magentoVersion, err := c.MagentoVersion()
	if err != nil {
		return errors.Wrap(err, "getting magento version")
	}

	if !util.AskForConfirmation("Would you like to bootstrap Magento v" + magentoVersion.String() + "?") {
		return nil
	}

	log.Printf("Bootstrapping Magento %s...", magentoVersion.String())

	if err := c.prepare(); err != nil {
		return errors.Wrap(err, "error during bootstrap preparation")
	}

	if err := c.composerInstall(); err != nil {
		return errors.Wrap(err, "installing composer")
	}

	if err := c.generateLocalXML(); err != nil {
		return err
	}

	if err := c.configureBasicSettings(); err != nil {
		return err
	}

	adminPassword, err := c.configureAdminUser()
	if err != nil {
		return err
	}

	if err := c.flushCache(); err != nil {
		return err
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

func (c *magento1) composerInstall() error {
	if !util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "composer.json")) {
		return nil
	}

	if err := c.composerPreInstall(); err != nil {
		return errors.Wrap(err, "configuring composer")
	}

	if err := c.bootstrapper.composerInstall(); err != nil {
		return errors.Wrap(err, "running composer install")
	}

	if err := c.composerPostInstall(); err != nil {
		return errors.Wrap(err, "running composer post install configuration")
	}

	return nil
}

func (c *magento1) generateLocalXML() error {
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

	if err := templates.New().AppendTemplatesFromPathsStatic(
		localXMLTemplate, tmpList, []string{filepath.Join("templates", "magento1", "local.xml")},
	); err != nil {
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

func (c *magento1) configureBasicSettings() error {
	log.Println("Configuring Magento basic settings...")

	command := fmt.Sprintf("%s config:set web/unsecure/base_url http://%s/", c.magerunCommand(), c.TraefikFullDomain())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento base url")
	}

	command = fmt.Sprintf("%s config:set web/secure/base_url https://%s/", c.magerunCommand(), c.TraefikFullDomain())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento secure base url")
	}

	command = fmt.Sprintf("%s config:set web/secure/use_in_frontend 1", c.magerunCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento to use https on frontend")
	}

	command = fmt.Sprintf("%s config:set web/secure/use_in_adminhtml 1", c.magerunCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "setting magento to use https on admin")
	}

	log.Println("...Magento basic settings configured.")

	return nil
}

func (c *magento1) configureAdminUser() (string, error) {
	if c.SkipAdminUser() {
		return "", nil
	}

	log.Println("Creating admin user...")

	adminPassword := c.generatePassword()

	command := fmt.Sprintf(
		"%s admin:user:create localadmin admin@example.com %s Local Admin",
		c.magerunCommand(),
		adminPassword,
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return "", errors.Wrap(err, "creating magento admin user")
	}

	log.Println("...admin user created.")

	return adminPassword, nil
}

func (c *magento1) flushCache() error {
	log.Println("Flushing cache...")

	command := fmt.Sprintf("%s cache:flush", c.magerunCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "flushing magento cache")
	}

	log.Println("...cache flushed.")

	return nil
}

func (c *magento1) magerunCommand() string {
	return "n98-magerun"
}

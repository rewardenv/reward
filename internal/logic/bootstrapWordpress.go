package logic

import (
	"bytes"
	"container/list"
	"fmt"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/internal/templates"
	"github.com/rewardenv/reward/pkg/util"
)

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

	_, err = c.download()
	if err != nil {
		return fmt.Errorf("cannot download wordpress: %w", err)
	}

	err = c.installWordpressConfig()
	if err != nil {
		return fmt.Errorf("error during wordpress configuration: %w", err)
	}

	log.Printf("Base Url: https://%s", c.TraefikFullDomain())
	log.Println("...bootstrap process finished.")

	return nil
}

func (c *bootstrapper) installWordpressConfig() error {
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

	err := templates.New().AppendTemplatesFromPathsStatic(tpl, tmpList, tplPath)
	if err != nil {
		return fmt.Errorf("cannot load wordpress wp-config.php template: %w", err)
	}

	if c.DBPrefix() != "" {
		c.Set("wordpress_table_prefix", c.DBPrefix())
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = templates.New().ExecuteTemplate(tpl.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("cannot execute wordpress wp-config.php template: %w", err)
		}

		err = util.CreateDirAndWriteToFile(bs.Bytes(), configFilePath)
		if err != nil {
			return fmt.Errorf("cannot write wordpress wp-config.php file: %w", err)
		}
	}

	return nil
}

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

// bootstrapWordpress runs a full WordPress bootstrap process.
func (c *bootstrapper) bootstrapWordpress() error {
	if !util.AskForConfirmation("Would you like to bootstrap Wordpress?") {
		return nil
	}

	log.Println("Bootstrapping Wordpress...")

	err := c.prepare()
	if err != nil {
		return errors.Wrap(err, "preparing bootstrap")
	}

	_, err = c.download()
	if err != nil {
		return errors.Wrap(err, "downloading wordpress")
	}

	err = c.installWordpressConfig()
	if err != nil {
		return errors.Wrap(err, "configuring wordpress")
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
		return errors.Wrap(err, "loading wordpress wp-config.php template")
	}

	if c.DBPrefix() != "" {
		c.Set("wordpress_table_prefix", c.DBPrefix())
	}

	for e := tmpList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = templates.New().ExecuteTemplate(tpl.Lookup(tplName), &bs)
		if err != nil {
			return errors.Wrap(err, "executing wordpress wp-config.php template")
		}

		err = util.CreateDirAndWriteToFile(bs.Bytes(), configFilePath)
		if err != nil {
			return errors.Wrap(err, "writing wordpress wp-config.php file")
		}
	}

	return nil
}

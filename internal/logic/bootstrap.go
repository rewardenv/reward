package logic

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/rewardenv/reward/pkg/util"
)

type bootstrapper struct {
	*Client
	composerVerbosityFlag string
	debug                 bool
}

func newBootstrapper(c *Client) *bootstrapper {
	composerVerbosityFlag := "--verbose"
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
		if err := newBootstrapper(c).bootstrapMagento2(); err != nil {
			return errors.Wrap(err, "bootstrapping magento2")
		}
	case "magento1":
		if err := newBootstrapper(c).bootstrapMagento1(); err != nil {
			return errors.Wrap(err, "bootstrapping magento1")
		}
	case "wordpress":
		if err := newBootstrapper(c).bootstrapWordpress(); err != nil {
			return errors.Wrap(err, "bootstrapping wordpress")
		}
	case "shopware":
		if err := newBootstrapper(c).bootstrapShopware(); err != nil {
			return errors.Wrap(err, "bootstrapping shopware")
		}
	default:
		return errors.New("currently not supported for bootstrapping")
	}

	return nil
}

func (c *bootstrapper) prepare() error {
	log.Println("Preparing common services...")

	if err := c.RunCmdSvc([]string{"up"}); err != nil {
		return errors.Wrap(err, "starting services")
	}

	log.Println("...common services started.")
	log.Println("Preparing certificate...")

	if err := c.RunCmdSignCertificate([]string{c.TraefikDomain()}, true); err != nil {
		return errors.Wrap(err, "signing certificate")
	}

	log.Println("...certificate ready.")

	if !c.NoPull() {
		log.Println("Pulling images...")

		if err := c.RunCmdEnv([]string{"pull"}); err != nil {
			return errors.Wrap(err, "pulling env containers")
		}

		log.Println("...images pulled.")
	}

	log.Println("Preparing environment...")

	if err := c.RunCmdEnv([]string{"build"}); err != nil {
		return errors.Wrap(err, "building env containers")
	}

	if err := c.RunCmdEnv([]string{"up"}); err != nil {
		return errors.Wrap(err, "starting env containers")
	}

	log.Println("...environment ready.")

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
				return false, errors.Wrap(err, "determining magento version")
			}

			freshInstall = true

			if err = c.RunCmdEnvExec(
				fmt.Sprintf(
					"composer create-project %s --profile --no-install "+
						"--repository-url=https://repo.magento.com/ "+
						"magento/project-%s-edition=%s /tmp/magento-tmp/",
					composerVerbosityFlag,
					c.MagentoType(),
					magentoVersion.String(),
				),
			); err != nil {
				return false, errors.Wrap(err, "creating composer magento project")
			}

			if err = c.RunCmdEnvExec(
				fmt.Sprintf(
					`rsync %s -au --remove-source-files --chmod=D2775,F644 /tmp/magento-tmp/ /var/www/html/`,
					rsyncVerbosityFlag,
				),
			); err != nil {
				return false, errors.Wrap(err, "moving magento project install files")
			}

			log.Println("...Magento 2 composer project created.")
		}

	case "wordpress":
		if !util.FileExists(filepath.Join(c.Cwd(), c.WebRoot(), "index.php")) {
			log.Println("Downloading and installing WordPress...")

			freshInstall = true

			if err := c.RunCmdEnvExec("wget -qO /tmp/wordpress.tar.gz https://wordpress.org/latest.tar.gz"); err != nil {
				return false, errors.Wrap(err, "downloading wordpress")
			}

			if err := c.RunCmdEnvExec("tar -zxf /tmp/wordpress.tar.gz --strip-components=1 -C /var/www/html"); err != nil {
				return false, errors.Wrap(err, "extracting wordpress")
			}

			if err := c.RunCmdEnvExec("rm -f /tmp/wordpress.tar.gz"); err != nil {
				return false, errors.Wrap(err, "removing wordpress archive")
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

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"wget -qO /tmp/shopware.tar.gz https://github.com/shopware/%s/archive/refs/tags/v%s.tar.gz",
					path,
					version.Must(c.ShopwareVersion()).String(),
				),
			); err != nil {
				return false, errors.Wrap(err, "downloading shopware")
			}

			if err := c.RunCmdEnvExec("tar -zxf /tmp/shopware.tar.gz --strip-components=1 -C /var/www/html"); err != nil {
				return false, errors.Wrap(err, "extracting shopware")
			}

			if err := c.RunCmdEnvExec("rm -f /tmp/shopware.tar.gz"); err != nil {
				return false, errors.Wrap(err, "removing shopware archive")
			}

			log.Println("...Shopware downloaded.")
		}
	}

	return freshInstall, nil
}

func (c *bootstrapper) composerInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	log.Println("Installing composer dependencies...")

	if err := c.RunCmdEnvExec(
		fmt.Sprintf(
			"composer install %s --profile",
			c.composerVerbosityFlag,
		),
	); err != nil {
		return errors.Wrap(err, "installing composer dependencies")
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
		//nolint:lll
		if err := c.RunCmdEnvExec(fmt.Sprintf("%s alternatives %s --set composer %s/composer1", c.SudoCommand(), c.AlternativesArgs(), c.LocalBinPath())); err != nil {
			return errors.Wrap(err, "changing default composer version")
		}
	} else {
		log.Println("Setting default composer version to 2.x")

		// Change default Composer Version
		//nolint:lll
		if err := c.RunCmdEnvExec(fmt.Sprintf("%s alternatives %s --set composer %s/composer2", c.SudoCommand(), c.AlternativesArgs(), c.LocalBinPath())); err != nil {
			return errors.Wrap(err, "changing default composer version")
		}

		// Specific Composer Version
		if !c.ComposerVersion().Equal(version.Must(version.NewVersion("2.0.0"))) {
			if err := c.RunCmdEnvExec(
				fmt.Sprintf("%s composer self-update %s", c.SudoCommand(), c.ComposerVersion().String()),
			); err != nil {
				return errors.Wrap(err, "changing default composer version")
			}
		}
	}

	// Composer Install
	if c.Parallel() && composerVersion < 2 {
		if err := c.RunCmdEnvExec(
			fmt.Sprintf(
				"composer global require %s --profile hirak/prestissimo",
				c.composerVerbosityFlag,
			),
		); err != nil {
			return errors.Wrap(err, "installing hirak/prestissimo composer module")
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

			if err := c.RunCmdEnvExec(
				fmt.Sprintf(
					"composer global remove %s --profile hirak/prestissimo",
					c.composerVerbosityFlag,
				),
			); err != nil {
				return errors.Wrap(err, "removing hirak/prestissimo module")
			}

			log.Println("...hirak/prestissimo composer module removed.")
		}
	}

	return nil
}

func (c *Client) RunCmdEnvExec(args string) error {
	return c.RunCmdEnv(append([]string{"exec", "-T", c.DefaultSyncedContainer(c.EnvType()), "bash", "-c"}, args))
}

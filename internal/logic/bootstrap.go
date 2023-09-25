package logic

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/go-version"
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
				return false, fmt.Errorf("cannot download wordpress: %w", err)
			}

			err = c.RunCmdEnvExec("tar -zxf /tmp/wordpress.tar.gz --strip-components=1 -C /var/www/html")
			if err != nil {
				return false, fmt.Errorf("cannot extract wordpress: %w", err)
			}

			err = c.RunCmdEnvExec("rm -f /tmp/wordpress.tar.gz")
			if err != nil {
				return false, fmt.Errorf("cannot remove wordpress archive: %w", err)
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
				return false, fmt.Errorf("cannot download shopware: %w", err)
			}

			err = c.RunCmdEnvExec("tar -zxf /tmp/shopware.tar.gz --strip-components=1 -C /var/www/html")
			if err != nil {
				return false, fmt.Errorf("cannot extract shopware: %w", err)
			}

			err = c.RunCmdEnvExec("rm -f /tmp/shopware.tar.gz")
			if err != nil {
				return false, fmt.Errorf("cannot remove shopware archive: %w", err)
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
		//nolint:lll
		err := c.RunCmdEnvExec(fmt.Sprintf("%s alternatives %s --set composer %s/composer1", c.SudoCommand(), c.AlternativesArgs(), c.LocalBinPath()))
		if err != nil {
			return fmt.Errorf("cannot change default composer version: %w", err)
		}
	} else {
		log.Println("Setting default composer version to 2.x")

		// Change default Composer Version
		//nolint:lll
		err := c.RunCmdEnvExec(fmt.Sprintf("%s alternatives %s --set composer %s/composer2", c.SudoCommand(), c.AlternativesArgs(), c.LocalBinPath()))
		if err != nil {
			return fmt.Errorf("cannot change default composer version: %w", err)
		}

		// Specific Composer Version
		if !c.ComposerVersion().Equal(version.Must(version.NewVersion("2.0.0"))) {
			err = c.RunCmdEnvExec(fmt.Sprintf("%s composer self-update %s", c.SudoCommand(), c.ComposerVersion().String()))
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

package logic

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	log "github.com/sirupsen/logrus"
)

type bootstrapper struct {
	*Client
}

func newBootstrapper(c *Client) *bootstrapper {
	return &bootstrapper{
		Client: c,
	}
}

type magento1 struct {
	*bootstrapper
}
type magento2 struct {
	*bootstrapper
}
type wordpress struct {
	*bootstrapper
}
type shopware struct {
	*bootstrapper
}

// RunCmdBootstrap represents the bootstrap command.
func (c *Client) RunCmdBootstrap() error {
	switch c.EnvType() {
	case "magento2":
		b := magento2{newBootstrapper(c)}
		if err := b.bootstrap(); err != nil {
			return errors.Wrap(err, "bootstrapping magento2")
		}

	case "magento1":
		b := magento1{newBootstrapper(c)}
		if err := b.bootstrap(); err != nil {
			return errors.Wrap(err, "bootstrapping magento1")
		}

	case "wordpress":
		b := wordpress{newBootstrapper(c)}
		if err := b.bootstrap(); err != nil {
			return errors.Wrap(err, "bootstrapping wordpress")
		}

	case "shopware":
		b := shopware{newBootstrapper(c)}
		if err := b.bootstrap(); err != nil {
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

func (c *bootstrapper) composerInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	log.Println("Installing composer dependencies...")

	command := fmt.Sprintf("%s install --profile", c.composerCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
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

	if c.ComposerVersion().LessThan(version.Must(version.NewVersion("2.0.0"))) {
		if err := c.setComposer1(); err != nil {
			return err
		}
	} else {
		if err := c.setComposer2(); err != nil {
			return err
		}
	}

	log.Println("...composer configured.")

	return nil
}

func (c *bootstrapper) setComposer2() error {
	log.Println("Setting default composer version to 2.x")

	// Change default Composer Version
	command := fmt.Sprintf("%s alternatives %s --set composer %s/composer2",
		c.SudoCommand(),
		c.AlternativesArgs(),
		c.LocalBinPath(),
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "changing default composer version")
	}

	// Specific Composer Version
	if !c.ComposerVersion().Equal(version.Must(version.NewVersion("2.0.0"))) {
		command = fmt.Sprintf("%s %s self-update %s",
			c.SudoCommand(),
			c.composerCommand(),
			c.ComposerVersion().String(),
		)
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "changing default composer version")
		}
	}

	return nil
}

func (c *bootstrapper) setComposer1() error {
	log.Println("Setting default composer version to 1.x")

	// Change default Composer Version
	command := fmt.Sprintf("%s alternatives %s --set composer %s/composer1",
		c.SudoCommand(),
		c.AlternativesArgs(),
		c.LocalBinPath(),
	)
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "changing default composer version")
	}

	// Configure parallelism for composer 1 using hirak/prestissimo
	if !c.Parallel() {
		return nil
	}

	command = fmt.Sprintf("%s global require --profile hirak/prestissimo", c.composerCommand())
	if err := c.RunCmdEnvExec(command); err != nil {
		return errors.Wrap(err, "installing hirak/prestissimo composer module")
	}

	return nil
}

func (c *bootstrapper) composerPostInstall() error {
	if c.SkipComposerInstall() {
		return nil
	}

	if c.ComposerVersion().LessThan(version.Must(version.NewVersion("2.0.0"))) {
		if !c.Parallel() {
			return nil
		}

		log.Println("Removing hirak/prestissimo composer module...")

		command := fmt.Sprintf("%s global remove --profile hirak/prestissimo", c.composerCommand())
		if err := c.RunCmdEnvExec(command); err != nil {
			return errors.Wrap(err, "removing hirak/prestissimo module")
		}

		log.Println("...hirak/prestissimo composer module removed.")
	}

	return nil
}

func (c *bootstrapper) RunCmdEnvExec(args string) error {
	return c.RunCmdEnv(append([]string{"exec", "-T", c.DefaultSyncedContainer(c.EnvType()), "bash", "-c"}, args))
}

func (c *bootstrapper) generatePassword() string {
	return password.MustGenerate(16, 2, 0, false, false)
}

func (c *bootstrapper) composerCommand() string {
	verbosity := "--verbose"
	if c.IsDebug() {
		verbosity = "-vvv"
	}

	return "composer " + verbosity
}

func (c *bootstrapper) rsyncCommand() string {
	verbosity := "-v"
	if c.IsDebug() {
		verbosity = "-vv"
	}

	return "rsync " + verbosity
}

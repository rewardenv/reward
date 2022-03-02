package commands

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/rewardenv/reward/internal/core"

	log "github.com/sirupsen/logrus"
)

// SvcCmd builds up the contents for the svc command.
func SvcCmd(args []string) error {
	if len(args) == 0 {
		args = append(args, "--help")

		err := SvcRunDockerCompose(args, true)
		if err != nil {
			return err
		}

		return nil
	}

	if core.ContainsString(args, "up") {
		sslDir := filepath.Join(core.GetAppHomeDir(), core.SslBaseDir)

		serviceDomain := core.GetServiceDomain()

		log.Debugln("Service Domain:", serviceDomain)

		if !core.CheckFileExists(filepath.Join(sslDir, "certs", serviceDomain+".crt.pem")) {
			err := SignCertificateCmd([]string{serviceDomain})
			if err != nil {
				return err
			}
		}

		if !core.CheckFileExists(filepath.Join(core.GetAppHomeDir(), "etc/traefik/traefik.yml")) {
			err := SvcGenerateTraefikConfig()
			if err != nil {
				return err
			}
		}

		err := SvcGenerateTraefikDynamicConfig()
		if err != nil {
			return err
		}

		newArgs := args

		if !core.ContainsString(args, "-d") && !core.ContainsString(args, "--detach") {
			for i, arg := range args {
				if arg == "up" {
					newArgs = []string{}
					newArgs = append(newArgs, args[:i+1]...)
					newArgs = append(newArgs, "--detach")
					newArgs = append(newArgs, args[i+1:]...)
				}
			}
		}

		args = newArgs
	}

	// pass orchestration through to docker-compose
	err := SvcRunDockerCompose(args, false)
	if err != nil {
		return err
	}

	// connect peered service containers to environment networks when 'svc up' is run
	networks, err := core.GetDockerNetworksWithLabel(fmt.Sprintf("label=dev.%v.environment.name", core.AppName))
	if err != nil {
		return err
	}

	for _, network := range networks {
		err = core.DockerPeeredServices("connect", network)
		if err != nil {
			return err
		}
	}

	return nil
}

// SvcRunDockerCompose function is a wrapper around the docker-compose command.
//   It appends the current directory and current project name to the args.
//   It also changes the output if the OS StdOut is suppressed.
func SvcRunDockerCompose(args []string, suppressOsStdOut ...bool) error {
	passedArgs := []string{
		"--project-directory",
		core.GetAppHomeDir(),
		"--project-name",
		core.AppName,
	}
	passedArgs = append(passedArgs, args...)

	// run docker-compose command
	out, err := SvcBuildDockerComposeCommand(passedArgs, suppressOsStdOut...)
	re := regexp.MustCompile("(?m)[\r\n]+^.*--file.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*--project-name.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*--project-directory.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*default: docker-compose.yml.*$")
	out = re.ReplaceAllString(out, "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*default: directory name.*$")
	out = re.ReplaceAllString(out, "")
	out = strings.ReplaceAll(out, "docker-compose", "env")

	_, _ = fmt.Fprint(os.Stdout, out)

	if err != nil {
		return err
	}

	return nil
}

// SvcBuildDockerComposeCommand builds up the docker-compose command by passing it the previously built templates for
// the common services..
func SvcBuildDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	svcTemplate := new(template.Template)
	svcTemplateList := list.New()

	err := SvcBuildDockerComposeTemplate(svcTemplate, svcTemplateList)
	if err != nil {
		return "", err
	}

	svcDockerComposeConfigs, err := core.ConvertTemplateToComposeConfig(svcTemplate, svcTemplateList)
	if err != nil {
		return "", err
	}

	out, err := core.RunDockerComposeWithConfig(args, svcDockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

// SvcBuildDockerComposeTemplate builds the templates which are used to invoke docker-compose for the common services.
func SvcBuildDockerComposeTemplate(t *template.Template, templateList *list.List) error {
	templatePaths := []string{
		"templates/_services/docker-compose.yml",
	}

	log.Traceln("Template Paths: ", templatePaths)

	err := core.AppendTemplatesFromPathsStatic(t, templateList, templatePaths)
	if err != nil {
		return err
	}

	return nil
}

// SvcGenerateTraefikConfig generates the default traefik configuration.
func SvcGenerateTraefikConfig() error {
	var configBuffer bytes.Buffer

	traefikTemplate := new(template.Template)
	traefikTemplateList := list.New()
	traefikConfig := "templates/_traefik/traefik.yml"

	err := core.AppendTemplatesFromPathsStatic(traefikTemplate, traefikTemplateList, []string{traefikConfig})
	if err != nil {
		return err
	}

	for e := traefikTemplateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err := core.ExecuteTemplate(traefikTemplate.Lookup(tplName), &configBuffer)
		if err != nil {
			return err
		}
	}

	err = core.CreateDirAndWriteBytesToFile(
		configBuffer.Bytes(),
		filepath.Join(core.GetAppHomeDir(), "etc/traefik/traefik.yml"),
		0o644,
	)
	if err != nil {
		return err
	}

	return nil
}

// SvcGenerateTraefikDynamicConfig generates the dynamic traefik configuration.
func SvcGenerateTraefikDynamicConfig() error {
	traefikConfig := fmt.Sprintf(
		`tls:
  stores:
    default:
    defaultCertificate:
      certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
  certificates:`, core.GetServiceDomain(),
	)

	files, err := filepath.Glob(filepath.Join(core.GetAppHomeDir(), "ssl/certs", "*.crt.pem"))
	if err != nil {
		return err
	}

	log.Debugln(files)

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".crt.pem")

		log.Debugln(name)
		log.Debugln(filepath.Ext(name))

		traefikConfig = traefikConfig + fmt.Sprintf(
			`
    - certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
`, name,
		)
	}

	err = core.CreateDirAndWriteBytesToFile(
		[]byte(traefikConfig), filepath.Join(core.GetAppHomeDir(), "etc/traefik", "dynamic.yml"), 0o644,
	)

	return err
}

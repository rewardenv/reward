package internal

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// SvcCmd builds up the contents for the svc subcommand.
func SvcCmd(args []string) error {
	if err := CheckDockerIsRunning(); err != nil {
		return err
	}

	// if err := EnvCheck(); err != nil {
	// 	return err
	// }

	if len(args) == 0 {
		args = append(args, "--help")

		err := SvcRunDockerCompose(args, true)
		if err != nil {
			return err
		}

		return nil
	}

	if ContainsString(args, "up") {
		sslDir := filepath.Join(GetAppHomeDir(), sslBaseDir)

		serviceDomain := GetServiceDomain()

		log.Debugln("Service Domain:", serviceDomain)

		if !CheckFileExists(filepath.Join(sslDir, "certs", serviceDomain+".crt.pem")) {
			err := SignCertificateCmd([]string{serviceDomain})
			if err != nil {
				return err
			}
		}

		if !CheckFileExists(filepath.Join(GetAppHomeDir(), "etc/traefik/traefik.yml")) {
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

		if !ContainsString(args, "-d") && !ContainsString(args, "--detach") {
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
	networks, err := GetDockerNetworksWithLabel(fmt.Sprintf("label=dev.%v.environment.name", AppName))
	if err != nil {
		return err
	}

	for _, network := range networks {
		err = DockerPeeredServices("connect", network)
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
		GetAppHomeDir(),
		"--project-name",
		AppName,
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

func SvcBuildDockerComposeCommand(args []string, suppressOsStdOut ...bool) (string, error) {
	svcTemplate := new(template.Template)
	svcTemplateList := list.New()

	err := SvcBuildDockerComposeTemplate(svcTemplate, svcTemplateList)
	if err != nil {
		return "", err
	}

	svcDockerComposeConfigs, err := ConvertTemplateToComposeConfig(svcTemplate, svcTemplateList)
	if err != nil {
		return "", err
	}

	out, err := RunDockerComposeWithConfig(args, svcDockerComposeConfigs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

func SvcBuildDockerComposeTemplate(t *template.Template, templateList *list.List) error {
	templatePaths := []string{
		"templates/_services/docker-compose.yml",
	}

	log.Traceln("Template Paths: ", templatePaths)

	err := AppendTemplatesFromPathsStatic(t, templateList, templatePaths)
	if err != nil {
		return err
	}

	return nil
}

func SvcGenerateTraefikConfig() error {
	var configBuffer bytes.Buffer

	traefikTemplate := new(template.Template)
	traefikTemplateList := list.New()
	traefikConfig := "templates/_traefik/traefik.yml"

	err := AppendTemplatesFromPathsStatic(traefikTemplate, traefikTemplateList, []string{traefikConfig})
	if err != nil {
		return err
	}

	for e := traefikTemplateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err := ExecuteTemplate(traefikTemplate.Lookup(tplName), &configBuffer)
		if err != nil {
			return err
		}
	}

	err = CreateDirAndWriteBytesToFile(configBuffer.Bytes(),
		filepath.Join(GetAppHomeDir(), "etc/traefik/traefik.yml"),
		0o644,
	)
	if err != nil {
		return err
	}

	return nil
}

func SvcGenerateTraefikDynamicConfig() error {
	traefikConfig := fmt.Sprintf(`tls:
  stores:
    default:
    defaultCertificate:
      certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
  certificates:`, GetServiceDomain())

	files, err := filepath.Glob(filepath.Join(GetAppHomeDir(), "ssl/certs", "*.crt.pem"))
	if err != nil {
		return err
	}

	log.Debugln(files)

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".crt.pem")

		log.Debugln(name)
		log.Debugln(filepath.Ext(name))

		traefikConfig = traefikConfig + fmt.Sprintf(`
    - certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
`, name)
	}

	err = CreateDirAndWriteBytesToFile(
		[]byte(traefikConfig), filepath.Join(GetAppHomeDir(), "etc/traefik", "dynamic.yml"), 0o644)

	return err
}

func SvcEnabled(name string) bool {
	key := AppName + "_" + name
	if viper.IsSet(key) {
		viper.GetBool(key)
	}

	return true
}

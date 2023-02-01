package templates

import (
	"bytes"
	"container/list"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/docker/cli/cli/compose/loader"
	compose "github.com/docker/cli/cli/compose/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/rewardenv/reward/assets"
	"github.com/rewardenv/reward/pkg/util"
)

type Client struct {
	fs embed.FS
}

func New() *Client {
	return &Client{
		fs: assets.Assets,
	}
}

// Cwd returns the current working directory.
func (c *Client) Cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}

	return cwd
}

func (c *Client) AppName() string {
	return viper.GetString("app_name")
}

// AppHomeDir returns the application's home directory.
func (c *Client) AppHomeDir() string {
	return viper.GetString(fmt.Sprintf("%s_home_dir", c.AppName()))
}

// ExecuteTemplate executes the templates, appending some specific template functions to the execution.
func (c *Client) ExecuteTemplate(t *template.Template, buffer io.Writer) error {
	data := viper.AllSettings()

	err := t.Funcs(sprig.TxtFuncMap()).
		Funcs(funcMap()).
		ExecuteTemplate(buffer, t.Name(), data)
	if err != nil {
		return fmt.Errorf("cannot execute template %s: %w", t.Name(), err)
	}

	return nil
}

// AppendTemplatesFromPathsStatic appends templates to t from templateList list searching them in paths path list.
// This function looks up templates built to the application's binary (static files).
// If it cannot find templates it's not going to fail.
// If a template with the same name already exists, it's going to skip that template.
func (c *Client) AppendTemplatesFromPathsStatic(t *template.Template, templateList *list.List, paths []string) error {
	for _, path := range paths {
		// Use the regular paths without filepath.Join() because the assets are embedded with forward slashes.
		templatePath := strings.ReplaceAll(path, "\\", "/")

		searchT := t.Lookup(templatePath)
		if searchT != nil {
			continue
		}

		content, err := c.fs.ReadFile(templatePath)
		if err != nil {
			continue
		}

		child, err := template.New(templatePath).Funcs(funcMap()).Parse(string(content))
		if err != nil {
			return fmt.Errorf("cannot parse template %s: %w", path, err)
		}

		_, err = t.AddParseTree(child.Name(), child.Tree)
		if err != nil {
			return fmt.Errorf("error adding template %s to tree: %w", child.Name(), err)
		}

		templateList.PushBack(child.Name())
	}

	return nil
}

// AppendTemplatesFromPaths appends templates to t from templateList list searching them in paths path list.
// If it cannot find templates it's not going to fail.
// If a template with the same name already exists, it's going to skip that template.
func (c *Client) AppendTemplatesFromPaths(t *template.Template, templateList *list.List, paths []string) error {
	for _, path := range paths {
		// Lookup templates in the current directory.
		{
			// Make sure to convert the path slashes to Windows \ format if we're on Windows.
			filePath := filepath.Join(c.Cwd(), fmt.Sprintf(".%s", c.AppName()), path)

			if !util.FileExists(filePath) {
				log.Tracef("Template not found in $CWD: %s", path)

				continue
			}

			searchT := t.Lookup(path)
			if searchT != nil {
				log.Tracef("Template already defined: %s. Skipping.", path)

				continue
			}

			child, err := template.New(path).Funcs(funcMap()).ParseFiles(filePath)
			if err != nil {
				return fmt.Errorf("cannot parse template %s: %w", path, err)
			}

			_, err = t.AddParseTree(child.Name(), child.Lookup(filepath.Base(filePath)).Tree)
			if err != nil {
				return fmt.Errorf("error adding template %s: %w", child.Name(), err)
			}

			templateList.PushBack(child.Name())
		}

		// Lookup templates in the home directory.
		{
			filePath := filepath.Join(c.AppHomeDir(), path)
			if !util.FileExists(filePath) {
				log.Traceln("Template not found in app home:", path)

				continue
			}

			searchT := t.Lookup(path)
			if searchT != nil {
				log.Tracef("Template already defined: %s. Skipping.", path)

				continue
			}

			child, err := template.New(path).Funcs(funcMap()).ParseFiles(filePath)
			if err != nil {
				return fmt.Errorf("cannot parse template %s: %w", path, err)
			}

			_, err = t.AddParseTree(child.Name(), child.Lookup(filepath.Base(filePath)).Tree)
			if err != nil {
				return fmt.Errorf("error adding template %s: %w", child.Name(), err)
			}

			templateList.PushBack(child.Name())
		}
	}

	return nil
}

// AppendEnvironmentTemplates tries to look up all the templates dedicated for an environment type.
func (c *Client) AppendEnvironmentTemplates(
	t *template.Template,
	templateList *list.List,
	partialName string,
	envType string,
) error {
	staticTemplatePaths := []string{
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			"includes",
			fmt.Sprintf("%s.base.yml", partialName)),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			"includes",
			fmt.Sprintf("%s.%s.yml", partialName, runtime.GOOS),
		),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			envType,
			fmt.Sprintf("%s.base.yml", partialName)),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			envType,
			fmt.Sprintf("%s.%s.yml", partialName, runtime.GOOS),
		),
	}
	templatePaths := []string{
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			"includes",
			fmt.Sprintf("%s.base.yml", partialName),
		),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			"includes",
			fmt.Sprintf("%s.%s.yml", partialName, runtime.GOOS),
		),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			envType,
			fmt.Sprintf("%s.base.yml", partialName),
		),
		filepath.Join(
			"templates",
			"docker-compose",
			"environments",
			envType,
			fmt.Sprintf("%s.%s.yml", partialName, runtime.GOOS),
		),
	}

	// First read the templates from the current directory. If they exist we will use them. If the don't
	//   then we will append them from the static content.
	err := c.AppendTemplatesFromPaths(t, templateList, templatePaths)
	if err != nil {
		return fmt.Errorf("cannot append templates from local paths: %w", err)
	}

	err = c.AppendTemplatesFromPathsStatic(t, templateList, staticTemplatePaths)
	if err != nil {
		return fmt.Errorf("cannot append static templates: %w", err)
	}

	return nil
}

// AppendMutagenTemplates is going to add mutagen configuration templates.
func (c *Client) AppendMutagenTemplates(
	t *template.Template,
	templateList *list.List,
	partialName string,
	envType string,
) error {
	staticTemplatePaths := []string{
		filepath.Join("templates", "docker-compose", "environments",
			envType,
			fmt.Sprintf("%s.%s.yml", envType, partialName)),
		filepath.Join(
			"templates", "docker-compose", "environments", envType,
			fmt.Sprintf("%s.%s.%s.yml", envType, partialName, runtime.GOOS),
		),
	}

	for _, v := range staticTemplatePaths {
		content, err := assets.Assets.ReadFile(v)
		if err != nil {
			log.Traceln(err)

			continue
		}

		child, err := template.New(v).Funcs(funcMap()).Parse(string(content))
		if err != nil {
			return fmt.Errorf("cannot parse template %s: %w", v, err)
		}

		_, err = t.AddParseTree(child.Name(), child.Tree)
		if err != nil {
			return fmt.Errorf("error adding template %s: %w", child.Name(), err)
		}

		templateList.PushBack(child.Name())
	}

	return nil
}

// SvcBuildDockerComposeTemplate builds the templates which are used to invoke docker-compose for the common services.
func (c *Client) RunCmdSvcBuildDockerComposeTemplate(t *template.Template, templateList *list.List) error {
	templatePaths := []string{
		"templates/docker-compose/common-services/docker-compose.yml",
	}

	err := c.AppendTemplatesFromPathsStatic(t, templateList, templatePaths)
	if err != nil {
		return fmt.Errorf("cannot append common-services/docker-compose.yml static template: %w", err)
	}

	return nil
}

// ConvertTemplateToComposeConfig iterates through all the templates and converts them to docker-compose configurations.
func (c *Client) ConvertTemplateToComposeConfig(
	t *template.Template,
	templateList *list.List,
) (compose.ConfigDetails, error) {
	log.Debugln("Converting templates to docker-compose configurations...")

	var (
		configs     = new(compose.ConfigDetails)
		configFiles = new([]compose.ConfigFile)
		bs          bytes.Buffer
	)

	for e := templateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err := c.ExecuteTemplate(t.Lookup(tplName), &bs)
		if err != nil {
			return *configs, fmt.Errorf("failed to execute template %s: %w", tplName, err)
		}

		templateBytes := bs.Bytes()
		templateBytes = append(templateBytes, []byte("\n")...)

		composeConfig, err := loader.ParseYAML(templateBytes)
		if err != nil {
			return *configs, fmt.Errorf("error parsing template %s: %w", tplName, err)
		}

		configFile := compose.ConfigFile{
			Filename: tplName,
			Config:   composeConfig,
		}

		*configFiles = append(*configFiles, configFile)
	}

	configs.ConfigFiles = *configFiles

	return *configs, nil
}

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	delete(f, "env")
	delete(f, "expandenv")

	extra := template.FuncMap{
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },
		"lookup": func(string, string, string, string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
		"isEnabled": isEnabled,
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

// isEnabled returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string).
func isEnabled(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return false
	}

	switch g.Kind() { //nolint:exhaustive
	case reflect.String:
		return strings.EqualFold(g.String(), "true") || g.String() == "1"
	case reflect.Bool:
		return g.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 1
	default:
		return false
	}
}

// GenerateMutagenTemplateFile generates mutagen configuration from template if it doesn't exist.
func (c *Client) GenerateMutagenTemplateFile(path, envType string) error {
	if util.FileExists(path) {
		// Mutagen sync file already exists, skipping.

		return nil
	}

	var (
		bs                  bytes.Buffer
		mutagenTemplate     = new(template.Template)
		mutagenTemplateList = list.New()
	)

	err := c.AppendMutagenTemplates(mutagenTemplate, mutagenTemplateList, "mutagen", envType)
	if err != nil {
		return fmt.Errorf("an error occurred while appending mutagen templates: %w", err)
	}

	for e := mutagenTemplateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = c.ExecuteTemplate(mutagenTemplate.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("an error occurred while executing mutagen template: %w", err)
		}
	}

	err = util.CreateDirAndWriteToFile(bs.Bytes(), path, 0o640)
	if err != nil {
		return fmt.Errorf("cannot create mutagen sync file: %w", err)
	}

	return nil
}

// SvcGenerateTraefikConfig generates the default traefik configuration.
func (c *Client) SvcGenerateTraefikConfig() error {
	var (
		bs      bytes.Buffer
		tpl     = template.New("traefik")
		tplList = list.New()
	)

	err := c.AppendTemplatesFromPathsStatic(
		tpl,
		tplList,
		[]string{"templates/traefik/traefik.yml"},
	)
	if err != nil {
		return fmt.Errorf("cannot append traefik.yml template: %w", err)
	}

	for e := tplList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err = c.ExecuteTemplate(tpl.Lookup(tplName), &bs)
		if err != nil {
			return fmt.Errorf("cannot execute traefik template %s: %w", tplName, err)
		}
	}

	err = util.CreateDirAndWriteToFile(
		bs.Bytes(),
		filepath.Join(c.AppHomeDir(), "etc/traefik/traefik.yml"),
		0o644,
	)
	if err != nil {
		return fmt.Errorf("cannot write traefik template file: %w", err)
	}

	return nil
}

// SvcGenerateTraefikDynamicConfig generates the dynamic traefik configuration.
func (c *Client) SvcGenerateTraefikDynamicConfig(svcDomain string) error {
	traefikConfig := fmt.Sprintf(
		`tls:
  stores:
    default:
    defaultCertificate:
      certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
  certificates:`, svcDomain,
	)

	files, err := filepath.Glob(filepath.Join(c.AppHomeDir(), "ssl/certs", "*.crt.pem"))
	if err != nil {
		return fmt.Errorf("cannot list ssl certificates: %w", err)
	}

	log.Debugf("Available certificates: %s", files)

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".crt.pem")

		log.Tracef("Certificate file name: %s", name)
		log.Tracef("Certificate domain: %s", filepath.Ext(name))

		traefikConfig += fmt.Sprintf(
			`
    - certFile: /etc/ssl/certs/%[1]v.crt.pem
      keyFile: /etc/ssl/certs/%[1]v.key.pem
`, name,
		)
	}

	err = util.CreateDirAndWriteToFile(
		[]byte(traefikConfig), filepath.Join(c.AppHomeDir(), "etc/traefik", "dynamic.yml"), 0o644,
	)
	if err != nil {
		return fmt.Errorf("cannot write traefik dynamic configuration file: %w", err)
	}

	return nil
}

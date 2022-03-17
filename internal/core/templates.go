package core

import (
	"bytes"
	"container/list"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"text/template"

	"github.com/rewardenv/reward/internal"
	"gopkg.in/yaml.v3"

	"github.com/Masterminds/sprig"
	"github.com/docker/cli/cli/compose/loader"
	compose "github.com/docker/cli/cli/compose/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	TmpFilesList        = list.New()
	customTemplateFuncs = map[string]interface{}{
		"isEnabledPermissive": isEnabledPermissive,
		"isEnabledStrict":     isEnabledStrict,
	}
	composeBuffer bytes.Buffer
)

// AppendTemplatesFromPaths appends templates to t from templateList list searching them in paths path list.
// If it cannot find templates it's not going to fail.
// If a template with the same name already exists, it's going to skip that template.
func AppendTemplatesFromPaths(t *template.Template, templateList *list.List, paths []string) error {
	log.Debugln()

	for _, path := range paths {
		templatePath := path
		filePath := filepath.Join(GetCwd(), "."+AppName, path)

		// Check for template in CWD
		if CheckFileExists(filePath) {
			log.Traceln("appending template from $CWD:", templatePath)

			searchT := t.Lookup(templatePath)
			if searchT != nil {
				log.Traceln("template already defined:", templatePath)
				continue
			}

			child, err := template.New(templatePath).
				Funcs(sprig.TxtFuncMap()).
				Funcs(customTemplateFuncs).
				ParseFiles(filePath)
			if err != nil {
				return err
			}

			_, err = t.AddParseTree(child.Name(), child.Lookup(filepath.Base(filePath)).Tree)
			if err != nil {
				return err
			}

			templateList.PushBack(child.Name())
		} else {
			log.Traceln("template not found in $CWD:", templatePath)
		}

		// Check for template in home
		filePath = filepath.Join(GetAppHomeDir(), templatePath)
		if CheckFileExists(filePath) {
			log.Traceln("appending template from app home:", templatePath)

			searchT := t.Lookup(templatePath)
			if searchT != nil {
				log.Traceln("template was already defined:", templatePath)
				continue
			}

			child, err := template.New(templatePath).
				Funcs(sprig.TxtFuncMap()).
				Funcs(customTemplateFuncs).
				ParseFiles(filePath)
			if err != nil {
				return err
			}

			_, err = t.AddParseTree(child.Name(), child.Lookup(filepath.Base(filePath)).Tree)
			if err != nil {
				return err
			}

			templateList.PushBack(child.Name())
		} else {
			log.Traceln("template not found in app home:", templatePath)
		}
	}

	return nil
}

// AppendTemplatesFromPathsStatic appends templates to t from templateList list searching them in paths path list.
// This function looks up templates built to the application's binary (static files).
// If it cannot find templates it's not going to fail.
// If a template with the same name already exists, it's going to skip that template.
func AppendTemplatesFromPathsStatic(t *template.Template, templateList *list.List, paths []string) error {
	log.Traceln(paths)

	for _, path := range paths {
		templatePath := filepath.Join(path)

		searchT := t.Lookup(templatePath)
		if searchT != nil {
			log.Traceln("template already defined:", templatePath)
			continue
		}

		content, err := internal.Asset(templatePath)
		if err != nil {
			log.Traceln(err)
		} else {
			log.Traceln("creating template:", templatePath)

			child, err := template.New(templatePath).
				Funcs(sprig.TxtFuncMap()).
				Funcs(customTemplateFuncs).
				Parse(string(content))
			if err != nil {
				return err
			}

			log.Traceln("appending template:", templatePath)
			_, err = t.AddParseTree(child.Name(), child.Tree)
			if err != nil {
				return err
			}
			templateList.PushBack(child.Name())
		}
	}

	return nil
}

// AppendEnvironmentTemplates tries to look up all the templates dedicated for an environment type.
func AppendEnvironmentTemplates(t *template.Template, templateList *list.List, partialName string) error {
	log.Debugln()

	envType := GetEnvType()
	staticTemplatePaths := []string{
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
	}
	templatePaths := []string{
		filepath.Join(
			"templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName),
		),
		filepath.Join(
			"templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS),
		),
		filepath.Join(
			"templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName),
		),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
	}

	log.Traceln("template paths:")
	log.Traceln(staticTemplatePaths, templatePaths)

	// First read the templates from the current directory. If they exist we will use them. If the don't
	//   then we will append them from the static content.
	err := AppendTemplatesFromPaths(t, templateList, templatePaths)
	if err != nil {
		return err
	}

	err = AppendTemplatesFromPathsStatic(t, templateList, staticTemplatePaths)
	if err != nil {
		return err
	}

	return nil
}

// AppendMutagenTemplates is going to add mutagen configuration templates.
func AppendMutagenTemplates(t *template.Template, templateList *list.List, partialName string) error {
	log.Debugln()

	envType := GetEnvType()
	staticTemplatePaths := []string{
		filepath.Join("templates/environments", envType, fmt.Sprintf("%v.%v.yml", envType, partialName)),
		filepath.Join(
			"templates/environments", envType, fmt.Sprintf("%v.%v.%v.yml", envType, partialName, runtime.GOOS),
		),
	}

	log.Traceln("template paths:")
	log.Traceln(staticTemplatePaths, staticTemplatePaths)

	for _, v := range staticTemplatePaths {
		content, err := internal.Asset(v)
		if err != nil {
			log.Traceln(err)
		} else {
			log.Traceln("appending template:", v)

			child, err := template.New(v).
				Funcs(sprig.TxtFuncMap()).
				Funcs(customTemplateFuncs).
				Parse(string(content))
			if err != nil {
				return err
			}
			_, err = t.AddParseTree(child.Name(), child.Tree)
			if err != nil {
				return err
			}
			templateList.PushBack(child.Name())
		}
	}

	return nil
}

// ExecuteTemplate executes the templates, appending some specific template functions to the execution.
func ExecuteTemplate(t *template.Template, buffer io.Writer) error {
	log.Debugln("Executing template:", t.Name())
	log.Traceln(viper.AllSettings())
	log.Traceln(t.DefinedTemplates())

	err := t.Funcs(sprig.TxtFuncMap()).
		Funcs(customTemplateFuncs).
		ExecuteTemplate(buffer, t.Name(), viper.AllSettings())

	return err
}

// ConvertTemplateToComposeConfig iterates through all the templates and converts them to docker-compose configurations.
func ConvertTemplateToComposeConfig(t *template.Template, templateList *list.List) (compose.ConfigDetails, error) {
	log.Debugln()

	configs := new(compose.ConfigDetails)
	configFiles := new([]compose.ConfigFile)

	log.Debugln("Converting templates to docker-compose configs...")
	log.Traceln("Templates:", t.DefinedTemplates())

	for e := templateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err := ExecuteTemplate(t.Lookup(tplName), &composeBuffer)
		if err != nil {
			return *configs, err
		}

		templateBytes := composeBuffer.Bytes()

		log.Traceln("docker-compose template:", string(templateBytes))

		composeConfig, err := loader.ParseYAML(templateBytes)

		log.Traceln("Parsing template:")
		log.Traceln(tplName)
		log.Traceln(composeConfig)

		if err != nil {
			return *configs, err
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

// RunDockerComposeWithConfig calls docker-compose with the converted configuration settings (from templates).
func RunDockerComposeWithConfig(
	args []string, details compose.ConfigDetails, suppressOsStdOut ...bool,
) (string, error) {
	log.Debugln()

	var tmpFiles, composeArgs []string

	log.Debugln("Reading configs...")

	for _, conf := range details.ConfigFiles {
		bs, err := yaml.Marshal(conf.Config)

		log.Traceln("Reading config:")
		log.Traceln(conf.Filename)
		log.Traceln(string(bs))

		if err != nil {
			return "", err
		}

		tmpFile, err := ioutil.TempFile(os.TempDir(), AppName+"-")
		if err != nil {
			return "", err
		}

		TmpFilesList.PushBack(tmpFile.Name())
		// defer os.Remove(tmpFile.Name())

		tmpFiles = append(tmpFiles, tmpFile.Name())

		if _, err = tmpFile.Write(bs); err != nil {
			return "", err
		}

		if err := tmpFile.Close(); err != nil {
			return "", err
		}
	}

	for _, file := range tmpFiles {
		composeArgs = append(composeArgs, "-f")
		composeArgs = append(composeArgs, file)
	}

	composeArgs = append(composeArgs, args...)

	out, err := RunDockerComposeCommand(composeArgs, suppressOsStdOut...)
	if err != nil {
		return out, err
	}

	return out, nil
}

// GenerateMutagenTemplateFileIfNotExist generates mutagen configuration from template if it doesn't exists.
func GenerateMutagenTemplateFileIfNotExist() error {
	log.Debugln()

	if CheckFileExists(GetMutagenSyncFile()) {
		// Use Local File
		return nil
	}

	var bs bytes.Buffer

	mutagenTemplate := new(template.Template)
	mutagenTemplateList := list.New()

	err := AppendMutagenTemplates(mutagenTemplate, mutagenTemplateList, "mutagen")
	if err != nil {
		return err
	}

	for e := mutagenTemplateList.Front(); e != nil; e = e.Next() {
		tplName := fmt.Sprint(e.Value)

		err := ExecuteTemplate(mutagenTemplate.Lookup(tplName), &bs)
		if err != nil {
			return err
		}
	}

	err = CreateDirAndWriteBytesToFile(bs.Bytes(), GetMutagenSyncFile(), 0o640)

	return err
}

// Cleanup removes all the temporary template files.
func Cleanup() error {
	log.Debugln()

	for e := TmpFilesList.Front(); e != nil; e = e.Next() {
		log.Traceln("Cleanup:", e.Value)

		err := os.Remove(fmt.Sprint(e.Value))
		if err != nil {
			return err
		}
	}

	return nil
}

// isEnabledPermissive returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string).
//   Also returns true if the given value is unset. (Permissive)
func isEnabledPermissive(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	switch g.Kind() { //nolint:exhaustive
	case reflect.String:
		return g.String() == "true" || g.String() == "1"
	case reflect.Bool:
		return g.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 1
	default:
		return false
	}
}

// isEnabledStrict returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string).
//   It returns false if the given value is unset. (Strict)
func isEnabledStrict(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return false
	}

	switch g.Kind() { //nolint:exhaustive
	case reflect.String:
		return g.String() == "true" || g.String() == "1"
	case reflect.Bool:
		return g.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 1
	default:
		return false
	}
}

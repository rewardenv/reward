package internal

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

	"gopkg.in/yaml.v3"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"

	"github.com/Masterminds/sprig"
	"github.com/docker/cli/cli/compose/loader"
	compose "github.com/docker/cli/cli/compose/types"
)

var tmpFilesList = list.New()

func AppendTemplatesFromPaths(t *template.Template, templateList *list.List, paths []string) error {
	funcs := make(map[string]interface{})
	funcs["isEnabled"] = isEnabled

	for _, path := range paths {
		templatePath := path
		filePath := filepath.Join(GetCwd(), path)

		// Check for template in CWD
		if CheckFileExists(filePath) {
			log.Traceln("appending template from $CWD:", templatePath)

			searchT := t.Lookup(templatePath)
			if searchT != nil {
				log.Traceln("template already defined:", templatePath)
				continue
			}

			child, err := template.New(templatePath).Funcs(sprig.TxtFuncMap()).Funcs(funcs).ParseFiles(filePath)
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

			child, err := template.New(templatePath).Funcs(sprig.TxtFuncMap()).Funcs(funcs).ParseFiles(filePath)
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

func AppendTemplatesFromPathsStatic(t *template.Template, templateList *list.List, paths []string) error {
	funcs := make(map[string]interface{})
	funcs["isEnabled"] = isEnabled

	log.Traceln(paths)

	for _, path := range paths {
		templatePath := filepath.Join(path)

		searchT := t.Lookup(templatePath)
		if searchT != nil {
			log.Traceln("template already defined:", templatePath)
			continue
		}

		content, err := Asset(templatePath)
		if err != nil {
			log.Traceln(err)
		} else {
			log.Traceln("appending template:", templatePath)
			child, err := template.New(templatePath).Funcs(sprig.TxtFuncMap()).Funcs(funcs).Parse(string(content))
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

func AppendEnvironmentTemplates(t *template.Template, templateList *list.List, partialName string) error {
	envType := GetEnvType()
	staticTemplatePaths := []string{
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
	}
	templatePaths := []string{
		filepath.Join(
			"templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join(
			"templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
		filepath.Join(
			"templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName)),
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

	// for _, v := range staticTemplatePaths {
	// 	content, err := Asset(v)
	// 	if err != nil {
	// 		log.Debugln(err)
	// 	} else {
	// 		log.Debugln("appending template:", v)
	// 		child := template.Must(template.New(v).Funcs(sprig.TxtFuncMap()).Parse(string(content)))
	// 		_, _ = t.AddParseTree(child.Name(), child.Tree)
	// 		templateList.PushBack(child.Name())
	// 	}
	// }

	// for _, v := range templatePaths {
	// 	if CheckFileExists(v) {
	// 		log.Debugln("appending template:", v)
	//
	// 		child := template.Must(template.New(v).Funcs(sprig.TxtFuncMap()).ParseFiles(v))
	// 		_, _ = t.AddParseTree(child.Name(), child.Tree)
	// 		t.Option()
	// 	} else {
	// 		log.Debugln("template not found:", v)
	// 	}
	// }

	return nil
}

func AppendMutagenTemplates(t *template.Template, templateList *list.List, partialName string) error {
	envType := GetEnvType()
	staticTemplatePaths := []string{
		filepath.Join("templates/environments", envType, fmt.Sprintf("%v.%v.yml", envType, partialName)),
		filepath.Join("templates/environments", envType, fmt.Sprintf("%v.%v.%v.yml", envType, partialName, runtime.GOOS)),
	}

	log.Traceln("template paths:")
	log.Traceln(staticTemplatePaths, staticTemplatePaths)

	for _, v := range staticTemplatePaths {
		content, err := Asset(v)
		if err != nil {
			log.Traceln(err)
		} else {
			log.Traceln("appending template:", v)
			child, err := template.New(v).Funcs(sprig.TxtFuncMap()).Parse(string(content))
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

func ExecuteTemplate(t *template.Template, buffer io.Writer) error {
	funcs := make(map[string]interface{})
	funcs["isEnabled"] = isEnabled

	log.Debugln("Executing template:", t.Name())
	log.Debugln(viper.AllSettings())
	log.Debugln(t.DefinedTemplates())

	err := t.Funcs(sprig.TxtFuncMap()).Funcs(funcs).ExecuteTemplate(buffer, t.Name(), viper.AllSettings())

	return err
}

func ConvertTemplateToComposeConfig(t *template.Template, templateList *list.List) (compose.ConfigDetails, error) {
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

func RunDockerComposeWithConfig(
	args []string, details compose.ConfigDetails, suppressOsStdOut ...bool) (string, error) {
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

		tmpFilesList.PushBack(tmpFile.Name())
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

func GenerateMutagenTemplateFileIfNotExist() error {
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

func Cleanup() error {
	for e := tmpFilesList.Front(); e != nil; e = e.Next() {
		log.Traceln("Cleanup:", e.Value)

		err := os.Remove(fmt.Sprint(e.Value))
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string).
func isEnabled(given interface{}) bool {
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

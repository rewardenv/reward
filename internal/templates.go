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
		templatePath := filepath.Join(path)
		if CheckFileExists(templatePath) {
			log.Traceln("appending template:", templatePath)

			child := template.Must(template.New(templatePath).Funcs(sprig.TxtFuncMap()).Funcs(funcs).ParseFiles(templatePath))
			_, _ = t.AddParseTree(child.Name(), child.Tree)
			templateList.PushBack(child.Name())
		} else {
			log.Traceln("template not found:", templatePath)
		}
	}

	return nil
}

func AppendTemplatesFromPathsStatic(t *template.Template, templateList *list.List, paths []string) error {
	funcs := make(map[string]interface{})
	funcs["isEnabled"] = isEnabled

	log.Traceln(paths)

	for _, v := range paths {
		content, err := Asset(v)
		if err != nil {
			log.Traceln(err)
		} else {
			log.Traceln("appending template:", v)
			child := template.Must(template.New(v).Funcs(sprig.TxtFuncMap()).Funcs(funcs).Parse(string(content)))
			_, _ = t.AddParseTree(child.Name(), child.Tree)
			templateList.PushBack(child.Name())
		}
	}

	return nil
}

func AppendEnvironmentTemplates(t *template.Template, templateList *list.List, partialName string) error {
	envType := GetEnvType()
	appHomeDir := viper.GetString("app_dir")
	staticTemplatePaths := []string{
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join("templates", "environments", envType, fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
	}
	templatePaths := []string{
		filepath.Join(appHomeDir, "templates", "environments", "includes", fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join(appHomeDir, "templates", "environments", "includes", fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
		filepath.Join(appHomeDir, "templates", "environments", envType, fmt.Sprintf("%v.base.yml", partialName)),
		filepath.Join(appHomeDir, "templates", "environments", envType, fmt.Sprintf("%v.%v.yml", partialName, runtime.GOOS)),
	}

	log.Traceln("template paths:")
	log.Traceln(staticTemplatePaths, templatePaths)

	err := AppendTemplatesFromPathsStatic(t, templateList, staticTemplatePaths)
	if err != nil {
		return err
	}

	err = AppendTemplatesFromPaths(t, templateList, templatePaths)
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
			child := template.Must(template.New(v).Funcs(sprig.TxtFuncMap()).Parse(string(content)))
			_, _ = t.AddParseTree(child.Name(), child.Tree)
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

// Returns true if given value is true (bool), 1 (int), "1" (string) or "true" (string)
func isEnabled(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	switch g.Kind() {
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

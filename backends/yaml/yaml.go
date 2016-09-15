package yaml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/qor/i18n"
	"gopkg.in/yaml.v2"
)

// New new YAML backend for I18n
func New(paths ...string) i18n.Backend {
	backend := &Backend{}

	var files []string
	for _, p := range paths {
		if file, err := os.Open(p); err == nil {
			defer file.Close()
			if fileInfo, err := file.Stat(); err == nil {
				if fileInfo.IsDir() {
					yamlFiles, _ := filepath.Glob(path.Join(p, "*.yaml"))
					files = append(files, yamlFiles...)

					ymlFiles, _ := filepath.Glob(path.Join(p, "*.yml"))
					files = append(files, ymlFiles...)
				} else if fileInfo.Mode().IsRegular() {
					files = append(files, p)
				}
			}
		}
	}
	for _, file := range files {
		if content, err := ioutil.ReadFile(file); err == nil {
			backend.contents = append(backend.contents, content)
		}
	}
	return backend
}

// NewWithWalk has the same functionality as New but uses filepath.Walk to find all the translation files recursively.
func NewWithWalk(paths ...string) i18n.Backend {
	backend := &Backend{}

	var files []string
	for _, p := range paths {
		filepath.Walk(p, func(path string, fileInfo os.FileInfo, err error) error {
			if isYamlFile(fileInfo) {
				files = append(files, path)
			}
			return nil
		})
	}
	for _, file := range files {
		if content, err := ioutil.ReadFile(file); err == nil {
			backend.contents = append(backend.contents, content)
		}
	}

	return backend
}

func isYamlFile(fileInfo os.FileInfo) bool {
	if fileInfo == nil {
		return false
	}
	return fileInfo.Mode().IsRegular() && (strings.HasSuffix(fileInfo.Name(), ".yml") || strings.HasSuffix(fileInfo.Name(), ".yaml"))
}

// Backend YAML backend
type Backend struct {
	contents [][]byte
}

func loadTranslationsFromYaml(locale string, value interface{}, scopes []string) (translations []*i18n.Translation) {
	switch v := value.(type) {
	case yaml.MapSlice:
		for _, s := range v {
			results := loadTranslationsFromYaml(locale, s.Value, append(scopes, fmt.Sprintf("%v", s.Key)))
			translations = append(translations, results...)
		}
	default:
		var translation = &i18n.Translation{
			Locale: locale,
			Key:    strings.Join(scopes, "."),
			Value:  fmt.Sprintf("%v", v),
		}
		translations = append(translations, translation)
	}
	return
}

// LoadTranslations load translations from YAML backend
func (backend *Backend) LoadTranslations() (translations []*i18n.Translation) {
	for _, content := range backend.contents {
		var slice yaml.MapSlice
		if err := yaml.Unmarshal(content, &slice); err == nil {
			for _, item := range slice {
				translations = append(translations, loadTranslationsFromYaml(item.Key.(string) /* locale */, item.Value, []string{})...)
			}
		} else {
			panic(err)
		}
	}
	return translations
}

// SaveTranslation save translation into YAML backend, not implemented
func (backend *Backend) SaveTranslation(t *i18n.Translation) error {
	return errors.New("not implemented")
}

// DeleteTranslation delete translation into YAML backend, not implemented
func (backend *Backend) DeleteTranslation(t *i18n.Translation) error {
	return errors.New("not implemented")
}

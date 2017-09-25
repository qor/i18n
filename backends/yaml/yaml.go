package yaml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qor/i18n"
	"gopkg.in/yaml.v2"
)

var _ i18n.Backend = &Backend{}

// New new YAML backend for I18n
func New(paths ...string) *Backend {
	backend := &Backend{}

	for _, p := range paths {
		if file, err := os.Open(p); err == nil {
			defer file.Close()
			if fileInfo, err := file.Stat(); err == nil {
				if fileInfo.IsDir() {
					yamlFiles, _ := filepath.Glob(filepath.Join(p, "*.yaml"))
					backend.files = append(backend.files, yamlFiles...)

					ymlFiles, _ := filepath.Glob(filepath.Join(p, "*.yml"))
					backend.files = append(backend.files, ymlFiles...)
				} else if fileInfo.Mode().IsRegular() {
					backend.files = append(backend.files, p)
				}
			}
		}
	}

	return backend
}

// Backend YAML backend
type Backend struct {
	files []string
}

func loadTranslationsFromYaml(locale string, value interface{}, scopes []string) (translations []*i18n.Translation) {
	switch v := value.(type) {
	case yaml.MapSlice:
		for _, s := range v {
			results := loadTranslationsFromYaml(locale, s.Value, append(scopes, fmt.Sprint(s.Key)))
			translations = append(translations, results...)
		}
	default:
		var translation = &i18n.Translation{
			Locale: locale,
			Key:    strings.Join(scopes, "."),
			Value:  fmt.Sprint(v),
		}
		translations = append(translations, translation)
	}
	return
}

// LoadYAMLContent load YAML content
func (backend *Backend) LoadYAMLContent(content []byte) (translations []*i18n.Translation, err error) {
	var slice yaml.MapSlice

	if err = yaml.Unmarshal(content, &slice); err == nil {
		for _, item := range slice {
			translations = append(translations, loadTranslationsFromYaml(item.Key.(string) /* locale */, item.Value, []string{})...)
		}
	}

	return translations, err
}

// LoadTranslations load translations from YAML backend
func (backend *Backend) LoadTranslations() (translations []*i18n.Translation) {
	for _, file := range backend.files {
		if content, err := ioutil.ReadFile(file); err == nil {
			if results, err := backend.LoadYAMLContent(content); err == nil {
				translations = append(translations, results...)
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

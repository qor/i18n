package yaml_test

import (
	"fmt"
	"testing"

	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

var values = map[string][][]string{
	"en": {
		{"hello", "Hello"},
		{"user.name", "User Name"},
		{"user.email", "Email"},
	},
	"de": {
		{"hello", "Hallo"},
		{"user.name", "Benutzername"},
		{"user.email", "E-Mail-Adresse"},
	},
	"zh-CN": {
		{"hello", "你好"},
		{"user.name", "用户名"},
		{"user.email", "邮箱"},
	},
}

func checkTranslations(translations []*i18n.Translation) error {
	for locale, results := range values {
		for _, result := range results {
			var found bool
			for _, translation := range translations {
				if (translation.Locale == locale) && (translation.Key == result[0]) && (translation.Value == result[1]) {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("failed to found translation %v for %v", result[0], locale)
			}
		}
	}
	return nil
}

func TestLoadTranslations(t *testing.T) {
	backend := yaml.New("tests", "tests/subdir")
	if err := checkTranslations(backend.LoadTranslations()); err != nil {
		t.Fatal(err)
	}
}

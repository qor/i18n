package exchange_actions

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/qor/i18n"
	"github.com/qor/media_library"
	"github.com/qor/qor/admin"
	"github.com/qor/worker"
)

func RegisterExchangeJobs(I18n *i18n.I18n, Worker *worker.Worker) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/i18n/exchange_actions/views"))
	}

	Worker.RegisterJob(worker.Job{
		Name:  "Export Translations",
		Group: "Translations",
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) (err error) {
			var (
				locales         []string
				translationKeys []string
				translationsMap = map[string]bool{}
				filename        = fmt.Sprintf("/downloads/translations.%v.csv", time.Now().UnixNano())
				fullFilename    = path.Join("public", filename)
			)
			qorJob.AddLog("Exporting translations...")

			// Sort locales
			for locale, _ := range I18n.Translations {
				locales = append(locales, locale)
			}
			sort.Strings(locales)

			// Create download file
			if _, err = os.Stat(filepath.Dir(fullFilename)); os.IsNotExist(err) {
				err = os.MkdirAll(filepath.Dir(fullFilename), os.ModePerm)
			}
			csvfile, err := os.OpenFile(fullFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			defer csvfile.Close()
			if err != nil {
				return err
			}

			writer := csv.NewWriter(csvfile)

			// Append Headers
			writer.Write(append([]string{"Translation Keys"}, locales...))

			// Sort translation keys
			for _, locale := range locales {
				for key, _ := range I18n.Translations[locale] {
					translationsMap[key] = true
				}
			}

			for key := range translationsMap {
				translationKeys = append(translationKeys, key)
			}
			sort.Strings(translationKeys)

			// Write CSV file
			for _, translationKey := range translationKeys {
				var translations = []string{translationKey}
				for _, locale := range locales {
					var value string
					if translation := I18n.Translations[locale][translationKey]; translation != nil {
						value = translation.Value
					}
					translations = append(translations, value)
				}
				writer.Write(translations)
			}
			writer.Flush()

			qorJob.SetProgressText(fmt.Sprintf("<a href='%v'>Download exported translations</a>", filename))
			return
		},
	})

	// Import Translations
	type importTranslationArgument struct {
		Translations media_library.FileSystem
	}

	Worker.RegisterJob(worker.Job{
		Name:     "Import Translations",
		Group:    "Translations",
		Resource: Worker.Admin.NewResource(&importTranslationArgument{}),
		Handler: func(arg interface{}, qorJob worker.QorJobInterface) (err error) {
			importTranslationArgument := arg.(*importTranslationArgument)
			qorJob.AddLog("Importing translations...")
			if csvfile, err := os.Open(path.Join("public", importTranslationArgument.Translations.URL())); err == nil {
				reader := csv.NewReader(csvfile)
				reader.TrimLeadingSpace = true
				if records, err := reader.ReadAll(); err == nil {
					if len(records) > 1 && len(records[0]) > 1 {
						locales := records[0][1:]

						for _, values := range records[1:] {
							for idx, value := range values[1:] {
								if value == "" {
									I18n.DeleteTranslation(&i18n.Translation{
										Key:    values[0],
										Locale: locales[idx],
									})
								} else {
									I18n.SaveTranslation(&i18n.Translation{
										Key:    values[0],
										Locale: locales[idx],
										Value:  value,
									})
								}
							}
						}
					}
				}
				qorJob.AddLog("Imported translations")
			}
			return
		},
	})
}

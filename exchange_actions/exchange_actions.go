package exchange_actions

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/qor/i18n"
	"github.com/qor/worker"
)

func RegisterExchangeJobs(I18n *i18n.I18n, Worker *worker.Worker) {
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
			return nil
		},
	})
}

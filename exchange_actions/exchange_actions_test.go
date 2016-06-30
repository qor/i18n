package exchange_actions_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/admin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/database"
	"github.com/qor/i18n/exchange_actions"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"
	"github.com/qor/worker"
	"io/ioutil"
	"os"
	"testing"
)

var db *gorm.DB

func init() {
	db = utils.TestDB()
	db.DropTable(&database.Translation{})
	database.New(db)
}

func TestExportScopedTranslations(t *testing.T) {
	Admin := admin.New(&qor.Config{DB: db})
	Worker := worker.New()
	Admin.AddResource(Worker)
	I18n := i18n.New(database.New(db))
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.title", Value: "title", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.subtitle", Value: "subtitle", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.description", Value: "description", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "header.title", Value: "Header Title", Locale: "en-US"})
	exchange_actions.RegisterExchangeJobs(I18n, Worker)
	clearDownloadDir()
	color.Green(fmt.Sprintf("Exchange TestCase #%d: Success\n", 1))
	for _, job := range Worker.Jobs {
		if job.Name == "Export Translations" {
			job.Handler(&exchange_actions.ExportTranslationArgument{Scope: "All"}, job.NewStruct().(worker.QorJobInterface))
			if downloadedFileContent() != loadFixture("export_translations.csv") {
				t.Errorf(color.RedString(fmt.Sprintf("\nExchange TestCase #%d: Failure (%s)\n", 1, "downloaded file should match file export_translations.csv")))
			}
		}
	}
}

func clearDownloadDir() {
	files, _ := ioutil.ReadDir("./public/downloads")
	for _, f := range files {
		os.Remove("./public/downloads/" + f.Name())
	}
}

func downloadedFileContent() string {
	files, _ := ioutil.ReadDir("./public/downloads")
	for _, f := range files {
		if content, err := ioutil.ReadFile("./public/downloads/" + f.Name()); err == nil {
			return string(content)
		}
	}
	return ""
}

func loadFixture(fileName string) string {
	if content, err := ioutil.ReadFile("./fixtures/" + fileName); err == nil {
		return string(content)
	}
	return ""
}

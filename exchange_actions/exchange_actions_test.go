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
	"github.com/qor/media_library"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"
	"github.com/qor/worker"
	"io/ioutil"
	"os"
	"testing"
)

var db *gorm.DB
var Worker *worker.Worker
var I18n *i18n.I18n

func init() {
	db = utils.TestDB()
	db.DropTable(&database.Translation{})
	database.New(db)

	Admin := admin.New(&qor.Config{DB: db})
	Worker = worker.New()
	Admin.AddResource(Worker)
	I18n = i18n.New(database.New(db))
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.title", Value: "title", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.subtitle", Value: "subtitle", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "qor_admin.description", Value: "description", Locale: "en-US"})
	I18n.SaveTranslation(&i18n.Translation{Key: "header.title", Value: "Header Title", Locale: "en-US"})
	exchange_actions.RegisterExchangeJobs(I18n, Worker)
}

type testExportWithScopedCase struct {
	Scope            string
	ExpectExportFile string
}

func TestExportTranslations(t *testing.T) {
	testCases := []*testExportWithScopedCase{
		&testExportWithScopedCase{Scope: "", ExpectExportFile: "export_all.csv"},
		&testExportWithScopedCase{Scope: "All", ExpectExportFile: "export_all.csv"},
		&testExportWithScopedCase{Scope: "Backend", ExpectExportFile: "export_backend.csv"},
		&testExportWithScopedCase{Scope: "Frontend", ExpectExportFile: "export_frontend.csv"},
	}

	for i, testcase := range testCases {
		clearDownloadDir()
		for _, job := range Worker.Jobs {
			if job.Name == "Export Translations" {
				job.Handler(&exchange_actions.ExportTranslationArgument{Scope: testcase.Scope}, job.NewStruct().(worker.QorJobInterface))
				if downloadedFileContent() != loadFixture(testcase.ExpectExportFile) {
					t.Errorf(color.RedString(fmt.Sprintf("\nExchange TestCase #%d: Failure (%s)\n", i+1, "downloaded file should match file export_translations.csv")))
				} else {
					color.Green(fmt.Sprintf("Export with scope TestCase #%d: Success\n", i+1))
				}
			}
		}
	}
}

func TestImportTranslations(t *testing.T) {
	color.Green(fmt.Sprintf("Import TestCase #%d: Success\n", 1))
	for _, job := range Worker.Jobs {
		if job.Name == "Import Translations" {
			job.Handler(&exchange_actions.ImportTranslationArgument{TranslationsFile: media_library.FileSystem{media_library.Base{Url: "imports/import_translations.csv"}}}, job.NewStruct().(worker.QorJobInterface))
			zhMapping := map[string]string{"qor_admin.title": "标题", "qor_admin.subtitle": "小标题", "qor_admin.description": "描述", "header.title": "标题"}
			for key, translation := range I18n.LoadTranslations()["zh-CN"] {
				if zhMapping[key] != translation.Value {
					t.Errorf(color.RedString(fmt.Sprintf("\nExchange TestCase #%d: Failure (%s)\n", 1, "zh translations not match")))
				}
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

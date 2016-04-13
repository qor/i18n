package i18n

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

// Default default locale for i18n
var Default = "en-US"

// I18n struct that hold all translations
type I18n struct {
	scope        string
	value        string
	isInlineEdit bool
	Backends     []Backend
	Translations map[string]map[string]*Translation

	mutex *sync.RWMutex
}

// ResourceName change display name in qor admin
func (I18n) ResourceName() string {
	return "Translation"
}

// Backend defined methods that needs for translation backend
type Backend interface {
	LoadTranslations() []*Translation
	SaveTranslation(*Translation) error
	DeleteTranslation(*Translation) error
}

// Translation is a struct for translations, including Translation Key, Locale, Value
type Translation struct {
	Key     string
	Locale  string
	Value   string
	Backend Backend
}

// New initialize I18n with backends
func New(backends ...Backend) *I18n {
	var mutex sync.RWMutex
	i18n := &I18n{
		Backends:     backends,
		Translations: map[string]map[string]*Translation{},
		mutex:        &mutex,
	}
	for i := len(backends) - 1; i >= 0; i-- {
		var backend = backends[i]
		for _, translation := range backend.LoadTranslations() {
			translation.Backend = backend
			i18n.AddTranslation(translation)
		}
	}
	return i18n
}

// AddTranslation add translation
func (i18n *I18n) AddTranslation(translation *Translation) {
	i18n.mutex.Lock()
	defer i18n.mutex.Unlock()
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]*Translation{}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
}

// SaveTranslation save translation
func (i18n *I18n) SaveTranslation(translation *Translation) error {
	var backends []Backend
	if backend := translation.Backend; backend != nil {
		backends = append(backends, backend)
	}

	for _, backend := range i18n.Backends {
		if backend.SaveTranslation(translation) == nil {
			i18n.AddTranslation(translation)
			return nil
		}
	}

	return errors.New("failed to save translation")
}

// DeleteTranslation delete translation
func (i18n *I18n) DeleteTranslation(translation *Translation) (err error) {
	if translation.Backend == nil {
		i18n.mutex.RLock()
		if ts := i18n.Translations[translation.Locale]; ts != nil && ts[translation.Key] != nil {
			translation = ts[translation.Key]
		}
		i18n.mutex.RUnlock()
	}

	if translation.Backend != nil {
		if err = translation.Backend.DeleteTranslation(translation); err == nil {
			i18n.mutex.Lock()
			delete(i18n.Translations[translation.Locale], translation.Key)
			i18n.mutex.Unlock()
		}
	}
	return err
}

// EnableInlineEdit enable inline edit, return HTML used to edit the translation
func (i18n *I18n) EnableInlineEdit(isInlineEdit bool) admin.I18n {
	return &I18n{
		Translations: i18n.Translations,
		mutex:        i18n.mutex,
		scope:        i18n.scope,
		value:        i18n.value,
		Backends:     i18n.Backends,
		isInlineEdit: isInlineEdit,
	}
}

// Scope i18n scope
func (i18n *I18n) Scope(scope string) admin.I18n {
	return &I18n{
		Translations: i18n.Translations,
		mutex:        i18n.mutex,
		scope:        scope,
		value:        i18n.value,
		Backends:     i18n.Backends,
		isInlineEdit: i18n.isInlineEdit,
	}
}

// Default default value of translation if key is missing
func (i18n *I18n) Default(value string) admin.I18n {
	return &I18n{
		Translations: i18n.Translations,
		mutex:        i18n.mutex,
		scope:        i18n.scope,
		value:        value,
		Backends:     i18n.Backends,
		isInlineEdit: i18n.isInlineEdit,
	}
}

// T translate with locale, key and arguments
func (i18n *I18n) T(locale, key string, args ...interface{}) template.HTML {
	var value = i18n.value
	var translationKey = key
	if i18n.scope != "" {
		translationKey = strings.Join([]string{i18n.scope, key}, ".")
	}

	i18n.mutex.RLock()
	if translations := i18n.Translations[locale]; translations != nil && translations[translationKey] != nil && translations[translationKey].Value != "" {
		// Get localized translation
		value = translations[translationKey].Value
		i18n.mutex.RUnlock()
	} else if translations := i18n.Translations[Default]; translations != nil && translations[translationKey] != nil {
		// Get default translation if not translated
		value = translations[translationKey].Value
		i18n.mutex.RUnlock()
	} else {
		i18n.mutex.RUnlock()
		if value == "" {
			value = key
		}
		// Save translations
		i18n.SaveTranslation(&Translation{Key: translationKey, Value: value, Locale: Default, Backend: i18n.Backends[0]})
	}

	if value == "" {
		value = key
	}

	if str, err := cldr.Parse(locale, value, args...); err == nil {
		value = str
	}

	if i18n.isInlineEdit {
		var editType string
		if len(value) > 25 {
			editType = "data-type=\"textarea\""
		}
		value = fmt.Sprintf("<span class=\"qor-i18n-inline\" %s data-locale=\"%s\" data-key=\"%s\">%s</span>", editType, locale, key, value)
	}

	return template.HTML(value)
}

// RenderInlineEditAssets render inline edit html, it is using: http://vitalets.github.io/x-editable/index.html
// You could use Bootstrap or JQuery UI by set isIncludeExtendAssetLib to false and load files by yourself
func RenderInlineEditAssets(isIncludeJQuery bool, isIncludeExtendAssetLib bool) (template.HTML, error) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		var content string
		var hasError bool

		if isIncludeJQuery {
			content = `<script src="http://code.jquery.com/jquery-2.0.3.min.js"></script>`
		}

		if isIncludeExtendAssetLib {
			if extendLib, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/i18n/views/themes/i18n/inline-edit-libs.tmpl")); err == nil {
				content += string(extendLib)
			} else {
				hasError = true
			}

			if css, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/i18n/views/themes/i18n/assets/stylesheets/i18n-inline.css")); err == nil {
				content += fmt.Sprintf("<style>%s</style>", string(css))
			} else {
				hasError = true
			}

		}

		if js, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/i18n/views/themes/i18n/assets/javascripts/i18n-inline.js")); err == nil {
			content += fmt.Sprintf("<script type=\"text/javascript\">%s</script>", string(js))
		} else {
			hasError = true
		}

		if !hasError {
			return template.HTML(content), nil
		}
	}

	return template.HTML(""), errors.New("templates not found")
}

func getLocaleFromContext(context *qor.Context) string {
	if locale := utils.GetLocale(context); locale != "" {
		return locale
	}

	return Default
}

type availableLocalesInterface interface {
	AvailableLocales() []string
}

type viewableLocalesInterface interface {
	ViewableLocales() []string
}

type editableLocalesInterface interface {
	EditableLocales() []string
}

func getAvailableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(viewableLocalesInterface); ok {
		return user.ViewableLocales()
	}

	if user, ok := currentUser.(availableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{Default}
}

func getEditableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(editableLocalesInterface); ok {
		return user.EditableLocales()
	}

	if user, ok := currentUser.(availableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{Default}
}

// ConfigureQorResource configure qor resource for qor admin
func (i18n *I18n) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.UseTheme("i18n")
		res.GetAdmin().I18n = i18n
		res.SearchAttrs("value") // generate search handler for i18n

		res.GetAdmin().RegisterFuncMap("lt", func(locale, key string, withDefault bool) string {
			i18n.mutex.RLock()
			defer i18n.mutex.RUnlock()

			if translations := i18n.Translations[locale]; translations != nil {
				if t := translations[key]; t != nil && t.Value != "" {
					return t.Value
				}
			}

			if withDefault {
				if translations := i18n.Translations[Default]; translations != nil {
					if t := translations[key]; t != nil {
						return t.Value
					}
				}
			}

			return ""
		})

		var getPrimaryLocale = func(context *admin.Context) string {
			if locale := context.Request.Form.Get("primary_locale"); locale != "" {
				return locale
			}
			if availableLocales := getAvailableLocales(context.Request, context.CurrentUser); len(availableLocales) > 0 {
				return availableLocales[0]
			}
			return ""
		}

		var getEditingLocale = func(context *admin.Context) string {
			if locale := context.Request.Form.Get("to_locale"); locale != "" {
				return locale
			}
			return getLocaleFromContext(context.Context)
		}

		res.GetAdmin().RegisterFuncMap("i18n_available_keys", func(context *admin.Context) (keys []string) {
			var (
				keysMap       = map[string]bool{}
				keyword       = strings.ToLower(context.Request.URL.Query().Get("keyword"))
				primaryLocale = getPrimaryLocale(context)
				editingLocale = getEditingLocale(context)
			)

			var filterTranslations = func(translations map[string]*Translation) {
				if translations != nil {
					for key, translation := range translations {
						if (keyword == "") || (strings.Index(strings.ToLower(translation.Key), keyword) != -1 ||
							strings.Index(strings.ToLower(translation.Value), keyword) != -1) {
							if _, ok := keysMap[key]; !ok {
								keysMap[key] = true
								keys = append(keys, key)
							}
						}
					}
				}
			}

			i18n.mutex.RLock()
			filterTranslations(i18n.Translations[getPrimaryLocale(context)])
			if primaryLocale != editingLocale {
				filterTranslations(i18n.Translations[getEditingLocale(context)])
			}
			i18n.mutex.RUnlock()

			sort.Strings(keys)

			pagination := context.Searcher.Pagination
			pagination.Total = len(keys)
			pagination.PrePage = 25
			pagination.CurrentPage, _ = strconv.Atoi(context.Request.URL.Query().Get("page"))
			if pagination.CurrentPage == 0 {
				pagination.CurrentPage = 1
			}
			if pagination.CurrentPage > 0 {
				pagination.Pages = pagination.Total / pagination.PrePage
			}
			context.Searcher.Pagination = pagination

			if pagination.CurrentPage == -1 {
				return keys
			}

			lastIndex := pagination.CurrentPage * pagination.PrePage
			if pagination.Total < lastIndex {
				lastIndex = pagination.Total
			}

			startIndex := (pagination.CurrentPage - 1) * pagination.PrePage
			if lastIndex >= startIndex {
				return keys[startIndex:lastIndex]
			}
			return []string{}
		})

		res.GetAdmin().RegisterFuncMap("i18n_primary_locale", getPrimaryLocale)

		res.GetAdmin().RegisterFuncMap("i18n_editing_locale", getEditingLocale)

		res.GetAdmin().RegisterFuncMap("i18n_viewable_locales", func(context admin.Context) []string {
			return getAvailableLocales(context.Request, context.CurrentUser)
		})

		res.GetAdmin().RegisterFuncMap("i18n_editable_locales", func(context admin.Context) []string {
			return getEditableLocales(context.Request, context.CurrentUser)
		})

		controller := i18nController{i18n}
		router := res.GetAdmin().GetRouter()
		router.Get(res.ToParam(), controller.Index)
		router.Post(res.ToParam(), controller.Update)
		router.Put(res.ToParam(), controller.Update)

		admin.RegisterViewPath("github.com/qor/i18n/views")
	}
}

package inline_edit

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/qor/i18n"
)

func enabledInlineEdit(request *http.Request) bool {
	return true
}

func GenerateFuncMaps(I18n *i18n.I18n, request *http.Request) template.FuncMap {
	if enabledInlineEdit(request) {
		return template.FuncMap{
			"t": inlineEdit(I18n),
		}
	} else {
		return template.FuncMap{
			"t": I18n.T,
		}
	}
}

func inlineEdit(I18n *i18n.I18n) func(string, string, ...interface{}) template.HTML {
	return func(locale, key string, args ...interface{}) template.HTML {
		var editType string
		if len(value) > 25 {
			editType = "data-type=\"textarea\""
		}
		assetsTag := fmt.Sprintf("<script data-prefix=\"%v\" src=\"/admin/assets/javascripts/i18n-checker.js?theme=i18n\"></script>", i18n.Resource.GetAdmin().GetRouter().Prefix)
		value = fmt.Sprintf("%s<span class=\"qor-i18n-inline\" %s data-locale=\"%s\" data-key=\"%s\">%s</span>", assetsTag, editType, locale, key, I18n.T(locale, key, args...))
	}
}

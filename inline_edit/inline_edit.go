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

func GenerateFuncMaps(I18n *i18n.I18n, locale string, request *http.Request) template.FuncMap {
	return template.FuncMap{
		"t": inlineEdit(I18n, locale, enabledInlineEdit(request)),
	}
}

func inlineEdit(I18n *i18n.I18n, locale string, isInline bool) func(string, ...interface{}) template.HTML {
	return func(key string, args ...interface{}) template.HTML {
		if isInline {
			var editType string
			value := I18n.T(locale, key, args...)
			if len(value) > 25 {
				editType = "data-type=\"textarea\""
			}
			assetsTag := fmt.Sprintf("<script data-prefix=\"%v\" src=\"/admin/assets/javascripts/i18n-checker.js?theme=i18n\"></script>", I18n.Resource.GetAdmin().GetRouter().Prefix)
			return template.HTML(fmt.Sprintf("%s<span class=\"qor-i18n-inline\" %s data-locale=\"%s\" data-key=\"%s\">%s</span>", assetsTag, editType, locale, key, value))
		} else {
			return I18n.T(locale, key, args...)
		}
	}
}

package i18n

import (
	"fmt"

	"github.com/qor/admin"
)

type i18nController struct {
	*I18n
}

func (controller *i18nController) Index(context *admin.Context) {
	context.Execute("index", controller.I18n)
}

func (controller *i18nController) Update(context *admin.Context) {
	form := context.Request.Form
	translation := Translation{Key: form.Get("Key"), Locale: form.Get("Locale"), Value: form.Get("Value")}

	if !controller.validateLocale(&translation) {
		context.Writer.WriteHeader(400)
		fmt.Fprintf(context.Writer, "requested locale '%s' is unexpected", translation.Locale)
	}

	if err := controller.I18n.SaveTranslation(&translation); err == nil {
		context.Writer.Write([]byte("OK"))
	} else {
		context.Writer.WriteHeader(422)
		context.Writer.Write([]byte(err.Error()))
	}
}

func (controller *i18nController) validateLocale(translation *Translation) (ok bool) {
	for locale := range controller.LoadTranslations() {
		if locale == translation.Locale {
			ok = true
		}
	}
	return
}

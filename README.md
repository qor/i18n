# I18n

I18n provides internationalization support for your application, it supports different storage solutions (*backends*) including a SQL database and YAML.

[![GoDoc](https://godoc.org/github.com/qor/i18n?status.svg)](https://godoc.org/github.com/qor/i18n)

## Usage

```go
import (
  "github.com/jinzhu/gorm"
  "github.com/qor/i18n"
  "github.com/qor/i18n/backends/database"
)

func main() {
  db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")

  // Using two backends, earlier backend has higher priority
  I18n = i18n.New(
    database.New(&db), // load translations from database
    yaml.New(filepath.Join(config.Root, "config/locales")),  // load translations from YAML files in directory `config/locales
  )

  // Add Translation
  I18n.AddTranslation(&i18n.Translation{Key: "hello-world", Locale: "en-US", Value: "hello world"})

  // Update Translation
  I18n.SaveTranslation(&i18n.Translation{Key: "hello-world", Locale: "en-US", Value: "Hello World"})

  // Delete Translation
  I18n.DeleteTranslation(&i18n.Translation{Key: "hello-world", Locale: "en-US", Value: "Hello World"})

  // Read Translation with key `hello-world`
  I18n.T("en-US", "hello-world")

  // Read Translation with `Scope`
  I18n.Scope("home-page").T("zh-CN", "hello-world") // read translation with translation key `home-page.hello-world`

  // Read Translation with `Default Value`
  I18n.Default("Default Value").T("zh-CN", "non-existing-key") // Will return default value `Default Value`
}
```

### Interpolation

I18n utilises Golang template to parse translations with interpolation variable.

```go
I18n.AddTranslation(&i18n.Translation{Key: "hello", Locale: "en-US", Value: "Hello {{.Name}}"})
type User struct {
  Name string
}
I18n.T("en-US", "hello", User{Name: "jinzhu"}) //=> Hello jinzhu
```

### Pluralization

I18n utilises [cldr](https://github.com/theplant/cldr) to achieve pluralization, it provides the functions `p`, `zero`, `one`, `two`, `few`, `many`, `other` for this purpose. Refer to cldr documentation for more information.

```go
I18n.AddTranslation(&i18n.Translation{Key: "count", Locale: "en-US", Value: "{{p "Count" (one "{{.Count}} item") (other "{{.Count}} items")}}"})
I18n.T("en-US", "count", map[string]int{"Count": 1}) //=> 1 item
```

### Ordered Params

```go
I18n.AddTranslation(&i18n.Translation{Key: "ordered_params", Locale: "en-US", Value: "{{$1}} {{$2}} {{$1}}"})
I18n.T("en-US", "ordered_params", "string1", "string2") //=> string1 string2 string1
```

### Integrate with Golang Templates

You could define a `t` method and register it as FuncMap:

```go

var I18n *i18n.I18n

func init() {
  I18n = i18n.New(database.New(&db), yaml.New(filepath.Join(config.Root, "config/locales")))
}

func T(key string, value string, args ...interface{}) string {
	return I18n.Default(value).T("en-US", key, args...)
}

// then use it like
{{t "home_page.how_it_works" "HOW DOES IT WORK? {{$1}}" "It works" }}
```

## [QOR Support](https://github.com/qor/qor)

[QOR](http://getqor.com) is architected from the ground up to accelerate development and deployment of Content Management Systems, E-commerce Systems, and Business Applications and as such is comprised of modules that abstract common features for such systems.

Although I18n can be used standalone, it works very nicely with QOR - if you have requirements to manage your application's data, be sure to check QOR out!

To use I18n with QOR, simply add it as resource to the admin:

```go
 Admin.AddResource(I18n)
 ```

[QOR Demo:  http://demo.getqor.com/admin](http://demo.getqor.com/admin)

[I18n Demo with QOR](http://demo.getqor.com/admin/translations)

### Inline Edit

You could manage translations with QOR Admin interface after register it into QOR Admin, But usually hard to translate a translation without its context, the inline edit is made to resolve the problem

With it, you could manage translations from frontend, just like [Integrate with Golang Templates](#integrate_with_golang_templates), you need to register a func map for Golang templates to render inline editable translations.

The good thing is we have created a package could be used to generate the funcmap for you, you could just use it when parseing your templates

```go
// `I18n` hold translations backends
// `en-US` current locale
// `true` enable inline edit or not, it inline edit not enabled, it works just like the funcmap in section "Integrate with Golang Templates"
inline_edit.FuncMap(I18n, "en-US", true)
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).

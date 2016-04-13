package i18n

import (
	"fmt"
	"testing"
)

type backend struct{}

func (b *backend) LoadTranslations() (translations []*Translation) { return translations }
func (b *backend) SaveTranslation(t *Translation) error            { return nil }
func (b *backend) DeleteTranslation(t *Translation) error          { return nil }

const BIGNUM = 100000

// run TestConcurrent* tests with -race flag would be better

func TestConcurrentReadWrite(t *testing.T) {
	i18n := New(&backend{})
	go func() {
		for i := 0; i < BIGNUM; i++ {
			i18n.AddTranslation(&Translation{Key: fmt.Sprintf("xx-%d", i), Locale: "xx", Value: fmt.Sprint(i)})
		}
	}()
	go func() {
		for i := 0; i < BIGNUM; i++ {
			i18n.AddTranslation(&Translation{Key: fmt.Sprintf("xx-%d", i), Locale: fmt.Sprintf("xx-%d", i), Value: fmt.Sprint(i)})
		}
	}()
	go func() {
		ni18n := i18n.EnableInlineEdit(true)
		for i := 0; i < BIGNUM; i++ {
			ni18n.T(fmt.Sprintf("xx-%d", i), fmt.Sprintf("xx-%d", i))
		}
	}()
	go func() {
		ni18n := i18n.Scope("")
		for i := 0; i < BIGNUM; i++ {
			ni18n.T(fmt.Sprintf("xx-%d", i), fmt.Sprintf("xx-%d", i))
		}
	}()
	go func() {
		ni18n := i18n.Default("")
		for i := 0; i < BIGNUM; i++ {
			ni18n.T(fmt.Sprintf("xx-%d", i), fmt.Sprintf("xx-%d", i))
		}
	}()
	for i := 0; i < BIGNUM; i++ {
		i18n.T("xx", fmt.Sprintf("xx-%d", i))
	}
}

func TestConcurrentDeleteWrite(t *testing.T) {
	i18n := New(&backend{})
	go func() {
		for i := 0; i < BIGNUM; i++ {
			i18n.AddTranslation(&Translation{Key: fmt.Sprintf("xx-%d", i), Locale: "xx", Value: fmt.Sprint(i)})
		}
	}()
	for i := 0; i < BIGNUM; i++ {
		i18n.DeleteTranslation(&Translation{Key: fmt.Sprintf("xx-%d", i), Locale: "xx", Value: fmt.Sprint(i)})
	}
}

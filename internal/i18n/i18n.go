// Package i18n is a tiny translation layer for the TUI.
//
// Strings are keyed by short identifiers (e.g. "store.title") and mapped to
// the current language's string via T(). Unknown keys fall back to the English
// catalog, then to the raw key — so a missing translation just looks ugly, it
// never crashes.
//
// State package calls SetLang() on load so the player's choice persists.
// UI views call T() whenever they need a human string.
package i18n

import (
	"fmt"
	"strings"
)

const (
	LangEN = "en"
	LangZH = "zh"
)

// Languages lists supported language codes in cycle order.
var Languages = []string{LangEN, LangZH}

var current = LangEN

// SetLang switches the active language. Unknown codes fall back to "en".
func SetLang(code string) {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, c := range Languages {
		if c == code {
			current = code
			return
		}
	}
	current = LangEN
}

// Lang returns the active language code.
func Lang() string { return current }

// CycleLang advances to the next language and returns it.
func CycleLang() string {
	for i, c := range Languages {
		if c == current {
			current = Languages[(i+1)%len(Languages)]
			return current
		}
	}
	current = LangEN
	return current
}

// Label returns the human-readable name of a language code.
func Label(code string) string {
	switch code {
	case LangEN:
		return "English"
	case LangZH:
		return "中文"
	}
	return code
}

// T returns the translation for key under the active language. Any args are
// forwarded to fmt.Sprintf when the translated string contains format verbs.
func T(key string, args ...any) string {
	cat := catalogs[current]
	if s, ok := cat[key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(s, args...)
		}
		return s
	}
	// Fallback to English.
	if cat2, ok := catalogs[LangEN]; ok {
		if s, ok := cat2[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(s, args...)
			}
			return s
		}
	}
	// Last resort — return the key, plus any args rendered.
	if len(args) > 0 {
		return fmt.Sprintf("%s %v", key, args)
	}
	return key
}

// Pick returns the first non-empty localized string when a translation exists
// inline on a data struct (e.g. GPU name in JSON).
func Pick(english, chinese string) string {
	if current == LangZH && chinese != "" {
		return chinese
	}
	return english
}

// catalogs maps language -> key -> text. Populated by en.go / zh.go init().
var catalogs = map[string]map[string]string{
	LangEN: {},
	LangZH: {},
}

// Register merges the given map into the named catalog. Called by
// language-specific files at init time.
func Register(lang string, m map[string]string) {
	if catalogs[lang] == nil {
		catalogs[lang] = map[string]string{}
	}
	for k, v := range m {
		catalogs[lang][k] = v
	}
}

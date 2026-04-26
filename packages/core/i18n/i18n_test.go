package i18n

import (
	"strings"
	"testing"
)

func TestLangCycleVisitsAll(t *testing.T) {
	SetLang(LangEN)
	seen := map[string]bool{Lang(): true}
	for i := 0; i < len(Languages)*2; i++ {
		seen[CycleLang()] = true
	}
	for _, code := range Languages {
		if !seen[code] {
			t.Errorf("cycle never visited %q", code)
		}
	}
	SetLang(LangEN)
}

func TestSetLangFallbackOnBadInput(t *testing.T) {
	SetLang("klingon")
	if Lang() != LangEN {
		t.Errorf("expected fallback to EN, got %q", Lang())
	}
}

// TestAllKeysTranslated asserts that every key in the English catalog also
// exists in the Chinese catalog. Missing translations silently fall back to
// English at runtime — this test makes them fail CI instead.
func TestAllKeysTranslated(t *testing.T) {
	en := catalogs[LangEN]
	zh := catalogs[LangZH]
	missing := []string{}
	for key := range en {
		if _, ok := zh[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		t.Errorf("zh catalog is missing %d keys: %s", len(missing), strings.Join(missing, ", "))
	}
}

func TestPickFallsBackWhenEmpty(t *testing.T) {
	SetLang(LangZH)
	defer SetLang(LangEN)
	if got := Pick("english", ""); got != "english" {
		t.Errorf("Pick should fall back to english when zh is empty, got %q", got)
	}
	if got := Pick("english", "中文"); got != "中文" {
		t.Errorf("Pick should use zh when available, got %q", got)
	}
}

func TestTFormatsArgs(t *testing.T) {
	SetLang(LangEN)
	got := T("hdr.tp", 42)
	if !strings.Contains(got, "42") {
		t.Errorf("T should interpolate args: got %q", got)
	}
}

func TestTUnknownKeyDoesNotCrash(t *testing.T) {
	got := T("no.such.key")
	if got == "" {
		t.Error("unknown key should not return empty string")
	}
}

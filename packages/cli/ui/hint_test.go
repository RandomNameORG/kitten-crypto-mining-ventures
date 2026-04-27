package ui

import (
	"strings"
	"testing"
)

func TestRenderViewHintMainViewsNonEmpty(t *testing.T) {
	mains := []viewID{
		viewDashboard, viewStore, viewGPUs, viewRooms,
		viewSkills, viewMercs, viewLab, viewPrestige, viewStats,
	}
	for _, v := range mains {
		a := App{view: v}
		if got := a.renderViewHint(); got == "" {
			t.Errorf("renderViewHint for view %d returned empty string, expected a hint", v)
		}
	}
}

func TestRenderViewHintLogAndHelpEmpty(t *testing.T) {
	for _, v := range []viewID{viewLog, viewHelp} {
		a := App{view: v}
		if got := a.renderViewHint(); got != "" {
			t.Errorf("renderViewHint for view %d returned %q, expected empty", v, got)
		}
	}
}

func TestRenderViewHintTokens(t *testing.T) {
	cases := []struct {
		view   viewID
		tokens []string
	}{
		{viewGPUs, []string{"[o]", "[b]"}},
		{viewPrestige, []string{"[R", "[y]", "[n]"}},
		{viewMercs, []string{"[tab]", "[h]", "[f]"}},
		{viewLab, []string{"[t]", "[r]", "[p]"}},
		{viewStats, []string{"[esc]"}},
		{viewDashboard, []string{"[p]", "[V]", "[space]"}},
		{viewStore, []string{"[b]", "[esc]"}},
		{viewRooms, []string{"[u]", "[enter]"}},
		{viewSkills, []string{"[u]", "[esc]"}},
	}
	for _, tc := range cases {
		a := App{view: tc.view}
		got := a.renderViewHint()
		for _, tok := range tc.tokens {
			if !strings.Contains(got, tok) {
				t.Errorf("hint for view %d = %q; missing token %q", tc.view, got, tok)
			}
		}
	}
}

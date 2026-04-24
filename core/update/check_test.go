package update

import (
	"strings"
	"testing"
)

// TestIsNewer_BasicOrdering anchors the critical happy path — a fresh
// patch release must be detected. If this breaks, players on old builds
// would miss prompts entirely, which is the whole feature's reason for
// existing.
func TestIsNewer_BasicOrdering(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
		name            string
	}{
		{"v1.0.0", "v1.0.1", true, "patch bump"},
		{"v1.2.3", "v1.10.0", true, "double-digit minor is not lexicographic"},
		{"v1.0", "v1.0.0", false, "padded equal"},
		{"v1.0.0", "v1.0.0", false, "identical"},
		{"v1.3.0", "v1.2.9", false, "older not newer"},
		{"1.0.0", "1.0.1", true, "no v prefix"},
		{"v0.1.0", "v1.0.0", true, "major bump"},
		{"v2.0.0", "v1.9.9", false, "major regression not newer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsNewer(tc.current, tc.latest)
			if got != tc.want {
				t.Fatalf("IsNewer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
			}
		})
	}
}

// TestIsNewer_FailClosed — malformed inputs MUST return false so a broken
// tag on GitHub never triggers a noisy prompt loop.
func TestIsNewer_FailClosed(t *testing.T) {
	cases := []struct {
		current, latest string
	}{
		{"", "v1.0.0"},
		{"v1.0.0", ""},
		{"notaversion", "v1.0.0"},
		{"v1.0.0", "garbage"},
		{"v-", "v1.0.0"},
	}
	for _, tc := range cases {
		if IsNewer(tc.current, tc.latest) {
			t.Fatalf("IsNewer(%q,%q) unexpectedly true", tc.current, tc.latest)
		}
	}
}

// TestIsNewer_PrereleaseSuffix — suffixes are dropped; the numeric prefix
// drives comparison. This keeps "v1.2.3-beta" from appearing newer than
// "v1.2.3" (the main channel), which would look wrong to the player.
func TestIsNewer_PrereleaseSuffix(t *testing.T) {
	if IsNewer("v1.2.3", "v1.2.3-beta") {
		t.Fatalf("prerelease suffix should not register as newer than stable")
	}
	if !IsNewer("v1.2.3-beta", "v1.2.4") {
		t.Fatalf("v1.2.4 should be newer than v1.2.3-beta")
	}
}

// TestStripMarkdown_Basic ensures the most common GitHub release syntax
// is normalised into readable text without leaning on a markdown library.
// Regressions here show up as garbled panel content, which would erode
// trust in the prompt.
func TestStripMarkdown_Basic(t *testing.T) {
	in := "## What's Changed\n" +
		"* **Big thing** by @you in [#42](https://example.com/42)\n" +
		"* Another `inline` fix\n" +
		"\n\n" +
		"**Full Changelog**: https://example.com/compare\n"
	out := StripMarkdown(in)
	if strings.Contains(out, "##") {
		t.Errorf("heading marker not stripped: %q", out)
	}
	if strings.Contains(out, "**") {
		t.Errorf("bold markers not stripped: %q", out)
	}
	if strings.Contains(out, "`") {
		t.Errorf("backticks not stripped: %q", out)
	}
	if !strings.Contains(out, "- ") {
		t.Errorf("bullet marker missing: %q", out)
	}
	if !strings.Contains(out, "(https://example.com/42)") {
		t.Errorf("link URL not preserved alongside text: %q", out)
	}
	// Runs of blanks should collapse.
	if strings.Contains(out, "\n\n\n") {
		t.Errorf("triple blank lines not collapsed: %q", out)
	}
}

func TestTruncateLines(t *testing.T) {
	in := "a\nb\nc\nd\ne"
	if got := TruncateLines(in, 3); got != "a\nb\nc\n…" {
		t.Fatalf("got %q", got)
	}
	if got := TruncateLines(in, 10); got != in {
		t.Fatalf("no-op truncate changed input: %q", got)
	}
	if got := TruncateLines(in, 0); got != "" {
		t.Fatalf("zero truncate should be empty: %q", got)
	}
}

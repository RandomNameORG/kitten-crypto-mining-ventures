// Package update implements a pragmatic, best-effort check against the
// project's GitHub releases feed to notify players when a newer build is
// available. It is explicitly NOT a self-updater: opening the release page
// in the player's browser is the honest UX for a cross-platform Go binary.
//
// Design notes:
//
//   - The check runs in a short-timeout goroutine at startup so offline /
//     slow networks never delay the TUI. Failures are swallowed silently.
//   - Semver parsing is deliberately minimal — split on ".", integer
//     compare, ignore suffixes. If a tag fails to parse we treat it as
//     "no update" (fail closed) so malformed tags never spam the prompt.
//   - The release body is GitHub's auto-generated notes (markdown). We do
//     NOT pull in a markdown library; we just strip the noisiest syntax
//     and keep it short enough for a splash panel.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ReleasesAPIURL is the canonical GitHub API endpoint for the "latest"
// release of the project.
const ReleasesAPIURL = "https://api.github.com/repos/RandomNameORG/kitten-crypto-mining-ventures/releases/latest"

// ReleasesPageURL is the human-facing releases page, used when we ask the
// OS to open the browser.
const ReleasesPageURL = "https://github.com/RandomNameORG/kitten-crypto-mining-ventures/releases/latest"

// Info summarises what we learned about the latest release. It is returned
// from Check and consumed by the UI layer; all fields are safe to render
// verbatim after being run through StripMarkdown for body text.
type Info struct {
	// TagName is the raw tag (e.g. "v1.2.3").
	TagName string
	// HTMLURL is the GitHub page for this release.
	HTMLURL string
	// Body is the raw release notes (markdown).
	Body string
}

// apiResponse is the subset of GitHub's release JSON we care about.
type apiResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	// Draft/Prerelease we use to skip those — we only prompt for real
	// stable releases so players aren't bounced onto unfinished builds.
	Draft      bool `json:"draft"`
	Prerelease bool `json:"prerelease"`
}

// Check queries the GitHub releases API and returns the latest release
// metadata. The supplied context controls cancellation; callers should
// supply a short timeout (3s is the house default).
//
// A non-nil error means "we couldn't determine"; callers should treat this
// as "no update" and proceed silently.
func Check(ctx context.Context) (*Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ReleasesAPIURL, nil)
	if err != nil {
		return nil, err
	}
	// GitHub recommends an explicit Accept header for the REST API.
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "meowmine-update-check")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("update check: unexpected status %d", resp.StatusCode)
	}

	var api apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&api); err != nil {
		return nil, err
	}
	if api.Draft || api.Prerelease {
		return nil, errors.New("update check: latest is a draft or prerelease")
	}
	if strings.TrimSpace(api.TagName) == "" {
		return nil, errors.New("update check: empty tag name")
	}
	return &Info{
		TagName: api.TagName,
		HTMLURL: api.HTMLURL,
		Body:    api.Body,
	}, nil
}

// IsNewer reports whether `latest` is strictly greater than `current`
// using a minimal semver comparison (leading `v` stripped, dot-separated
// integers, shorter version treated as having trailing zeroes).
//
// If either version fails to parse we return false — callers should
// prefer silence over a spammy prompt when data is ambiguous.
func IsNewer(current, latest string) bool {
	cur, ok1 := parseSemver(current)
	lat, ok2 := parseSemver(latest)
	if !ok1 || !ok2 {
		return false
	}
	// Pad to equal length so e.g. "1.0" vs "1.0.0" compares as equal.
	for len(cur) < len(lat) {
		cur = append(cur, 0)
	}
	for len(lat) < len(cur) {
		lat = append(lat, 0)
	}
	for i := range cur {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	return false
}

// parseSemver converts "v1.2.3" / "1.2.3" / "1.2" into a slice of ints.
// Trailing suffixes like "-beta.1" or "+build.7" are dropped after the
// first non-numeric segment, keeping the numeric prefix. Returns ok=false
// if the string has zero numeric segments.
func parseSemver(s string) ([]int, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if s == "" {
		return nil, false
	}
	// Cut pre-release / build metadata; everything before the first
	// non-digit-non-dot byte is the numeric core.
	cut := len(s)
	for i, r := range s {
		if r == '.' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		cut = i
		break
	}
	core := s[:cut]
	if core == "" {
		return nil, false
	}
	parts := strings.Split(core, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			return nil, false
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// StripMarkdown converts a GitHub release body (lightweight markdown) into
// plain-ish text suitable for a TUI splash. It is intentionally simple —
// the goal is legibility, not fidelity. Headings become plain lines,
// bullets keep their shape, inline emphasis (`**bold**`, `*em*`, `` ` ``)
// is dropped, link syntax `[text](url)` becomes `text (url)`.
func StripMarkdown(md string) string {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		// Trim ATX headings: "## Foo" -> "Foo".
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "#") {
			line = strings.TrimLeft(trimmed, "#")
			line = strings.TrimSpace(line)
		}
		// Normalise bullets: "* foo", "+ foo" -> "- foo".
		t := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
			indent := line[:len(line)-len(t)]
			line = indent + "- " + strings.TrimPrefix(strings.TrimPrefix(t, "* "), "+ ")
		}
		// Drop bold/italic/code markers and HTML comments.
		line = stripInline(line)
		out = append(out, line)
	}
	// Collapse runs of blank lines — GitHub notes often have lots.
	collapsed := make([]string, 0, len(out))
	blank := false
	for _, l := range out {
		if strings.TrimSpace(l) == "" {
			if blank {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		collapsed = append(collapsed, l)
	}
	return strings.TrimSpace(strings.Join(collapsed, "\n"))
}

// stripInline removes the noisiest inline markdown tokens.
func stripInline(s string) string {
	// [text](url) -> text (url)
	s = linkRegex(s)
	// Drop **bold** and __bold__ markers but keep the text.
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	// Drop inline code backticks.
	s = strings.ReplaceAll(s, "`", "")
	// Drop single-char emphasis markers ( *em* / _em_ ) — naive, but
	// markdown that reaches us here is GitHub-generated so it's tidy.
	s = strings.ReplaceAll(s, "*", "")
	// Drop leading "> " blockquotes.
	trim := strings.TrimLeft(s, " \t")
	if strings.HasPrefix(trim, "> ") {
		indent := s[:len(s)-len(trim)]
		s = indent + strings.TrimPrefix(trim, "> ")
	}
	return s
}

// linkRegex converts "[text](url)" -> "text (url)" without pulling in the
// regexp package for such a tiny rewrite.
func linkRegex(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '[' {
			b.WriteByte(s[i])
			i++
			continue
		}
		// Find matching "](".
		close1 := strings.Index(s[i:], "](")
		if close1 < 0 {
			b.WriteByte(s[i])
			i++
			continue
		}
		close1 += i
		close2 := strings.Index(s[close1:], ")")
		if close2 < 0 {
			b.WriteByte(s[i])
			i++
			continue
		}
		close2 += close1
		text := s[i+1 : close1]
		url := s[close1+2 : close2]
		if text != "" && url != "" {
			b.WriteString(text)
			b.WriteString(" (")
			b.WriteString(url)
			b.WriteString(")")
		} else {
			b.WriteString(s[i : close2+1])
		}
		i = close2 + 1
	}
	return b.String()
}

// TruncateLines clips `text` to at most `max` lines, adding an ellipsis
// marker line when truncation occurs. Used to keep the changelog panel
// from exploding on verbose releases.
func TruncateLines(text string, max int) string {
	if max <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= max {
		return text
	}
	kept := lines[:max]
	return strings.Join(kept, "\n") + "\n…"
}

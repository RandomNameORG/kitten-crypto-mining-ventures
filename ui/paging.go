package ui

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// listPageSize returns how many list rows the current view can show without
// pushing the header/footer offscreen. Conservative: subtracts a fixed
// amount of chrome (header, nav, view title, view help, sort label, blank
// spacer, footer, hint, padding) plus a safety margin.
func (a App) listPageSize() int {
	const chrome = 12
	const minRows = 5
	if a.h <= 0 {
		// Pre-WindowSizeMsg fallback. Pick a roomy default so the first
		// frame doesn't render a stub.
		return 30
	}
	if a.h <= chrome+minRows {
		return minRows
	}
	return a.h - chrome
}

// pageWindow returns [start, end) — the slice of items visible on the
// current page given total items, the cursor position, and a target page
// size. The cursor is kept centred when possible; clamps at the edges so
// no out-of-bounds slicing is needed at the call site.
func pageWindow(total, cursor, size int) (start, end int) {
	if total <= 0 || size <= 0 {
		return 0, 0
	}
	if total <= size {
		return 0, total
	}
	half := size / 2
	start = cursor - half
	if start < 0 {
		start = 0
	}
	if start+size > total {
		start = total - size
	}
	if start < 0 {
		start = 0
	}
	end = start + size
	if end > total {
		end = total
	}
	return start, end
}

// pagingHint renders a single-line "page X/Y · ↑↓ to scroll" indicator,
// or empty string when the whole list fits on one page.
func pagingHint(total, cursor, size int) string {
	if total == 0 || size <= 0 || total <= size {
		return ""
	}
	start, _ := pageWindow(total, cursor, size)
	page := start/size + 1
	totalPages := (total + size - 1) / size
	return DimStyle.Render(fmt.Sprintf(i18n.T("paging.hint"), page, totalPages, total))
}

// bodyMaxRows is the max body height the View can render without pushing
// the header offscreen. Reserves a fixed budget for header+nav+footer+hint.
func (a App) bodyMaxRows() int {
	const reserved = 6
	if a.h <= 0 {
		return 50 // pre-WindowSizeMsg sane default
	}
	if a.h <= reserved+5 {
		return 5
	}
	return a.h - reserved
}

// clipBody truncates `body` to at most maxRows visible lines. If trimming
// happened, replaces the last line with a hint that more content was hidden.
// Lipgloss-rendered borders count as lines too — this is a coarse but
// reliable safety net.
func clipBody(body string, maxRows int) string {
	if maxRows <= 0 {
		return body
	}
	lines := splitLines(body)
	if len(lines) <= maxRows {
		return body
	}
	keep := lines[:maxRows-1]
	hidden := len(lines) - (maxRows - 1)
	keep = append(keep, DimStyle.Render(fmt.Sprintf(i18n.T("paging.clip"), hidden)))
	return joinLines(keep)
}

// splitLines walks body once, faster than strings.Split which allocates a
// slice eagerly even for short inputs.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	out := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

// joinLines is the inverse of splitLines.
func joinLines(lines []string) string {
	switch len(lines) {
	case 0:
		return ""
	case 1:
		return lines[0]
	}
	total := len(lines) - 1
	for _, l := range lines {
		total += len(l)
	}
	out := make([]byte, 0, total)
	for i, l := range lines {
		if i > 0 {
			out = append(out, '\n')
		}
		out = append(out, l...)
	}
	return string(out)
}

package ui

import (
	"fmt"

	"github.com/RandomNameORG/kitten-crypto-mining-ventures/core/i18n"
)

// listPageSize is the FIXED rows-per-page for cursor-driven list views.
// Intentionally NOT terminal-height-aware: the player wants pages to mean
// "groups of N", not "however many fit on screen", so the indicator and
// arrow-key flips behave predictably even on tall terminals.
//
// Per-view overrides exist in renderXxxView when a different chunk size
// makes more sense (e.g. mercs has bounded hireable count).
func (a App) listPageSize() int { return 10 }

// pageWindow returns [start, end) — the slice of items visible on the
// current page. STRICT pagination: items are partitioned into fixed
// non-overlapping groups of `size` each, and the page is determined by
// floor(cursor / size). Use ←/→ keys (or [/]) to jump full pages; ↑/↓
// glides one item at a time, naturally crossing into the adjacent page
// when the cursor steps over the boundary.
func pageWindow(total, cursor, size int) (start, end int) {
	if total <= 0 || size <= 0 {
		return 0, 0
	}
	if total <= size {
		return 0, total
	}
	page := cursor / size
	start = page * size
	end = start + size
	if end > total {
		end = total
	}
	return start, end
}

// pagingHint renders the "← page X/Y → · N total" indicator. Empty when
// everything fits on one page.
func pagingHint(total, cursor, size int) string {
	if total == 0 || size <= 0 || total <= size {
		return ""
	}
	page := cursor/size + 1
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

// bodyPagePerPage is how many body lines fit per page once the paging hint
// is reserved. Always at least 1 so a tiny terminal still shows something.
func bodyPagePerPage(maxRows int) int {
	per := maxRows - 1 // reserve one row for the page hint
	if per < 1 {
		per = 1
	}
	return per
}

// bodyTotalPages reports how many pages `body` would split into at this
// terminal height. 1 when everything fits.
func bodyTotalPages(body string, maxRows int) int {
	if maxRows <= 0 {
		return 1
	}
	lines := splitLines(body)
	if len(lines) <= maxRows {
		return 1
	}
	per := bodyPagePerPage(maxRows)
	return (len(lines) + per - 1) / per
}

// paginateBody slices `body` into the page indexed by `page` (0-based) at
// the current terminal height. When body fits in one screen, returns it
// unchanged with totalPages=1. Otherwise reserves the bottom row for a
// "← page X/Y →" hint, which the caller appends.
func paginateBody(body string, maxRows, page int) (slice string, totalPages, currentPage int) {
	if maxRows <= 0 {
		return body, 1, 0
	}
	lines := splitLines(body)
	if len(lines) <= maxRows {
		return body, 1, 0
	}
	per := bodyPagePerPage(maxRows)
	totalPages = (len(lines) + per - 1) / per
	currentPage = page
	if currentPage >= totalPages {
		currentPage = totalPages - 1
	}
	if currentPage < 0 {
		currentPage = 0
	}
	start := currentPage * per
	end := start + per
	if end > len(lines) {
		end = len(lines)
	}
	return joinLines(lines[start:end]), totalPages, currentPage
}

// bodyPagingHint renders the "← page X/Y →" indicator shown at the bottom
// of a paginated body. Mirrors pagingHint's styling.
func bodyPagingHint(totalPages, page int) string {
	if totalPages <= 1 {
		return ""
	}
	return DimStyle.Render(fmt.Sprintf(i18n.T("paging.body_hint"), page+1, totalPages))
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

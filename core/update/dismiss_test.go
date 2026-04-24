package update

import (
	"path/filepath"
	"testing"
)

// TestDismissRoundtrip — the user-visible promise is "Skip this version"
// silences future prompts for that exact tag. This test anchors that
// round-trip: write, read, confirm the tag is remembered.
func TestDismissRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dismissed_versions.txt")

	list, err := LoadDismissed(path)
	if err != nil {
		t.Fatalf("load missing file: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %v", list)
	}

	list = AppendDismissed(list, "v1.2.3")
	list = AppendDismissed(list, "v1.2.3") // dedup check
	list = AppendDismissed(list, "v1.3.0")
	if len(list) != 2 {
		t.Fatalf("expected dedup to 2 entries, got %v", list)
	}
	if err := SaveDismissed(path, list); err != nil {
		t.Fatalf("save: %v", err)
	}

	back, err := LoadDismissed(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(back) != 2 {
		t.Fatalf("expected 2 persisted, got %v", back)
	}
	if !IsDismissed(back, "v1.2.3") {
		t.Fatalf("v1.2.3 should be dismissed: %v", back)
	}
	if !IsDismissed(back, "V1.2.3") {
		t.Fatalf("case-insensitive match failed: %v", back)
	}
	if IsDismissed(back, "v9.9.9") {
		t.Fatalf("v9.9.9 should not be dismissed")
	}
}

// TestAppendDismissed_EmptyTag — we never want stray blank lines in the
// persisted file; they'd accumulate on every write.
func TestAppendDismissed_EmptyTag(t *testing.T) {
	got := AppendDismissed(nil, "   ")
	if len(got) != 0 {
		t.Fatalf("blank tag should not be appended: %v", got)
	}
}

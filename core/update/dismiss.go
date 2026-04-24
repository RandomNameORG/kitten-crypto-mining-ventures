// dismiss.go manages the tiny cross-run config file that remembers which
// release tags the player has "skip"-dismissed. Kept deliberately outside
// the game State because:
//
//   - It's meta-config (user preference), not save data.
//   - It persists even when the player starts a fresh save.
//   - It needs to be cheap to read/write at startup before the full
//     state is loaded.
//
// Format is the simplest thing that works: one tag per line in
// `~/.meowmine/dismissed_versions.txt`. Order / dedup is enforced on
// write. Missing file is fine — means "none dismissed yet".
package update

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// DefaultDismissFile returns the conventional location for the dismissed
// versions file. Mirrors game.SavePath() by living under ~/.meowmine/.
func DefaultDismissFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".meowmine", "dismissed_versions.txt")
	}
	return filepath.Join(home, ".meowmine", "dismissed_versions.txt")
}

// LoadDismissed reads the dismissed-versions file at `path`. A missing
// file yields an empty slice and nil error — "nothing dismissed yet" is
// the normal first-run state.
func LoadDismissed(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveDismissed writes the deduplicated list to `path`. Callers should
// typically pass the result of AppendDismissed so ordering is preserved.
func SaveDismissed(path string, tags []string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	var buf bytes.Buffer
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		buf.WriteString(t)
		buf.WriteByte('\n')
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// IsDismissed reports whether `tag` is present in `list` (exact match,
// case-insensitive — tags are conventionally lowercase but we don't want
// a stray "V1.2.3" to defeat the check).
func IsDismissed(list []string, tag string) bool {
	needle := strings.ToLower(strings.TrimSpace(tag))
	if needle == "" {
		return false
	}
	for _, t := range list {
		if strings.ToLower(strings.TrimSpace(t)) == needle {
			return true
		}
	}
	return false
}

// AppendDismissed returns a new list containing `tag` exactly once.
// If `tag` is already present the original list is returned unchanged.
func AppendDismissed(list []string, tag string) []string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return list
	}
	if IsDismissed(list, tag) {
		return list
	}
	return append(append([]string(nil), list...), tag)
}

package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"syl-md2ppt/internal/config"
)

func TestDiscoverSortAndPair(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "SPI")
	en := filepath.Join(source, "EN")
	cn := filepath.Join(source, "CN")
	if err := os.MkdirAll(filepath.Join(en, "Domain"), 0o755); err != nil {
		t.Fatalf("mkdir en: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cn, "Domain"), 0o755); err != nil {
		t.Fatalf("mkdir cn: %v", err)
	}

	mustWrite(t, filepath.Join(en, "Domain", "1-002-Back.md"), "EN back")
	mustWrite(t, filepath.Join(en, "Domain", "1-002-Front.md"), "EN front")
	mustWrite(t, filepath.Join(en, "Domain", "1-003-Front.md"), "EN next front")
	mustWrite(t, filepath.Join(en, "Domain", "1-003-Back.md"), "EN next back")
	mustWrite(t, filepath.Join(en, "Domain", "README.md"), "ignored")

	mustWrite(t, filepath.Join(cn, "Domain", "1-002-Back.md"), "CN back")
	mustWrite(t, filepath.Join(cn, "Domain", "1-002-Front.md"), "CN front")
	mustWrite(t, filepath.Join(cn, "Domain", "1-003-Front.md"), "CN next front")
	mustWrite(t, filepath.Join(cn, "Domain", "1-003-Back.md"), "CN next back")

	cfg := &config.Config{}
	cfg.Filename.Pattern = `^(\d+)-(\d{3})-(Front|Back)\.md$`
	cfg.Filename.Groups.Domain = 1
	cfg.Filename.Groups.Card = 2
	cfg.Filename.Groups.Side = 3
	cfg.Filename.Order.Side = []string{"Front", "Back"}
	cfg.Filename.IgnoreUnmatched = true

	pairs, warnings, err := Discover(source, cfg)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for unmatched file, got %d", len(warnings))
	}
	if len(pairs) != 4 {
		t.Fatalf("expected 4 pairs, got %d", len(pairs))
	}

	got := []string{pairs[0].RelPath, pairs[1].RelPath, pairs[2].RelPath, pairs[3].RelPath}
	want := []string{
		"Domain/1-002-Front.md",
		"Domain/1-002-Back.md",
		"Domain/1-003-Front.md",
		"Domain/1-003-Back.md",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected order at %d: want=%s got=%s", i, want[i], got[i])
		}
	}
}

func TestDiscoverFailsOnMissingPair(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "SPI")
	en := filepath.Join(source, "EN")
	cn := filepath.Join(source, "CN")
	if err := os.MkdirAll(filepath.Join(en, "Domain"), 0o755); err != nil {
		t.Fatalf("mkdir en: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(cn, "Domain"), 0o755); err != nil {
		t.Fatalf("mkdir cn: %v", err)
	}

	mustWrite(t, filepath.Join(en, "Domain", "1-002-Front.md"), "EN")
	mustWrite(t, filepath.Join(cn, "Domain", "1-002-Front.md"), "CN")
	mustWrite(t, filepath.Join(en, "Domain", "1-002-Back.md"), "EN back")

	cfg := &config.Config{}
	cfg.Filename.Pattern = `^(\d+)-(\d{3})-(Front|Back)\.md$`
	cfg.Filename.Groups.Domain = 1
	cfg.Filename.Groups.Card = 2
	cfg.Filename.Groups.Side = 3
	cfg.Filename.Order.Side = []string{"Front", "Back"}
	cfg.Filename.IgnoreUnmatched = true

	_, _, err := Discover(source, cfg)
	if err == nil {
		t.Fatalf("expected pairing error, got nil")
	}
}

func mustWrite(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

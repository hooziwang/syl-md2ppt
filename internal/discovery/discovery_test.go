package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"syl-md2ppt/internal/config"
)

func TestNumberFingerprintUsesNonRepeatedNumbers(t *testing.T) {
	key, nums := numberFingerprint("deck-1-1-002-front.md")
	if key != "002" {
		t.Fatalf("unexpected key: %s", key)
	}
	if len(nums) != 1 || nums[0] != 2 {
		t.Fatalf("unexpected nums: %#v", nums)
	}

	key, nums = numberFingerprint("lesson-01-topic-003.md")
	if key != "01-003" {
		t.Fatalf("unexpected key: %s", key)
	}
	if len(nums) != 2 || nums[0] != 1 || nums[1] != 3 {
		t.Fatalf("unexpected nums: %#v", nums)
	}

	key, nums = numberFingerprint("readme.md")
	if key != "" || nums != nil {
		t.Fatalf("expected empty fingerprint for file without numbers")
	}
}

func TestDiscoverSortAndPairByUniqueNumbers(t *testing.T) {
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

	mustWrite(t, filepath.Join(en, "Domain", "deck-1-1-002-front.md"), "EN front")
	mustWrite(t, filepath.Join(en, "Domain", "deck-1-1-002-back.md"), "EN back")
	mustWrite(t, filepath.Join(en, "Domain", "L-3-3-010-A.md"), "EN third")
	mustWrite(t, filepath.Join(en, "Domain", "README.md"), "ignored")

	mustWrite(t, filepath.Join(cn, "Domain", "课程-1-1-002-front.md"), "CN front")
	mustWrite(t, filepath.Join(cn, "Domain", "课程-1-1-002-back.md"), "CN back")
	mustWrite(t, filepath.Join(cn, "Domain", "课程-3-3-010-A.md"), "CN third")

	cfg := &config.Config{}
	cfg.Filename.IgnoreUnmatched = true

	pairs, warnings, err := Discover(source, cfg)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for unmatched file, got %d", len(warnings))
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	if pairs[0].RelPath != "Domain/deck-1-1-002-front.md" {
		t.Fatalf("expected front first, got %s", pairs[0].RelPath)
	}
	if pairs[1].RelPath != "Domain/deck-1-1-002-back.md" {
		t.Fatalf("expected back second, got %s", pairs[1].RelPath)
	}
	if pairs[2].RelPath != "Domain/L-3-3-010-A.md" {
		t.Fatalf("expected remaining file third, got %s", pairs[2].RelPath)
	}
}

func TestDiscoverRecognizesAAndBAsFrontBack(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "SPI")
	en := filepath.Join(source, "EN", "D")
	cn := filepath.Join(source, "CN", "D")
	if err := os.MkdirAll(en, 0o755); err != nil {
		t.Fatalf("mkdir en: %v", err)
	}
	if err := os.MkdirAll(cn, 0o755); err != nil {
		t.Fatalf("mkdir cn: %v", err)
	}

	mustWrite(t, filepath.Join(en, "2-2-011-B.md"), "EN back")
	mustWrite(t, filepath.Join(en, "2-2-011-A.md"), "EN front")
	mustWrite(t, filepath.Join(cn, "2-2-011-B.md"), "CN back")
	mustWrite(t, filepath.Join(cn, "2-2-011-A.md"), "CN front")

	cfg := &config.Config{}
	cfg.Filename.IgnoreUnmatched = true

	pairs, _, err := Discover(source, cfg)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].RelPath != "D/2-2-011-A.md" || pairs[1].RelPath != "D/2-2-011-B.md" {
		t.Fatalf("expected A then B order, got %s then %s", pairs[0].RelPath, pairs[1].RelPath)
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

	mustWrite(t, filepath.Join(en, "Domain", "deck-1-1-002-front.md"), "EN")
	mustWrite(t, filepath.Join(cn, "Domain", "课程-1-1-002-front.md"), "CN")
	mustWrite(t, filepath.Join(en, "Domain", "deck-1-1-002-back.md"), "EN back")

	cfg := &config.Config{}
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

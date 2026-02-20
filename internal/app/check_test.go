package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheck_OK(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "SPI")
	enDir := filepath.Join(source, "EN", "D")
	cnDir := filepath.Join(source, "CN", "D")
	if err := os.MkdirAll(enDir, 0o755); err != nil {
		t.Fatalf("mkdir en: %v", err)
	}
	if err := os.MkdirAll(cnDir, 0o755); err != nil {
		t.Fatalf("mkdir cn: %v", err)
	}
	if err := os.WriteFile(filepath.Join(enDir, "deck-1-1-002-front.md"), []byte("EN"), 0o644); err != nil {
		t.Fatalf("write en: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "课-1-1-002-front.md"), []byte("CN"), 0o644); err != nil {
		t.Fatalf("write cn: %v", err)
	}
	if err := os.WriteFile(filepath.Join(enDir, "deck-1-1-002-back.md"), []byte("EN back"), 0o644); err != nil {
		t.Fatalf("write en back: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "课-1-1-002-back.md"), []byte("CN back"), 0o644); err != nil {
		t.Fatalf("write cn back: %v", err)
	}

	res, err := Check(Options{SourceDir: source, CWD: tmp})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if res.PairCount != 2 {
		t.Fatalf("expected pair count 2, got %d", res.PairCount)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res.Items))
	}
	if res.Items[0].No != 1 || res.Items[1].No != 2 {
		t.Fatalf("expected sequential page numbers 1,2 got %d,%d", res.Items[0].No, res.Items[1].No)
	}
	if res.Items[0].ENPath != filepath.Join(enDir, "deck-1-1-002-front.md") {
		t.Fatalf("unexpected EN path: %s", res.Items[0].ENPath)
	}
	if res.Items[1].ENPath != filepath.Join(enDir, "deck-1-1-002-back.md") {
		t.Fatalf("unexpected EN back path: %s", res.Items[1].ENPath)
	}
}

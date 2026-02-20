package output

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"
)

func TestResolveOutputPath_DefaultName(t *testing.T) {
	now := time.Date(2026, 2, 20, 18, 4, 5, 0, time.UTC)
	cwd := "/tmp/work"
	randSrc := bytes.NewBufferString("ABCDEF")

	got, err := ResolveOutputPath("", cwd, now, randSrc)
	if err != nil {
		t.Fatalf("ResolveOutputPath returned error: %v", err)
	}

	want := filepath.Join(cwd, "20260220_180405_ABCDEF.pptx")
	if got != want {
		t.Fatalf("unexpected output path\nwant: %s\n got: %s", want, got)
	}
}

func TestResolveOutputPath_AsDirectory(t *testing.T) {
	now := time.Date(2026, 2, 20, 18, 4, 5, 0, time.UTC)
	cwd := "/tmp/work"
	randSrc := bytes.NewBufferString("QWERTY")

	got, err := ResolveOutputPath("/tmp/outdir", cwd, now, randSrc)
	if err != nil {
		t.Fatalf("ResolveOutputPath returned error: %v", err)
	}

	want := filepath.Join("/tmp/outdir", "20260220_180405_QWERTY.pptx")
	if got != want {
		t.Fatalf("unexpected output path\nwant: %s\n got: %s", want, got)
	}
}

func TestResolveOutputPath_AsFile(t *testing.T) {
	now := time.Date(2026, 2, 20, 18, 4, 5, 0, time.UTC)
	cwd := "/tmp/work"
	randSrc := bytes.NewBufferString("ZXCVBN")

	got, err := ResolveOutputPath("./final.pptx", cwd, now, randSrc)
	if err != nil {
		t.Fatalf("ResolveOutputPath returned error: %v", err)
	}

	want := filepath.Join(cwd, "final.pptx")
	if got != want {
		t.Fatalf("unexpected output path\nwant: %s\n got: %s", want, got)
	}
}

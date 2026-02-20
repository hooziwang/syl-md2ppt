package app

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun_EndToEnd(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(enDir, "1-002-Front.md"), []byte("★ **EN** $a+b$"), 0o644); err != nil {
		t.Fatalf("write en1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "1-002-Front.md"), []byte("★ **中** $甲+乙$"), 0o644); err != nil {
		t.Fatalf("write cn1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(enDir, "1-002-Back.md"), []byte("▲ warn"), 0o644); err != nil {
		t.Fatalf("write en2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "1-002-Back.md"), []byte("▲ 警示"), 0o644); err != nil {
		t.Fatalf("write cn2: %v", err)
	}

	now := time.Date(2026, 2, 20, 19, 0, 0, 0, time.UTC)
	rand := bytes.NewBufferString("ABCDEF")

	res, err := Run(Options{
		SourceDir: source,
		OutputArg: filepath.Join(tmp, "out"),
		CWD:       tmp,
		Now:       now,
		Rand:      rand,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if res.SlideCount != 2 {
		t.Fatalf("expected 2 slides, got %d", res.SlideCount)
	}
	if !strings.HasSuffix(res.OutputPath, ".pptx") {
		t.Fatalf("expected pptx output, got %s", res.OutputPath)
	}

	zr, err := zip.OpenReader(res.OutputPath)
	if err != nil {
		t.Fatalf("open output pptx: %v", err)
	}
	defer zr.Close()
	count := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 slide xml files, got %d", count)
	}
}

func TestRun_WarningContainsSlideNumber(t *testing.T) {
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

	heavyEN := strings.Repeat("long english text ", 500)
	heavyCN := strings.Repeat("很长的中文内容", 500)
	if err := os.WriteFile(filepath.Join(enDir, "1-002-Front.md"), []byte(heavyEN), 0o644); err != nil {
		t.Fatalf("write en: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "1-002-Front.md"), []byte(heavyCN), 0o644); err != nil {
		t.Fatalf("write cn: %v", err)
	}

	now := time.Date(2026, 2, 20, 19, 0, 0, 0, time.UTC)
	rand := bytes.NewBufferString("ABCDEF")
	res, err := Run(Options{
		SourceDir: source,
		OutputArg: filepath.Join(tmp, "out"),
		CWD:       tmp,
		Now:       now,
		Rand:      rand,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(res.Warnings) == 0 {
		t.Fatalf("expected at least one warning")
	}
	if strings.HasPrefix(res.Warnings[0], "提醒：") {
		t.Fatalf("warning should not have 提醒 prefix, got: %s", res.Warnings[0])
	}
	if !strings.Contains(res.Warnings[0], "[  1] - ") {
		t.Fatalf("warning should include formatted slide number, got: %s", res.Warnings[0])
	}
	enAbs := filepath.Join(enDir, "1-002-Front.md")
	cnAbs := filepath.Join(cnDir, "1-002-Front.md")
	if len(res.Warnings) < 2 {
		t.Fatalf("expected at least two warnings, got: %d", len(res.Warnings))
	}
	if !strings.Contains(res.Warnings[0], enAbs) {
		t.Fatalf("first warning should contain EN absolute path, got: %s", res.Warnings[0])
	}
	if !strings.Contains(res.Warnings[1], cnAbs) {
		t.Fatalf("second warning should contain CN absolute path, got: %s", res.Warnings[1])
	}
	if !strings.HasSuffix(res.Warnings[0], "内容有点多，部分截断") {
		t.Fatalf("warning should use concise truncate message, got: %s", res.Warnings[0])
	}
}

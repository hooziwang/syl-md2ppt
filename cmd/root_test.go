package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func TestBuildRequiresSourceArg(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"build"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for missing data_source_dir")
	}
	if err.Error() != "还没给数据源目录。用法：syl-md2ppt <数据源目录>" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output when source arg is missing, got: %q", stdout.String())
	}
}

func TestRootNoArgsErrorThenHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for missing data_source_dir")
	}
	if !errors.Is(err, errAlreadyPrinted) {
		t.Fatalf("expected handled error marker, got: %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got: %q", stdout.String())
	}
	got := stderr.String()
	if !strings.HasPrefix(got, "还没给数据源目录。用法：syl-md2ppt <数据源目录>\n\n") {
		t.Fatalf("expected error then blank line, got: %q", got)
	}
	if !strings.Contains(got, "Usage:\n  syl-md2ppt [data_source_dir] [flags]") {
		t.Fatalf("expected help output after error, got: %q", got)
	}
}

func TestNormalizeArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "direct run", in: []string{"./SPI", "--output", "./out"}, want: []string{"build", "./SPI", "--output", "./out"}},
		{name: "flag first", in: []string{"--output", "./out", "./SPI"}, want: []string{"build", "--output", "./out", "./SPI"}},
		{name: "build command", in: []string{"build", "./SPI"}, want: []string{"build", "./SPI"}},
		{name: "check command", in: []string{"check", "./SPI"}, want: []string{"check", "./SPI"}},
		{name: "help flag", in: []string{"--help"}, want: []string{"--help"}},
	}

	for _, tc := range tests {
		got := normalizeArgs(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: unexpected args want=%v got=%v", tc.name, tc.want, got)
		}
	}
}

func TestCheckRequiresSourceArg(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"check"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for missing data_source_dir")
	}
	if err.Error() != "还没给数据源目录。用法：syl-md2ppt check <数据源目录>" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output when source arg is missing, got: %q", stdout.String())
	}
}

func TestCheckShowsDetailedList(t *testing.T) {
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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"check", source})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "检查通过：共识别 2 对双语文件，可生成 2 页 PPT") {
		t.Fatalf("unexpected check summary output: %q", out)
	}
	if !strings.Contains(out, "[001] - "+filepath.Join(enDir, "deck-1-1-002-front.md")) {
		t.Fatalf("expected first page output, got: %q", out)
	}
	if !strings.Contains(out, "[001] - "+filepath.Join(cnDir, "课-1-1-002-front.md")) {
		t.Fatalf("expected first page cn output, got: %q", out)
	}
	if !strings.Contains(out, "[002] - "+filepath.Join(enDir, "deck-1-1-002-back.md")) {
		t.Fatalf("expected second page output, got: %q", out)
	}
	if !strings.Contains(out, "[002] - "+filepath.Join(cnDir, "课-1-1-002-back.md")) {
		t.Fatalf("expected second page cn output, got: %q", out)
	}
	if strings.Index(out, "[001] - ") > strings.Index(out, "检查通过：") {
		t.Fatalf("expected detail list before summary, got: %q", out)
	}
}

func TestCheckShowsConflictSummary(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(enDir, "deck-1-1-012-X.md"), []byte("EN X"), 0o644); err != nil {
		t.Fatalf("write en x: %v", err)
	}
	if err := os.WriteFile(filepath.Join(enDir, "deck-1-1-012-Y.md"), []byte("EN Y"), 0o644); err != nil {
		t.Fatalf("write en y: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "课-1-1-012-X.md"), []byte("CN X"), 0o644); err != nil {
		t.Fatalf("write cn x: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cnDir, "课-1-1-012-Y.md"), []byte("CN Y"), 0o644); err != nil {
		t.Fatalf("write cn y: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"check", source})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "检查完成：共识别 2 对双语文件，可生成 2 页 PPT；发现 1 组冲突，请先人工确认") {
		t.Fatalf("unexpected conflict summary output: %q", out)
	}
	if strings.Contains(out, "检查通过：") {
		t.Fatalf("should not show pass summary when conflict exists: %q", out)
	}
}

func TestVersionCommand(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := stripANSI(strings.TrimSpace(stdout.String()))
	if !strings.Contains(got, versionText()) {
		t.Fatalf("unexpected version output: %q", got)
	}
	if !strings.Contains(got, "DADDYLOVESYL") {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestVersionFlagWithoutSource(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root := NewRootCmd(func() time.Time {
		return time.Date(2026, 2, 20, 20, 0, 0, 0, time.UTC)
	}, bytes.NewBufferString("ABCDEF"), stdout, stderr)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := stripANSI(strings.TrimSpace(stdout.String()))
	if !strings.Contains(got, versionText()) {
		t.Fatalf("unexpected version output: %q", got)
	}
	if !strings.Contains(got, "DADDYLOVESYL") {
		t.Fatalf("unexpected version output: %q", got)
	}
}

package cmd

import (
	"bytes"
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
		{name: "help flag", in: []string{"--help"}, want: []string{"--help"}},
	}

	for _, tc := range tests {
		got := normalizeArgs(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%s: unexpected args want=%v got=%v", tc.name, tc.want, got)
		}
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

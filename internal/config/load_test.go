package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Priority(t *testing.T) {
	tmp := t.TempDir()
	projectCfg := filepath.Join(tmp, "syl-md2ppt.yaml")
	overrideCfg := filepath.Join(tmp, "override.yaml")

	if err := os.WriteFile(projectCfg, []byte("filename:\n  ignore_unmatched: true\n"), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	if err := os.WriteFile(overrideCfg, []byte("filename:\n  ignore_unmatched: false\n"), 0o644); err != nil {
		t.Fatalf("write override config: %v", err)
	}

	cfg, src, err := Load(overrideCfg, tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if src != overrideCfg {
		t.Fatalf("expected source %s, got %s", overrideCfg, src)
	}
	if cfg.Filename.IgnoreUnmatched {
		t.Fatalf("expected override ignore_unmatched=false")
	}

	cfg, src, err = Load("", tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if src != projectCfg {
		t.Fatalf("expected source %s, got %s", projectCfg, src)
	}
	if !cfg.Filename.IgnoreUnmatched {
		t.Fatalf("expected project ignore_unmatched=true")
	}
}

func TestLoadConfig_DefaultFallback(t *testing.T) {
	tmp := t.TempDir()
	cfg, src, err := Load("", tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if src != "embedded:default.yaml" {
		t.Fatalf("unexpected source: %s", src)
	}
	if cfg.Layout.Typography.BaseSize <= 0 {
		t.Fatalf("default base font should be positive")
	}
}

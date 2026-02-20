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

	if err := os.WriteFile(projectCfg, []byte("filename:\n  pattern: '^A-(\\d+)-(Front|Back)\\.md$'\n"), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}
	if err := os.WriteFile(overrideCfg, []byte("filename:\n  pattern: '^B-(\\d+)-(Front|Back)\\.md$'\n"), 0o644); err != nil {
		t.Fatalf("write override config: %v", err)
	}

	cfg, src, err := Load(overrideCfg, tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if src != overrideCfg {
		t.Fatalf("expected source %s, got %s", overrideCfg, src)
	}
	if cfg.Filename.Pattern != "^B-(\\d+)-(Front|Back)\\.md$" {
		t.Fatalf("unexpected override pattern: %s", cfg.Filename.Pattern)
	}

	cfg, src, err = Load("", tmp)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if src != projectCfg {
		t.Fatalf("expected source %s, got %s", projectCfg, src)
	}
	if cfg.Filename.Pattern != "^A-(\\d+)-(Front|Back)\\.md$" {
		t.Fatalf("unexpected project pattern: %s", cfg.Filename.Pattern)
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
	if cfg.Filename.Pattern == "" {
		t.Fatalf("default filename pattern should not be empty")
	}
	if len(cfg.Filename.Order.Side) == 0 {
		t.Fatalf("default side order should not be empty")
	}
}

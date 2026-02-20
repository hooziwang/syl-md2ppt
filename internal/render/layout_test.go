package render

import (
	"strings"
	"testing"

	"syl-md2ppt/internal/config"
)

func TestBuildSlideTwoColumns(t *testing.T) {
	cfg := minimalConfig()
	slide, warnings := BuildSlide("hello EN", "你好 CN", cfg)
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
	if len(slide.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(slide.Columns))
	}
	if slide.Columns[0].Lang != "EN" || slide.Columns[1].Lang != "CN" {
		t.Fatalf("unexpected language order: %#v", slide.Columns)
	}
}

func TestBuildSlideShrinkAndTruncate(t *testing.T) {
	cfg := minimalConfig()
	cfg.Layout.Typography.BaseSize = 24
	cfg.Layout.Typography.MinSize = 10
	heavy := strings.Repeat("long text ", 400)
	slide, warnings := BuildSlide(heavy, heavy, cfg)
	if slide.FontSize > cfg.Layout.Typography.BaseSize {
		t.Fatalf("font size should not exceed base")
	}
	if slide.FontSize < cfg.Layout.Typography.MinSize {
		t.Fatalf("font size should not be below minimum")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected truncation warning for heavy content")
	}
	if !slide.HasTruncationBadge {
		t.Fatalf("expected truncation badge when warnings exist")
	}
}

func TestBuildSlide_CNCanAddOneExtraColumnIndependently(t *testing.T) {
	cfg := minimalConfig()
	en := "short english"
	cn := strings.Repeat("这是比较长的中文内容。", 520)
	slide, _ := BuildSlide(en, cn, cfg)
	if slide.ENNumCol != 1 {
		t.Fatalf("expected EN to keep single column, got EN=%d", slide.ENNumCol)
	}
	if slide.CNNumCol != 2 {
		t.Fatalf("expected CN side to receive one extra column, got EN=%d CN=%d", slide.ENNumCol, slide.CNNumCol)
	}
}

func TestBuildSlide_ENAndCNCanEachAddOneExtraColumn(t *testing.T) {
	cfg := minimalConfig()
	en := strings.Repeat("long english content. ", 520)
	cn := strings.Repeat("这是比较长的中文内容。", 520)
	slide, _ := BuildSlide(en, cn, cfg)
	if slide.ENNumCol != 2 || slide.CNNumCol != 2 {
		t.Fatalf("expected EN=2 and CN=2, got EN=%d CN=%d", slide.ENNumCol, slide.CNNumCol)
	}
}

func minimalConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Filename.Pattern = `^(\d+)-(\d{3})-(Front|Back)\.md$`
	cfg.Layout.Slide.Width = 13.333
	cfg.Layout.Slide.Height = 7.5
	cfg.Layout.Columns.LeftRatio = 0.5
	cfg.Layout.Columns.Gap = 0.2
	cfg.Layout.Columns.Padding = 0.3
	cfg.Layout.Typography.FontFamily = "Calibri"
	cfg.Layout.Typography.BaseSize = 20
	cfg.Layout.Typography.MinSize = 12
	cfg.Layout.Typography.LineSpacing = 1.2
	cfg.Styles.Markers.Star.Prefix = "★"
	cfg.Styles.Markers.Dot.Prefix = "●"
	cfg.Styles.Markers.Warn.Prefix = "▲"
	cfg.Styles.InlineFormula.Delimiter = "$"
	return cfg
}

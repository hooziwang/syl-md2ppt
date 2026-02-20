package app

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"syl-md2ppt/internal/config"
	"syl-md2ppt/internal/discovery"
	"syl-md2ppt/internal/output"
	"syl-md2ppt/internal/pptx"
	"syl-md2ppt/internal/render"
)

type Options struct {
	SourceDir  string
	OutputArg  string
	ConfigPath string
	CWD        string
	Now        time.Time
	Rand       io.Reader
}

type Result struct {
	OutputPath   string
	SlideCount   int
	WarningCount int
	Warnings     []string
	ConfigSource string
}

func Run(opts Options) (Result, error) {
	if strings.TrimSpace(opts.SourceDir) == "" {
		return Result{}, fmt.Errorf("还没给数据源目录")
	}

	cwd := opts.CWD
	if cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, fmt.Errorf("读取当前目录失败：%w", err)
		}
		cwd = wd
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	rnd := opts.Rand
	if rnd == nil {
		rnd = rand.Reader
	}

	cfg, cfgSrc, err := config.Load(opts.ConfigPath, cwd)
	if err != nil {
		return Result{}, err
	}

	outPath, err := output.ResolveOutputPath(opts.OutputArg, cwd, now, rnd)
	if err != nil {
		return Result{}, err
	}

	pairs, discoverWarn, err := discovery.Discover(opts.SourceDir, cfg, discovery.DiscoverOptions{
		FailOnConflict: true,
	})
	if err != nil {
		return Result{}, err
	}
	if len(pairs) == 0 {
		return Result{}, fmt.Errorf("没找到可用的双语 Markdown 文件，请检查 EN/CN 目录和命名规则")
	}

	slides := make([]render.Slide, 0, len(pairs))
	warnings := make([]string, 0)
	warnings = append(warnings, dedupeStrings(discoverWarn)...)
	for i, pair := range pairs {
		enRaw, err := os.ReadFile(pair.ENPath)
		if err != nil {
			return Result{}, fmt.Errorf("读取英文文件失败（%s）：%w", pair.ENPath, err)
		}
		cnRaw, err := os.ReadFile(pair.CNPath)
		if err != nil {
			return Result{}, fmt.Errorf("读取中文文件失败（%s）：%w", pair.CNPath, err)
		}
		slide, ws := render.BuildSlide(string(enRaw), string(cnRaw), cfg)
		for _, w := range ws {
			warnPath := pair.ENPath
			if w.Code == "truncate_cn" {
				warnPath = pair.CNPath
			}
			warnings = append(warnings, fmt.Sprintf("%s - %s 内容有点多，部分截断", formatSlideNo(i+1), warnPath))
		}
		slides = append(slides, slide)
	}

	deck := pptx.Deck{
		SlideWidthIn:  cfg.Layout.Slide.Width,
		SlideHeightIn: cfg.Layout.Slide.Height,
		LeftRatio:     cfg.Layout.Columns.LeftRatio,
		GapIn:         cfg.Layout.Columns.Gap,
		PaddingIn:     cfg.Layout.Columns.Padding,
		FontFamily:    cfg.Layout.Typography.FontFamily,
		Styles: pptx.StylePalette{
			BaseColor:    "1F2937",
			StarColor:    sanitizeHex(cfg.Styles.Markers.Star.Color, "8A6D1D"),
			DotColor:     sanitizeHex(cfg.Styles.Markers.Dot.Color, "1F2937"),
			WarnColor:    sanitizeHex(cfg.Styles.Markers.Warn.Color, "9A3412"),
			FormulaColor: sanitizeHex(cfg.Styles.InlineFormula.Color, "111827"),
			FormulaFill:  sanitizeHex(cfg.Styles.InlineFormula.Highlight, "FFF176"),
		},
		Slides: slides,
	}

	if err := pptx.Write(outPath, deck); err != nil {
		return Result{}, err
	}

	return Result{
		OutputPath:   outPath,
		SlideCount:   len(slides),
		WarningCount: len(warnings),
		Warnings:     warnings,
		ConfigSource: cfgSrc,
	}, nil
}

func sanitizeHex(v string, fallback string) string {
	v = strings.TrimSpace(strings.TrimPrefix(v, "#"))
	if v == "" {
		return fallback
	}
	v = strings.ToUpper(v)
	if len(v) != 6 {
		return fallback
	}
	for _, ch := range v {
		if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
			return fallback
		}
	}
	return v
}

func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func formatSlideNo(n int) string {
	return fmt.Sprintf("[%3d]", n)
}

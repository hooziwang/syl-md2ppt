package render

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"syl-md2ppt/internal/config"
)

func BuildSlide(enRaw, cnRaw string, cfg *config.Config) (Slide, []Warning) {
	opts := ParseOptions{
		FormulaDelimiter: cfg.Styles.InlineFormula.Delimiter,
		StarPrefix:       cfg.Styles.Markers.Star.Prefix,
		DotPrefix:        cfg.Styles.Markers.Dot.Prefix,
		WarnPrefix:       cfg.Styles.Markers.Warn.Prefix,
	}
	enBlocks := ParseMarkdown(enRaw, opts)
	cnBlocks := ParseMarkdown(cnRaw, opts)

	font := cfg.Layout.Typography.BaseSize
	if font <= 0 {
		font = 20
	}
	minFont := cfg.Layout.Typography.MinSize
	if minFont <= 0 {
		minFont = 12
	}

	enNumCol := 1
	cnNumCol := 1
	for font > minFont {
		if fits(enBlocks, cfg, font, true, enNumCol) && fits(cnBlocks, cfg, font, false, cnNumCol) {
			break
		}
		font--
	}

	enOverflow := overflowLines(enBlocks, cfg, font, true, 1)
	cnOverflow := overflowLines(cnBlocks, cfg, font, false, 1)
	// 两侧独立判断：英文最多加一栏，中文最多加一栏。
	if enOverflow > 0 {
		enNumCol = 2
	}
	if cnOverflow > 0 {
		cnNumCol = 2
	}

	warnings := make([]Warning, 0)
	if !fits(enBlocks, cfg, font, true, enNumCol) {
		var truncated bool
		enBlocks, truncated = truncateToFit(enBlocks, cfg, font, true, enNumCol)
		if truncated {
			warnings = append(warnings, Warning{Code: "truncate_en", Message: "英文内容有点多，部分截断"})
		}
	}
	if !fits(cnBlocks, cfg, font, false, cnNumCol) {
		var truncated bool
		cnBlocks, truncated = truncateToFit(cnBlocks, cfg, font, false, cnNumCol)
		if truncated {
			warnings = append(warnings, Warning{Code: "truncate_cn", Message: "中文内容有点多，部分截断"})
		}
	}

	slide := Slide{
		FontSize:           font,
		ENNumCol:           enNumCol,
		CNNumCol:           cnNumCol,
		HasTruncationBadge: len(warnings) > 0,
		Columns: []Column{
			{Lang: "EN", Blocks: enBlocks},
			{Lang: "CN", Blocks: cnBlocks},
		},
	}
	return slide, warnings
}

func fits(blocks []Block, cfg *config.Config, font int, isLeft bool, numCol int) bool {
	max := maxLines(cfg, font) * max(1, numCol)
	used := usedLines(blocks, cfg, font, isLeft, numCol)
	return used <= max
}

func truncateToFit(blocks []Block, cfg *config.Config, font int, isLeft bool, numCol int) ([]Block, bool) {
	max := maxLines(cfg, font) * max(1, numCol)
	if max <= 0 {
		return blocks, false
	}
	chars := charsPerLine(cfg, font, isLeft, numCol)
	out := make([]Block, 0, len(blocks))
	used := 0
	truncated := false

	for _, block := range blocks {
		text := flattenRuns(block.Runs)
		if text == "" {
			continue
		}
		need := int(math.Ceil(float64(utf8.RuneCountInString(text)) / float64(chars)))
		if need < 1 {
			need = 1
		}

		if used+need <= max {
			out = append(out, block)
			used += need
			continue
		}

		remaining := max - used
		if remaining <= 0 {
			truncated = true
			break
		}

		allowed := remaining * chars
		clipped := clipRunText(block.Runs, allowed)
		if len(clipped) > 0 {
			last := clipped[len(clipped)-1]
			last.Text = strings.TrimSpace(last.Text) + " ..."
			clipped[len(clipped)-1] = last
			out = append(out, Block{Marker: block.Marker, Runs: clipped})
		}
		truncated = true
		break
	}
	if truncated {
		// 给截断页留一个醒目标志，保证页面上可见。
		out = append(out, Block{
			Marker: MarkerWarn,
			Runs: []Run{
				{Text: "【内容有截断】", Bold: true},
			},
		})
	}
	return out, truncated
}

func maxLines(cfg *config.Config, font int) int {
	heightIn := cfg.Layout.Slide.Height - 2*cfg.Layout.Columns.Padding
	if heightIn <= 0 {
		heightIn = 6
	}
	lineSpacing := cfg.Layout.Typography.LineSpacing
	if lineSpacing <= 0 {
		lineSpacing = 1.2
	}
	lineHeightPt := float64(font) * lineSpacing
	max := int(math.Floor(heightIn * 72 / lineHeightPt))
	if max < 1 {
		max = 1
	}
	return max
}

func usedLines(blocks []Block, cfg *config.Config, font int, isLeft bool, numCol int) int {
	chars := charsPerLine(cfg, font, isLeft, numCol)
	total := 0
	for _, block := range blocks {
		text := flattenRuns(block.Runs)
		if text == "" {
			continue
		}
		lines := int(math.Ceil(float64(utf8.RuneCountInString(text)) / float64(chars)))
		if lines < 1 {
			lines = 1
		}
		total += lines
	}
	return total
}

func charsPerLine(cfg *config.Config, font int, isLeft bool, numCol int) int {
	slideWidth := cfg.Layout.Slide.Width
	if slideWidth <= 0 {
		slideWidth = 13.333
	}
	padding := cfg.Layout.Columns.Padding
	gap := cfg.Layout.Columns.Gap
	leftRatio := cfg.Layout.Columns.LeftRatio
	if leftRatio <= 0 || leftRatio >= 1 {
		leftRatio = 0.5
	}
	usable := slideWidth - (2 * padding) - gap
	if usable <= 0 {
		usable = 10
	}
	colWidth := usable * leftRatio
	if !isLeft {
		colWidth = usable * (1 - leftRatio)
	}
	if numCol < 1 {
		numCol = 1
	}
	if numCol > 1 {
		innerGap := gap * 0.5
		if innerGap <= 0 {
			innerGap = 0.08
		}
		colWidth = (colWidth - innerGap*float64(numCol-1)) / float64(numCol)
	}
	if colWidth <= 0 {
		colWidth = 2.5
	}
	chars := int(math.Floor((colWidth * 144.0) / float64(font)))
	if chars < 12 {
		chars = 12
	}
	return chars
}

func flattenRuns(runs []Run) string {
	var b strings.Builder
	for _, r := range runs {
		b.WriteString(r.Text)
	}
	return b.String()
}

func clipRunText(runs []Run, maxRunes int) []Run {
	if maxRunes <= 0 {
		return nil
	}
	out := make([]Run, 0, len(runs))
	remaining := maxRunes
	for _, run := range runs {
		if remaining <= 0 {
			break
		}
		runes := []rune(run.Text)
		if len(runes) <= remaining {
			out = append(out, run)
			remaining -= len(runes)
			continue
		}
		clipped := run
		clipped.Text = string(runes[:remaining])
		out = append(out, clipped)
		remaining = 0
	}
	if len(out) == 0 {
		out = append(out, Run{Text: fmt.Sprintf("%s", "...")})
	}
	return out
}

func overflowLines(blocks []Block, cfg *config.Config, font int, isLeft bool, numCol int) int {
	used := usedLines(blocks, cfg, font, isLeft, numCol)
	capacity := maxLines(cfg, font) * max(1, numCol)
	if used <= capacity {
		return 0
	}
	return used - capacity
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

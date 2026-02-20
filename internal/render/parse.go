package render

import (
	"strings"
)

func ParseMarkdown(raw string, opts ParseOptions) []Block {
	opts = normalizeOptions(opts)
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	blocks := make([]Block, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		marker, text := parseMarker(trimmed, opts)
		runs := parseInline(text, opts.FormulaDelimiter)
		if len(runs) == 0 {
			runs = []Run{{Text: text}}
		}
		blocks = append(blocks, Block{Marker: marker, Runs: runs})
	}
	return blocks
}

func normalizeOptions(opts ParseOptions) ParseOptions {
	if opts.FormulaDelimiter == "" {
		opts.FormulaDelimiter = "$"
	}
	if opts.StarPrefix == "" {
		opts.StarPrefix = "★"
	}
	if opts.DotPrefix == "" {
		opts.DotPrefix = "●"
	}
	if opts.WarnPrefix == "" {
		opts.WarnPrefix = "▲"
	}
	return opts
}

func parseMarker(line string, opts ParseOptions) (MarkerType, string) {
	work := strings.TrimSpace(line)
	work = stripListPrefix(work)

	if strings.HasPrefix(work, opts.StarPrefix) {
		return MarkerStar, strings.TrimSpace(strings.TrimPrefix(work, opts.StarPrefix))
	}
	if strings.HasPrefix(work, opts.DotPrefix) {
		return MarkerDot, strings.TrimSpace(strings.TrimPrefix(work, opts.DotPrefix))
	}
	if strings.HasPrefix(work, opts.WarnPrefix) {
		return MarkerWarn, strings.TrimSpace(strings.TrimPrefix(work, opts.WarnPrefix))
	}

	if strings.HasPrefix(work, "#") {
		return MarkerNormal, strings.TrimSpace(strings.TrimLeft(work, "#"))
	}
	return MarkerNormal, work
}

func stripListPrefix(line string) string {
	work := strings.TrimSpace(line)
	for {
		if strings.HasPrefix(work, "* ") || strings.HasPrefix(work, "- ") || strings.HasPrefix(work, "+ ") {
			work = strings.TrimSpace(work[2:])
			continue
		}
		return work
	}
}

func parseInline(line string, formulaDelimiter string) []Run {
	if line == "" {
		return nil
	}
	var runs []Run
	var b strings.Builder

	bold := false
	italic := false
	formula := false

	flush := func() {
		if b.Len() == 0 {
			return
		}
		runs = append(runs, Run{
			Text:    b.String(),
			Bold:    bold,
			Italic:  italic,
			Formula: formula,
		})
		b.Reset()
	}

	for i := 0; i < len(line); {
		if strings.HasPrefix(line[i:], "**") {
			flush()
			bold = !bold
			i += 2
			continue
		}
		if strings.HasPrefix(line[i:], formulaDelimiter) {
			flush()
			formula = !formula
			i += len(formulaDelimiter)
			continue
		}
		if line[i] == '*' {
			flush()
			italic = !italic
			i++
			continue
		}
		b.WriteByte(line[i])
		i++
	}
	flush()
	return runs
}

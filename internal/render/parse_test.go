package render

import "testing"

func TestParseInlineStyles(t *testing.T) {
	blocks := ParseMarkdown("plain **bold** *it* $x+y$", ParseOptions{})
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	runs := blocks[0].Runs
	if len(runs) < 4 {
		t.Fatalf("expected at least 4 runs, got %d", len(runs))
	}

	var hasBold, hasItalic, hasFormula bool
	for _, r := range runs {
		if r.Bold && r.Text == "bold" {
			hasBold = true
		}
		if r.Italic && r.Text == "it" {
			hasItalic = true
		}
		if r.Formula && r.Text == "x+y" {
			hasFormula = true
		}
	}
	if !hasBold {
		t.Fatalf("bold run not detected")
	}
	if !hasItalic {
		t.Fatalf("italic run not detected")
	}
	if !hasFormula {
		t.Fatalf("formula run not detected")
	}
}

func TestParseMarkerLine(t *testing.T) {
	raw := "★ hero\n● point\n▲ warning"
	blocks := ParseMarkdown(raw, ParseOptions{})
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}
	if blocks[0].Marker != MarkerStar {
		t.Fatalf("expected MarkerStar, got %v", blocks[0].Marker)
	}
	if blocks[1].Marker != MarkerDot {
		t.Fatalf("expected MarkerDot, got %v", blocks[1].Marker)
	}
	if blocks[2].Marker != MarkerWarn {
		t.Fatalf("expected MarkerWarn, got %v", blocks[2].Marker)
	}
}

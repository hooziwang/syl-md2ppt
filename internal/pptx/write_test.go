package pptx

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"syl-md2ppt/internal/render"
)

func TestWritePPTX_MinimalPackage(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "out.pptx")

	deck := Deck{
		SlideWidthIn:  13.333,
		SlideHeightIn: 7.5,
		Slides: []render.Slide{{
			FontSize: 20,
			Columns:  []render.Column{{Lang: "EN", Blocks: []render.Block{{Runs: []render.Run{{Text: "Hello"}}}}}, {Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "你好"}}}}}},
		}},
	}

	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presProps.xml",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/viewProps.xml",
		"ppt/tableStyles.xml",
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels",
		"ppt/slides/slide1.xml",
		"ppt/slides/_rels/slide1.xml.rels",
	}
	for _, want := range required {
		if !zipHasFile(&zr.Reader, want) {
			t.Fatalf("missing %s", want)
		}
	}
}

func TestWritePPTX_AppPropsTitlesConsistent(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "props.pptx")

	deck := Deck{SlideWidthIn: 13.333, SlideHeightIn: 7.5}
	for i := 0; i < 2; i++ {
		deck.Slides = append(deck.Slides, render.Slide{
			FontSize: 20,
			Columns: []render.Column{
				{Lang: "EN", Blocks: []render.Block{{Runs: []render.Run{{Text: "A"}}}}},
				{Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "B"}}}}},
			},
		})
	}
	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var appXML []byte
	for _, f := range zr.File {
		if f.Name == "docProps/app.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open app.xml: %v", err)
			}
			defer rc.Close()
			appXML, err = io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read app.xml: %v", err)
			}
			break
		}
	}
	if len(appXML) == 0 {
		t.Fatalf("docProps/app.xml not found")
	}

	type vector struct {
		Size  int      `xml:"size,attr"`
		Parts []string `xml:"lpstr"`
	}
	type props struct {
		Titles vector `xml:"TitlesOfParts>vector"`
	}
	var p props
	if err := xml.Unmarshal(appXML, &p); err != nil {
		t.Fatalf("unmarshal app.xml: %v", err)
	}
	if p.Titles.Size != len(p.Titles.Parts) {
		t.Fatalf("TitlesOfParts size mismatch: size=%d parts=%d", p.Titles.Size, len(p.Titles.Parts))
	}
}

func TestWritePPTX_SlideCount(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "slides.pptx")

	deck := Deck{SlideWidthIn: 13.333, SlideHeightIn: 7.5}
	for i := 0; i < 3; i++ {
		deck.Slides = append(deck.Slides, render.Slide{
			FontSize: 20,
			Columns:  []render.Column{{Lang: "EN", Blocks: []render.Block{{Runs: []render.Run{{Text: "A"}}}}}, {Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "B"}}}}}},
		})
	}

	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	slideCount := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideCount++
		}
	}
	if slideCount != 3 {
		t.Fatalf("expected 3 slides, got %d", slideCount)
	}
}

func TestWritePPTX_TextboxUsesWrappedLayout(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "wrap.pptx")

	deck := Deck{
		SlideWidthIn:  13.333,
		SlideHeightIn: 7.5,
		Slides: []render.Slide{{
			FontSize: 20,
			Columns: []render.Column{
				{Lang: "EN", Blocks: []render.Block{{Runs: []render.Run{{Text: "This is a very long line that should wrap inside the left column instead of overflowing into the right column."}}}}},
				{Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "这是一段很长的中文文本，应该在右侧栏内自动换行，而不是跑到左侧或和左侧重叠。"}}}}},
			},
		}},
	}

	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var slideXML string
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open slide1.xml: %v", err)
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("read slide1.xml: %v", err)
			}
			slideXML = string(b)
			break
		}
	}
	if slideXML == "" {
		t.Fatalf("ppt/slides/slide1.xml not found")
	}
	if !strings.Contains(slideXML, `wrap="square"`) {
		t.Fatalf("textbox should use wrap=\\\"square\\\" to avoid overflow/overlap")
	}
}

func TestWritePPTX_TruncationBadgeRendered(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "badge.pptx")

	deck := Deck{
		SlideWidthIn:  13.333,
		SlideHeightIn: 7.5,
		Slides: []render.Slide{{
			FontSize:           12,
			ENNumCol:           1,
			CNNumCol:           2,
			HasTruncationBadge: true,
			Columns: []render.Column{
				{Lang: "EN", Blocks: []render.Block{{Runs: []render.Run{{Text: "EN"}}}}},
				{Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "CN"}}}}},
			},
		}},
	}

	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var slideXML string
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open slide1.xml: %v", err)
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("read slide1.xml: %v", err)
			}
			slideXML = string(b)
			break
		}
	}
	if slideXML == "" {
		t.Fatalf("slide1.xml not found")
	}
	if !strings.Contains(slideXML, `numCol="2"`) {
		t.Fatalf("expected multi-column textbox setting in slide xml")
	}
	if !strings.Contains(slideXML, `Truncation Badge`) {
		t.Fatalf("expected truncation badge shape in slide xml")
	}
}

func TestWritePPTX_FormulaKeepsRawAndHighlighted(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "formula.pptx")

	deck := Deck{
		SlideWidthIn:  13.333,
		SlideHeightIn: 7.5,
		Slides: []render.Slide{{
			FontSize: 20,
			Columns: []render.Column{
				{
					Lang: "EN",
					Blocks: []render.Block{{
						Runs: []render.Run{
							{Text: "10^9", Formula: true},
						},
					}},
				},
				{Lang: "CN", Blocks: []render.Block{{Runs: []render.Run{{Text: "测试"}}}}},
			},
		}},
	}

	if err := Write(out, deck); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer zr.Close()

	var slideXML string
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open slide1.xml: %v", err)
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("read slide1.xml: %v", err)
			}
			slideXML = string(b)
			break
		}
	}
	if slideXML == "" {
		t.Fatalf("slide1.xml not found")
	}
	if !strings.Contains(slideXML, `<a:t>$10^9$</a:t>`) {
		t.Fatalf("expected formula raw text with delimiters, got: %s", slideXML)
	}
	if !strings.Contains(slideXML, `<a:highlight>`) {
		t.Fatalf("expected formula highlight in slide xml")
	}
}

func zipHasFile(zr *zip.Reader, name string) bool {
	for _, f := range zr.File {
		if f.Name == name {
			return true
		}
	}
	return false
}

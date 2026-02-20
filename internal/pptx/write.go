package pptx

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"syl-md2ppt/internal/render"
)

const emuPerInch = 914400

//go:embed templates/default.pptx
var defaultTemplate []byte

var (
	xmlDeclRe       = regexp.MustCompile(`^\s*<\?xml[^>]*\?>`)
	slideOverrideRe = regexp.MustCompile(`<Override PartName="/ppt/slides/slide\d+\.xml" ContentType="application/vnd\.openxmlformats-officedocument\.presentationml\.slide\+xml"\s*/>`)
	slideRelRe      = regexp.MustCompile(`<Relationship Id="rId\d+" Type="http://schemas\.openxmlformats\.org/officeDocument/2006/relationships/slide" Target="slides/slide\d+\.xml"\s*/>`)
)

func Write(outPath string, deck Deck) error {
	if outPath == "" {
		return fmt.Errorf("输出路径为空，没法生成 PPT")
	}
	if len(deck.Slides) == 0 {
		return fmt.Errorf("没有可写入的页面，PPT 生成不了")
	}
	applyDeckDefaults(&deck)

	files, err := loadTemplateFiles()
	if err != nil {
		return err
	}

	files["docProps/core.xml"] = []byte(corePropsXML(time.Now().UTC()))
	files["docProps/app.xml"] = []byte(appPropsXML(len(deck.Slides)))
	files["ppt/presentation.xml"] = []byte(presentationXML(string(files["ppt/presentation.xml"]), len(deck.Slides), toEMU(deck.SlideWidthIn), toEMU(deck.SlideHeightIn)))
	files["ppt/_rels/presentation.xml.rels"] = []byte(presentationRelsXML(string(files["ppt/_rels/presentation.xml.rels"]), len(deck.Slides)))
	files["[Content_Types].xml"] = []byte(contentTypesXML(string(files["[Content_Types].xml"]), len(deck.Slides)))

	for i, s := range deck.Slides {
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		relPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", i+1)
		files[slidePath] = []byte(slideXML(s, deck))
		files[relPath] = []byte(slideRelsXML())
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("创建输出目录失败：%w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败：%w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			return fmt.Errorf("写入 PPT 结构失败（%s）：%w", name, err)
		}
		if _, err := w.Write(files[name]); err != nil {
			return fmt.Errorf("写入 PPT 内容失败（%s）：%w", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("写入 PPT 文件收尾失败：%w", err)
	}
	return nil
}

func loadTemplateFiles() (map[string][]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(defaultTemplate), int64(len(defaultTemplate)))
	if err != nil {
		return nil, fmt.Errorf("加载内置 PPT 模板失败：%w", err)
	}
	files := make(map[string][]byte, len(zr.File)+16)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("读取模板文件失败（%s）：%w", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("读取模板内容失败（%s）：%w", f.Name, err)
		}
		files[f.Name] = b
	}
	return files, nil
}

func toEMU(in float64) int64 {
	return int64(in * emuPerInch)
}

func applyDeckDefaults(deck *Deck) {
	if deck.SlideWidthIn <= 0 {
		deck.SlideWidthIn = 13.333
	}
	if deck.SlideHeightIn <= 0 {
		deck.SlideHeightIn = 7.5
	}
	if deck.LeftRatio <= 0 || deck.LeftRatio >= 1 {
		deck.LeftRatio = 0.5
	}
	if deck.GapIn <= 0 {
		deck.GapIn = 0.2
	}
	if deck.PaddingIn <= 0 {
		deck.PaddingIn = 0.3
	}
	if deck.FontFamily == "" {
		deck.FontFamily = "Calibri"
	}
	if deck.Styles.BaseColor == "" {
		deck.Styles.BaseColor = "1F2937"
	}
	if deck.Styles.StarColor == "" {
		deck.Styles.StarColor = "8A6D1D"
	}
	if deck.Styles.DotColor == "" {
		deck.Styles.DotColor = "1F2937"
	}
	if deck.Styles.WarnColor == "" {
		deck.Styles.WarnColor = "9A3412"
	}
	if deck.Styles.FormulaColor == "" {
		deck.Styles.FormulaColor = "F59E0B"
	}
	if deck.Styles.FormulaFill == "" {
		deck.Styles.FormulaFill = "FFF176"
	}
}

func slideXML(slide render.Slide, deck Deck) string {
	pad := toEMU(deck.PaddingIn)
	gap := toEMU(deck.GapIn)
	totalW := toEMU(deck.SlideWidthIn)
	totalH := toEMU(deck.SlideHeightIn)
	usableW := totalW - 2*pad - gap
	leftW := int64(float64(usableW) * deck.LeftRatio)
	rightW := usableW - leftW
	h := totalH - 2*pad

	en := renderColumnXML(slide, 0, pad, pad, leftW, h, 2, deck)
	cn := renderColumnXML(slide, 1, pad+leftW+gap, pad, rightW, h, 3, deck)
	badge := ""
	if slide.HasTruncationBadge {
		badge = truncationBadgeXML(totalW, totalH, pad)
	}

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/>` + en + cn + badge + `</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>`
}

func renderColumnXML(slide render.Slide, colIndex int, x, y, cx, cy int64, shapeID int, deck Deck) string {
	if colIndex >= len(slide.Columns) {
		return ""
	}
	column := slide.Columns[colIndex]
	lang := "en-US"
	if strings.EqualFold(column.Lang, "CN") {
		lang = "zh-CN"
	}

	var paragraphs strings.Builder
	for _, block := range column.Blocks {
		paragraphs.WriteString(paragraphXML(block, slide.FontSize, lang, deck.Styles))
	}
	if paragraphs.Len() == 0 {
		paragraphs.WriteString(`<a:p><a:endParaRPr lang="` + lang + `"/></a:p>`)
	}

	name := "TextBox EN"
	numCol := slide.ENNumCol
	if colIndex == 1 {
		name = "TextBox CN"
		numCol = slide.CNNumCol
	}
	if numCol < 1 {
		numCol = 1
	}
	bodyPr := `<a:bodyPr wrap="square" numCol="` + strconv.Itoa(numCol) + `"><a:spAutoFit/></a:bodyPr>`

	return fmt.Sprintf(`<p:sp><p:nvSpPr><p:cNvPr id="%d" name="%s"/><p:cNvSpPr txBox="1"/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/></p:spPr><p:txBody>%s<a:lstStyle/>%s</p:txBody></p:sp>`, shapeID, name, x, y, cx, cy, bodyPr, paragraphs.String())
}

func paragraphXML(block render.Block, fontSize int, lang string, styles StylePalette) string {
	runs := block.Runs
	if len(runs) == 0 {
		return `<a:p><a:endParaRPr lang="` + lang + `"/></a:p>`
	}

	prefix := ""
	switch block.Marker {
	case render.MarkerStar:
		prefix = "★ "
	case render.MarkerDot:
		prefix = "● "
	case render.MarkerWarn:
		prefix = "▲ "
	}
	if prefix != "" {
		runs = append([]render.Run{{Text: prefix, Bold: block.Marker != render.MarkerDot}}, runs...)
	}

	var b strings.Builder
	b.WriteString(`<a:p>`) // keep minimal to maximize compatibility
	for _, r := range runs {
		rawText := r.Text
		if r.Formula {
			// 公式按原样展示：保留 $...$ 包裹符号。
			rawText = "$" + rawText + "$"
		}
		text := escapeXMLText(rawText)
		if text == "" {
			continue
		}
		color := styles.BaseColor
		highlight := ""
		switch block.Marker {
		case render.MarkerStar:
			color = styles.StarColor
		case render.MarkerWarn:
			color = styles.WarnColor
		case render.MarkerDot:
			color = styles.DotColor
		}
		if r.Formula {
			color = styles.FormulaColor
			highlight = styles.FormulaFill
		}
		b.WriteString(`<a:r><a:rPr lang="` + lang + `" sz="` + strconv.Itoa(fontSize*100) + `"`)
		if r.Bold {
			b.WriteString(` b="1"`)
		}
		if r.Italic {
			b.WriteString(` i="1"`)
		}
		b.WriteString(`><a:solidFill><a:srgbClr val="` + color + `"/></a:solidFill>`)
		if highlight != "" {
			b.WriteString(`<a:highlight><a:srgbClr val="` + highlight + `"/></a:highlight>`)
		}
		b.WriteString(`</a:rPr><a:t>` + text + `</a:t></a:r>`)
	}
	b.WriteString(`<a:endParaRPr lang="` + lang + `"/></a:p>`)
	return b.String()
}

func escapeXMLText(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func contentTypesXML(base string, slideCount int) string {
	clean := slideOverrideRe.ReplaceAllString(base, "")
	idx := strings.LastIndex(clean, "</Types>")
	if idx < 0 {
		return clean
	}
	var overrides strings.Builder
	for i := 1; i <= slideCount; i++ {
		overrides.WriteString(fmt.Sprintf(`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`, i))
	}
	return clean[:idx] + overrides.String() + clean[idx:]
}

func presentationRelsXML(base string, slideCount int) string {
	clean := slideRelRe.ReplaceAllString(base, "")
	idx := strings.LastIndex(clean, "</Relationships>")
	if idx < 0 {
		return clean
	}
	var rels strings.Builder
	for i := 1; i <= slideCount; i++ {
		rels.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`, i+6, i))
	}
	return clean[:idx] + rels.String() + clean[idx:]
}

func presentationXML(base string, slideCount int, cx, cy int64) string {
	noDecl := xmlDeclRe.ReplaceAllString(base, "")
	noDecl = strings.TrimSpace(noDecl)
	openEnd := strings.Index(noDecl, ">")
	if openEnd < 0 {
		return base
	}
	openTag := noDecl[:openEnd+1]
	defaultTextStyle := extractTag(noDecl, "p:defaultTextStyle")
	if defaultTextStyle == "" {
		defaultTextStyle = `<p:defaultTextStyle><a:defPPr><a:defRPr lang="en-US"/></a:defPPr></p:defaultTextStyle>`
	}

	var slideIDs strings.Builder
	if slideCount > 0 {
		slideIDs.WriteString(`<p:sldIdLst>`)
		for i := 1; i <= slideCount; i++ {
			slideIDs.WriteString(fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 255+i, 6+i))
		}
		slideIDs.WriteString(`</p:sldIdLst>`)
	}

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + openTag +
		`<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>` +
		slideIDs.String() +
		`<p:sldSz cx="` + strconv.FormatInt(cx, 10) + `" cy="` + strconv.FormatInt(cy, 10) + `" type="screen16x9"/>` +
		`<p:notesSz cx="6858000" cy="9144000"/>` +
		defaultTextStyle +
		`</p:presentation>`
}

func extractTag(xmlText string, tag string) string {
	start := strings.Index(xmlText, "<"+tag)
	if start < 0 {
		return ""
	}
	endToken := "</" + tag + ">"
	end := strings.Index(xmlText[start:], endToken)
	if end < 0 {
		return ""
	}
	end += start + len(endToken)
	return xmlText[start:end]
}

func slideRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout7.xml"/></Relationships>`
}

func truncationBadgeXML(totalW, totalH, pad int64) string {
	badgeH := toEMU(0.32)
	badgeW := toEMU(3.2)
	x := (totalW - badgeW) / 2
	y := totalH - pad - badgeH
	return fmt.Sprintf(`<p:sp><p:nvSpPr><p:cNvPr id="99" name="Truncation Badge"/><p:cNvSpPr txBox="1"/><p:nvPr/></p:nvSpPr><p:spPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="roundRect"><a:avLst/></a:prstGeom><a:solidFill><a:srgbClr val="FFF3CD"/></a:solidFill><a:ln w="12700"><a:solidFill><a:srgbClr val="DC2626"/></a:solidFill></a:ln></p:spPr><p:txBody><a:bodyPr wrap="square"/><a:lstStyle/><a:p><a:pPr algn="ctr"/><a:r><a:rPr lang="zh-CN" sz="1200" b="1"><a:solidFill><a:srgbClr val="B91C1C"/></a:solidFill></a:rPr><a:t>【本页内容有截断】</a:t></a:r><a:endParaRPr lang="zh-CN"/></a:p></p:txBody></p:sp>`, x, y, badgeW, badgeH)
}

func appPropsXML(slideCount int) string {
	var titles strings.Builder
	titles.WriteString(`<vt:lpstr>Office Theme</vt:lpstr>`)
	for i := 1; i <= slideCount; i++ {
		titles.WriteString(fmt.Sprintf(`<vt:lpstr>幻灯片 %d</vt:lpstr>`, i))
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><TotalTime>1</TotalTime><Words>0</Words><Application>syl-md2ppt</Application><PresentationFormat>On-screen Show (16:9)</PresentationFormat><Paragraphs>0</Paragraphs><Slides>%d</Slides><Notes>0</Notes><HiddenSlides>0</HiddenSlides><MMClips>0</MMClips><ScaleCrop>false</ScaleCrop><HeadingPairs><vt:vector size="4" baseType="variant"><vt:variant><vt:lpstr>Theme</vt:lpstr></vt:variant><vt:variant><vt:i4>1</vt:i4></vt:variant><vt:variant><vt:lpstr>Slide Titles</vt:lpstr></vt:variant><vt:variant><vt:i4>%d</vt:i4></vt:variant></vt:vector></HeadingPairs><TitlesOfParts><vt:vector size="%d" baseType="lpstr">%s</vt:vector></TitlesOfParts><Manager></Manager><Company></Company><LinksUpToDate>false</LinksUpToDate><SharedDoc>false</SharedDoc><HyperlinkBase></HyperlinkBase><HyperlinksChanged>false</HyperlinksChanged><AppVersion>16.0000</AppVersion></Properties>`, slideCount, slideCount, slideCount+1, titles.String())
}

func corePropsXML(now time.Time) string {
	t := now.Format("2006-01-02T15:04:05Z")
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>syl-md2ppt</dc:title>
  <dc:creator>syl-md2ppt</dc:creator>
  <cp:lastModifiedBy>syl-md2ppt</cp:lastModifiedBy>
  <dcterms:created xsi:type="dcterms:W3CDTF">` + t + `</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">` + t + `</dcterms:modified>
</cp:coreProperties>`
}

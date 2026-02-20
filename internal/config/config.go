package config

type Config struct {
	Filename FilenameConfig `yaml:"filename"`
	Layout   LayoutConfig   `yaml:"layout"`
	Styles   StylesConfig   `yaml:"styles"`
	Output   OutputConfig   `yaml:"output"`
}

type FilenameConfig struct {
	Pattern         string         `yaml:"pattern"`
	Groups          FilenameGroups `yaml:"groups"`
	Order           FilenameOrder  `yaml:"order"`
	IgnoreUnmatched bool           `yaml:"ignore_unmatched"`
}

type FilenameGroups struct {
	Domain int `yaml:"domain"`
	Card   int `yaml:"card"`
	Side   int `yaml:"side"`
}

type FilenameOrder struct {
	Side []string `yaml:"side"`
}

type LayoutConfig struct {
	Slide      SlideConfig      `yaml:"slide"`
	Columns    ColumnsConfig    `yaml:"columns"`
	Typography TypographyConfig `yaml:"typography"`
}

type SlideConfig struct {
	Width  float64 `yaml:"width"`
	Height float64 `yaml:"height"`
	Unit   string  `yaml:"unit"`
}

type ColumnsConfig struct {
	LeftRatio float64 `yaml:"left_ratio"`
	Gap       float64 `yaml:"gap"`
	Padding   float64 `yaml:"padding"`
}

type TypographyConfig struct {
	FontFamily  string  `yaml:"font_family"`
	BaseSize    int     `yaml:"base_size"`
	MinSize     int     `yaml:"min_size"`
	LineSpacing float64 `yaml:"line_spacing"`
}

type StylesConfig struct {
	Markers       MarkerSetConfig    `yaml:"markers"`
	InlineFormula InlineFormulaStyle `yaml:"inline_formula"`
}

type MarkerSetConfig struct {
	Star MarkerStyle `yaml:"star"`
	Dot  MarkerStyle `yaml:"dot"`
	Warn MarkerStyle `yaml:"warn"`
}

type MarkerStyle struct {
	Prefix    string `yaml:"prefix"`
	AccentBar bool   `yaml:"accent_bar"`
	Color     string `yaml:"color"`
	Highlight string `yaml:"highlight"`
}

type InlineFormulaStyle struct {
	Delimiter string `yaml:"delimiter"`
	Highlight string `yaml:"highlight"`
	Color     string `yaml:"color"`
}

type OutputConfig struct {
	DefaultName DefaultNameConfig `yaml:"default_name"`
}

type DefaultNameConfig struct {
	TimestampFormat string `yaml:"timestamp_format"`
	RandomSuffixLen int    `yaml:"random_suffix_len"`
}

func (c *Config) applyDefaults() {
	if c.Filename.Pattern == "" {
		c.Filename.Pattern = `^(\d+)-(\d{3})-(Front|Back)\.md$`
	}
	if c.Filename.Groups.Domain <= 0 {
		c.Filename.Groups.Domain = 1
	}
	if c.Filename.Groups.Card <= 0 {
		c.Filename.Groups.Card = 2
	}
	if c.Filename.Groups.Side <= 0 {
		c.Filename.Groups.Side = 3
	}
	if len(c.Filename.Order.Side) == 0 {
		c.Filename.Order.Side = []string{"Front", "Back"}
	}
	if c.Layout.Slide.Width == 0 {
		c.Layout.Slide.Width = 13.333
	}
	if c.Layout.Slide.Height == 0 {
		c.Layout.Slide.Height = 7.5
	}
	if c.Layout.Slide.Unit == "" {
		c.Layout.Slide.Unit = "in"
	}
	if c.Layout.Columns.LeftRatio <= 0 || c.Layout.Columns.LeftRatio >= 1 {
		c.Layout.Columns.LeftRatio = 0.5
	}
	if c.Layout.Columns.Gap <= 0 {
		c.Layout.Columns.Gap = 0.2
	}
	if c.Layout.Columns.Padding <= 0 {
		c.Layout.Columns.Padding = 0.3
	}
	if c.Layout.Typography.FontFamily == "" {
		c.Layout.Typography.FontFamily = "Calibri"
	}
	if c.Layout.Typography.BaseSize <= 0 {
		c.Layout.Typography.BaseSize = 20
	}
	if c.Layout.Typography.MinSize <= 0 {
		c.Layout.Typography.MinSize = 12
	}
	if c.Layout.Typography.LineSpacing <= 0 {
		c.Layout.Typography.LineSpacing = 1.2
	}
	if c.Styles.Markers.Star.Prefix == "" {
		c.Styles.Markers.Star.Prefix = "★"
	}
	if c.Styles.Markers.Dot.Prefix == "" {
		c.Styles.Markers.Dot.Prefix = "●"
	}
	if c.Styles.Markers.Warn.Prefix == "" {
		c.Styles.Markers.Warn.Prefix = "▲"
	}
	if c.Styles.InlineFormula.Delimiter == "" {
		c.Styles.InlineFormula.Delimiter = "$"
	}
	if c.Output.DefaultName.TimestampFormat == "" {
		c.Output.DefaultName.TimestampFormat = "20060102_150405"
	}
	if c.Output.DefaultName.RandomSuffixLen <= 0 {
		c.Output.DefaultName.RandomSuffixLen = 6
	}
}

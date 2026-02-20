package pptx

import "syl-md2ppt/internal/render"

type Deck struct {
	SlideWidthIn  float64
	SlideHeightIn float64
	LeftRatio     float64
	GapIn         float64
	PaddingIn     float64
	FontFamily    string
	Styles        StylePalette
	Slides        []render.Slide
}

type StylePalette struct {
	BaseColor    string
	StarColor    string
	DotColor     string
	WarnColor    string
	FormulaColor string
	FormulaFill  string
}

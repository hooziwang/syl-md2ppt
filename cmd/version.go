package cmd

import (
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func versionText() string {
	return fmt.Sprintf("syl-md2ppt 版本：%s（commit: %s，构建时间: %s）", Version, Commit, BuildTime)
}

type rgb struct {
	r int
	g int
	b int
}

func lerpColor(a, b, t float64) int {
	return int(a + (b-a)*t)
}

func hueToRGB(h float64) rgb {
	s := 0.78
	l := 0.62
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2

	r1, g1, b1 := 0.0, 0.0, 0.0
	switch {
	case h < 60:
		r1, g1, b1 = c, x, 0
	case h < 120:
		r1, g1, b1 = x, c, 0
	case h < 180:
		r1, g1, b1 = 0, c, x
	case h < 240:
		r1, g1, b1 = 0, x, c
	case h < 300:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}

	return rgb{
		r: int(math.Round((r1 + m) * 255)),
		g: int(math.Round((g1 + m) * 255)),
		b: int(math.Round((b1 + m) * 255)),
	}
}

func randomComplementaryColors() (rgb, rgb) {
	n, err := rand.Int(rand.Reader, big.NewInt(360))
	if err != nil {
		return rgb{r: 90, g: 150, b: 255}, rgb{r: 255, g: 120, b: 190}
	}
	baseHue := float64(n.Int64())
	return hueToRGB(baseHue), hueToRGB(math.Mod(baseHue+180, 360))
}

func gradientPaint(line string, start rgb, end rgb) string {
	runes := []rune(line)
	if len(runes) == 0 {
		return ""
	}
	var b strings.Builder
	max := len(runes) - 1
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(' ')
			continue
		}
		t := 0.0
		if max > 0 {
			t = float64(i) / float64(max)
		}
		rv := lerpColor(float64(start.r), float64(end.r), t)
		gv := lerpColor(float64(start.g), float64(end.g), t)
		bv := lerpColor(float64(start.b), float64(end.b), t)
		b.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", rv, gv, bv, r))
	}
	return b.String()
}

func loveBanner() string {
	start, end := randomComplementaryColors()
	return gradientPaint("DADDYLOVESYL", start, end)
}

func printVersion(w io.Writer) {
	fmt.Fprintln(w, versionText())
	fmt.Fprintln(w, loveBanner())
}

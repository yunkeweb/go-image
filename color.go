package goimage

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// Color is an 8-bit-per-channel sRGB color with alpha (255 = opaque).
type Color struct {
	R, G, B, A uint8
}

// ColorspaceName identifies the working colorspace of an image.
type ColorspaceName string

const (
	ColorspaceRGB  ColorspaceName = "rgb"
	ColorspaceCMYK ColorspaceName = "cmyk"
)

var (
	ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
	ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
	ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)

// ParseColor accepts hex, rgb()/rgba(), HTML names, "transparent", or Color.
func ParseColor(v any) (Color, error) {
	switch c := v.(type) {
	case Color:
		return c, nil
	case *Color:
		if c == nil {
			return Color{}, wrap(ErrColor, "nil color")
		}
		return *c, nil
	case color.Color:
		r, g, b, a := c.RGBA()
		return Color{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}, nil
	case string:
		return parseColorString(c)
	default:
		return Color{}, wrap(ErrColor, "unable to decode color from %T", v)
	}
}

func mustColor(v any) Color {
	c, err := ParseColor(v)
	if err != nil {
		return ColorWhite
	}
	return c
}

func parseColorString(s string) (Color, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Color{}, wrap(ErrColor, "empty color string")
	}
	lower := strings.ToLower(s)
	if lower == "transparent" {
		return ColorTransparent, nil
	}
	if hex, ok := htmlColorNames[lower]; ok {
		return parseHex(hex)
	}
	if strings.HasPrefix(lower, "rgb") || strings.HasPrefix(lower, "srgb") {
		return parseRGBFunc(s)
	}
	return parseHex(s)
}

func parseHex(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSpace(s)
	if s == "" {
		return Color{}, wrap(ErrColor, "unable to decode hex color")
	}
	expand := func(ch byte) uint8 {
		v, err := strconv.ParseUint(string([]byte{ch, ch}), 16, 8)
		if err != nil {
			return 0
		}
		return uint8(v)
	}
	pair := func(i int) (uint8, error) {
		v, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return 0, wrap(ErrColor, "unable to decode hex color")
		}
		return uint8(v), nil
	}
	c := Color{A: 255}
	switch len(s) {
	case 3:
		c.R, c.G, c.B = expand(s[0]), expand(s[1]), expand(s[2])
	case 4:
		c.R, c.G, c.B, c.A = expand(s[0]), expand(s[1]), expand(s[2]), expand(s[3])
	case 6:
		var err error
		if c.R, err = pair(0); err != nil {
			return Color{}, err
		}
		if c.G, err = pair(2); err != nil {
			return Color{}, err
		}
		if c.B, err = pair(4); err != nil {
			return Color{}, err
		}
	case 8:
		var err error
		if c.R, err = pair(0); err != nil {
			return Color{}, err
		}
		if c.G, err = pair(2); err != nil {
			return Color{}, err
		}
		if c.B, err = pair(4); err != nil {
			return Color{}, err
		}
		if c.A, err = pair(6); err != nil {
			return Color{}, err
		}
	default:
		return Color{}, wrap(ErrColor, "unable to decode hex color %q", s)
	}
	return c, nil
}

func parseRGBFunc(s string) (Color, error) {
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end <= start {
		return Color{}, wrap(ErrColor, "unable to decode rgb color")
	}
	parts := strings.Split(s[start+1:end], ",")
	if len(parts) < 3 || len(parts) > 4 {
		return Color{}, wrap(ErrColor, "unable to decode rgb color")
	}
	ch := func(p string) (uint8, error) {
		p = strings.TrimSpace(p)
		if strings.HasSuffix(p, "%") {
			f, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
			if err != nil {
				return 0, wrap(ErrColor, "unable to decode rgb color")
			}
			return clamp8(int(math.Round(f / 100 * 255))), nil
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0, wrap(ErrColor, "unable to decode rgb color")
		}
		return clamp8(int(math.Round(f))), nil
	}
	r, err := ch(parts[0])
	if err != nil {
		return Color{}, err
	}
	g, err := ch(parts[1])
	if err != nil {
		return Color{}, err
	}
	b, err := ch(parts[2])
	if err != nil {
		return Color{}, err
	}
	c := Color{R: r, G: g, B: b, A: 255}
	if len(parts) == 4 {
		p := strings.TrimSpace(parts[3])
		if strings.HasSuffix(p, "%") {
			f, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
			if err != nil {
				return Color{}, wrap(ErrColor, "unable to decode rgb color")
			}
			c.A = clamp8(int(math.Round(f / 100 * 255)))
		} else {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return Color{}, wrap(ErrColor, "unable to decode rgb color")
			}
			if f <= 1 {
				c.A = clamp8(int(math.Round(f * 255)))
			} else {
				c.A = clamp8(int(math.Round(f)))
			}
		}
	}
	return c, nil
}

func clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func (c Color) RGBA() (r, g, b, a uint32) {
	return c.NRGBA().RGBA()
}

func colorFromNRGBA(n color.NRGBA) Color {
	return Color{R: n.R, G: n.G, B: n.B, A: n.A}
}

func (c Color) IsTransparent() bool { return c.A < 255 }
func (c Color) IsClear() bool       { return c.A == 0 }
func (c Color) IsGreyscale() bool   { return c.R == c.G && c.G == c.B }

func (c Color) ToHex(prefix string) string {
	if c.IsTransparent() {
		return fmt.Sprintf("%s%02x%02x%02x%02x", prefix, c.R, c.G, c.B, c.A)
	}
	return fmt.Sprintf("%s%02x%02x%02x", prefix, c.R, c.G, c.B)
}

func (c Color) String() string {
	if c.IsTransparent() {
		return fmt.Sprintf("rgba(%d, %d, %d, %.1f)", c.R, c.G, c.B, float64(c.A)/255)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

// HSL returns hue in degrees [0,360), saturation and luminance in [0,1].
func (c Color) HSL() (h, s, l float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	return h, s, l
}

// HSV returns hue in degrees [0,360), saturation and value in [0,1].
func (c Color) HSV() (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	v = max
	d := max - min
	if max == 0 {
		return 0, 0, v
	}
	s = d / max
	if d == 0 {
		return 0, s, v
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	return h, s, v
}

// CMYK returns C,M,Y,K in [0,1].
func (c Color) CMYK() (cyan, magenta, yellow, key float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	key = 1 - math.Max(r, math.Max(g, b))
	if key >= 1 {
		return 0, 0, 0, 1
	}
	cyan = (1 - r - key) / (1 - key)
	magenta = (1 - g - key) / (1 - key)
	yellow = (1 - b - key) / (1 - key)
	return cyan, magenta, yellow, key
}

func ColorFromHSL(h, s, l float64, a uint8) Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = clamp01(s)
	l = clamp01(l)
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return Color{
		R: clamp8(int(math.Round((r + m) * 255))),
		G: clamp8(int(math.Round((g + m) * 255))),
		B: clamp8(int(math.Round((b + m) * 255))),
		A: a,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func overNRGBA(dst, src color.NRGBA) color.NRGBA {
	if src.A == 0 {
		return dst
	}
	if src.A == 255 {
		return src
	}
	sa := float64(src.A) / 255
	da := float64(dst.A) / 255
	outA := sa + da*(1-sa)
	if outA == 0 {
		return color.NRGBA{}
	}
	r := (float64(src.R)*sa + float64(dst.R)*da*(1-sa)) / outA
	g := (float64(src.G)*sa + float64(dst.G)*da*(1-sa)) / outA
	b := (float64(src.B)*sa + float64(dst.B)*da*(1-sa)) / outA
	return color.NRGBA{
		R: clamp8(int(math.Round(r))),
		G: clamp8(int(math.Round(g))),
		B: clamp8(int(math.Round(b))),
		A: clamp8(int(math.Round(outA * 255))),
	}
}

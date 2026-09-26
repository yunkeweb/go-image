package modifier

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/yunkeweb/go-image/internal/errs"
	"github.com/yunkeweb/go-image/internal/pool"
)

// TextStyle describes a DrawText call.
type TextStyle struct {
	Filename    string
	Size        float64
	Angle       float64
	Color       color.NRGBA
	StrokeColor color.NRGBA
	StrokeWidth int
	Align       string
	Valign      string
	LineHeight  float64
	WrapWidth   int
}

// DrawText paints text onto a clone of n.
func DrawText(n *image.NRGBA, text string, x, y int, style TextStyle) (*image.NRGBA, error) {
	face, closer, err := LoadFace(style.Filename, style.Size)
	if err != nil {
		return nil, err
	}
	if closer != nil {
		defer closer()
	}
	measure := func(s string) int {
		return font.MeasureString(face, s).Round()
	}
	lines := WrapText(text, style.WrapWidth, measure)
	lineH := int(math.Round(float64(face.Metrics().Height.Round()) * style.LineHeight))
	if lineH < 1 {
		lineH = face.Metrics().Height.Round()
	}
	dst := pool.Clone(n)
	for i, line := range lines {
		w := measure(line)
		px, py := x, y+i*lineH
		switch style.Align {
		case "center":
			px -= w / 2
		case "right":
			px -= w
		}
		ascent := face.Metrics().Ascent.Round()
		switch style.Valign {
		case "top":
			py += ascent
		case "middle", "center":
			py += ascent / 2
		}
		if style.Angle != 0 {
			drawStringRotated(dst, face, px, py, line, style.Color, style.StrokeColor, style.StrokeWidth, style.Angle)
		} else {
			drawString(dst, face, px, py, line, style.Color, style.StrokeColor, style.StrokeWidth)
		}
	}
	return dst, nil
}

// LoadFace opens a TTF/OTF file, or returns basicfont when filename is empty.
func LoadFace(filename string, size float64) (font.Face, func(), error) {
	if filename != "" {
		st, err := os.Stat(filename)
		if err == nil && !st.IsDir() {
			data, err := os.ReadFile(filename)
			if err != nil {
				return nil, nil, errs.Wrap(errs.ErrFont, "unable to read font: %v", err)
			}
			tt, err := opentype.Parse(data)
			if err != nil {
				return nil, nil, errs.Wrap(errs.ErrFont, "unable to parse font: %v", err)
			}
			if size <= 0 {
				size = 12
			}
			face, err := opentype.NewFace(tt, &opentype.FaceOptions{
				Size:    size,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			if err != nil {
				return nil, nil, errs.Wrap(errs.ErrFont, "unable to create font face: %v", err)
			}
			return face, func() {
				if c, ok := face.(interface{ Close() error }); ok {
					_ = c.Close()
				}
			}, nil
		}
	}
	return basicfont.Face7x13, nil, nil
}

// WrapText splits text into lines that fit maxWidth according to measure.
func WrapText(text string, maxWidth int, measure func(string) int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	raw := strings.Split(text, "\n")
	if maxWidth <= 0 {
		return raw
	}
	var lines []string
	for _, para := range raw {
		if para == "" {
			lines = append(lines, "")
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		cur := words[0]
		for _, w := range words[1:] {
			trial := cur + " " + w
			if measure(trial) <= maxWidth {
				cur = trial
			} else {
				lines = append(lines, cur)
				cur = w
			}
		}
		lines = append(lines, cur)
	}
	return lines
}

func drawString(dst *image.NRGBA, face font.Face, x, y int, s string, col, stroke color.NRGBA, strokeW int) {
	if strokeW > 0 {
		for dy := -strokeW; dy <= strokeW; dy++ {
			for dx := -strokeW; dx <= strokeW; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				drawer := &font.Drawer{
					Dst:  dst,
					Src:  image.NewUniform(stroke),
					Face: face,
					Dot:  fixed.P(x+dx, y+dy),
				}
				drawer.DrawString(s)
			}
		}
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(s)
}

func drawStringRotated(dst *image.NRGBA, face font.Face, x, y int, s string, col, stroke color.NRGBA, strokeW int, angle float64) {
	w := font.MeasureString(face, s).Round()
	h := face.Metrics().Height.Round() + 4
	pad := strokeW + 2
	tmp := pool.Acquire(w+pad*2, h+pad*2)
	drawString(tmp, face, pad, pad+face.Metrics().Ascent.Round(), s, col, stroke, strokeW)
	rotated := Rotate(tmp, angle, color.NRGBA{})
	pt := image.Pt(x-pad, y-pad-face.Metrics().Ascent.Round())
	draw.Draw(dst, rotated.Bounds().Add(pt), rotated, rotated.Bounds().Min, draw.Over)
	if rotated != tmp {
		pool.Release(rotated)
	}
	pool.Release(tmp)
}

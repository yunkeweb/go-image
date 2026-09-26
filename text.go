package goimage

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func (img *Image) Text(text string, x, y int, fontInit any) *Image {
	if img.fail() {
		return img
	}
	fnt, err := applyFontInit(fontInit)
	if err != nil {
		return img.setErr(err)
	}
	face, closer, err := loadFace(fnt)
	if err != nil {
		return img.setErr(err)
	}
	if closer != nil {
		defer closer()
	}
	col, err := ParseColor(fnt.color)
	if err != nil {
		col = ColorBlack
	}
	strokeCol := ColorWhite
	if fnt.strokeWidth > 0 {
		if sc, e := ParseColor(fnt.strokeColor); e == nil {
			strokeCol = sc
		}
	}
	measure := func(s string) int {
		return font.MeasureString(face, s).Round()
	}
	lines := wrapText(text, fnt.wrapWidth, measure)
	lineH := int(math.Round(float64(face.Metrics().Height.Round()) * fnt.lineHeight))
	if lineH < 1 {
		lineH = face.Metrics().Height.Round()
	}

	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		for i, line := range lines {
			w := measure(line)
			px, py := x, y+i*lineH
			switch fnt.align {
			case "center":
				px -= w / 2
			case "right":
				px -= w
			}
			ascent := face.Metrics().Ascent.Round()
			switch fnt.valign {
			case "top":
				py += ascent
			case "middle", "center":
				py += ascent / 2
			case "bottom":
				// baseline at y (PHP default)
			}
			if fnt.angle != 0 {
				drawStringRotated(dst, face, px, py, line, col, strokeCol, fnt.strokeWidth, fnt.angle)
			} else {
				drawString(dst, face, px, py, line, col, strokeCol, fnt.strokeWidth)
			}
		}
		return dst, nil
	})
}

func loadFace(f *Font) (font.Face, func(), error) {
	if f.hasFile() {
		data, err := os.ReadFile(f.filename)
		if err != nil {
			return nil, nil, wrap(ErrFont, "unable to read font: %v", err)
		}
		tt, err := opentype.Parse(data)
		if err != nil {
			return nil, nil, wrap(ErrFont, "unable to parse font: %v", err)
		}
		size := f.size
		if size <= 0 {
			size = 12
		}
		face, err := opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			return nil, nil, wrap(ErrFont, "unable to create font face: %v", err)
		}
		return face, func() {
			if c, ok := face.(interface{ Close() error }); ok {
				_ = c.Close()
			}
		}, nil
	}
	return basicfont.Face7x13, nil, nil
}

func drawString(dst *image.NRGBA, face font.Face, x, y int, s string, col, stroke Color, strokeW int) {
	if strokeW > 0 {
		for dy := -strokeW; dy <= strokeW; dy++ {
			for dx := -strokeW; dx <= strokeW; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				drawer := &font.Drawer{
					Dst:  dst,
					Src:  image.NewUniform(stroke.NRGBA()),
					Face: face,
					Dot:  fixed.P(x+dx, y+dy),
				}
				drawer.DrawString(s)
			}
		}
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col.NRGBA()),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(s)
}

func drawStringRotated(dst *image.NRGBA, face font.Face, x, y int, s string, col, stroke Color, strokeW int, angle float64) {
	w := font.MeasureString(face, s).Round()
	h := face.Metrics().Height.Round() + 4
	pad := strokeW + 2
	tmp := acquireNRGBA(w+pad*2, h+pad*2)
	drawString(tmp, face, pad, pad+face.Metrics().Ascent.Round(), s, col, stroke, strokeW)
	rotated := rotateNRGBA(tmp, angle, color.NRGBA{})
	pt := image.Pt(x-pad, y-pad-face.Metrics().Ascent.Round())
	draw.Draw(dst, rotated.Bounds().Add(pt), rotated, rotated.Bounds().Min, draw.Over)
	if rotated != tmp {
		releaseNRGBA(rotated)
	}
	releaseNRGBA(tmp)
}

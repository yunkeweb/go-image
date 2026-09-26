package encoder

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/yunkeweb/go-image/internal/pool"
)

const maxGIFColors = 256

// QuantizePaletted copies src into a paletted image using MedianCutPalette.
func QuantizePaletted(src *image.NRGBA, limit int) *image.Paletted {
	if src == nil {
		return image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.NRGBA{A: 0}})
	}
	b := src.Bounds()
	pal := MedianCutPalette(src, limit)
	p := image.NewPaletted(b, pal)
	draw.Draw(p, b, src, b.Min, draw.Src)
	return p
}

// MedianCutPalette builds a GIF-safe palette of at most 256 colors.
// A transparent slot is reserved only when src contains a pixel with A < 255.
func MedianCutPalette(src *image.NRGBA, limit int) color.Palette {
	if limit < 1 {
		limit = 1
	}
	if limit > maxGIFColors {
		limit = maxGIFColors
	}
	if src == nil {
		return color.Palette{color.NRGBA{A: 0}}
	}
	hasTransparent := pool.HasTransparency(src)
	colorLimit := limit
	if hasTransparent && colorLimit > 1 {
		colorLimit--
	}

	b := src.Bounds()
	type pix struct{ r, g, bl, a uint8 }
	seen := map[uint32]pix{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			if hasTransparent && c.A == 0 {
				continue
			}
			key := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
			if _, ok := seen[key]; !ok {
				seen[key] = pix{c.R, c.G, c.B, c.A}
			}
		}
	}
	pts := make([]pix, 0, len(seen))
	for _, p := range seen {
		pts = append(pts, p)
	}
	if len(pts) == 0 {
		return color.Palette{color.NRGBA{A: 0}}
	}

	build := func(colors []pix) color.Palette {
		pal := make(color.Palette, 0, len(colors)+1)
		if hasTransparent {
			pal = append(pal, color.NRGBA{A: 0})
		}
		for _, p := range colors {
			pal = append(pal, color.NRGBA{R: p.r, G: p.g, B: p.bl, A: p.a})
		}
		if len(pal) > maxGIFColors {
			pal = pal[:maxGIFColors]
		}
		if len(pal) == 0 {
			return color.Palette{color.NRGBA{A: 0}}
		}
		return pal
	}

	if len(pts) <= colorLimit {
		return build(pts)
	}

	boxes := [][]pix{pts}
	for len(boxes) < colorLimit {
		bi, channel, spread := -1, 0, 0
		for i, box := range boxes {
			if len(box) < 2 {
				continue
			}
			minR, maxR := 255, 0
			minG, maxG := 255, 0
			minB, maxB := 255, 0
			for _, p := range box {
				if int(p.r) < minR {
					minR = int(p.r)
				}
				if int(p.r) > maxR {
					maxR = int(p.r)
				}
				if int(p.g) < minG {
					minG = int(p.g)
				}
				if int(p.g) > maxG {
					maxG = int(p.g)
				}
				if int(p.bl) < minB {
					minB = int(p.bl)
				}
				if int(p.bl) > maxB {
					maxB = int(p.bl)
				}
			}
			ranges := []int{maxR - minR, maxG - minG, maxB - minB}
			ch, sp := 0, ranges[0]
			for c, r := range ranges {
				if r > sp {
					ch, sp = c, r
				}
			}
			if sp > spread {
				bi, channel, spread = i, ch, sp
			}
		}
		if bi < 0 {
			break
		}
		box := boxes[bi]
		val := func(p pix) int {
			switch channel {
			case 0:
				return int(p.r)
			case 1:
				return int(p.g)
			default:
				return int(p.bl)
			}
		}
		for i := 1; i < len(box); i++ {
			j := i
			for j > 0 && val(box[j-1]) > val(box[j]) {
				box[j-1], box[j] = box[j], box[j-1]
				j--
			}
		}
		mid := len(box) / 2
		if mid == 0 {
			mid = 1
		}
		left, right := box[:mid], box[mid:]
		boxes[bi] = left
		boxes = append(boxes, right)
	}

	reduced := make([]pix, 0, len(boxes))
	for _, box := range boxes {
		var r, g, bl, a, n int
		for _, p := range box {
			r += int(p.r)
			g += int(p.g)
			bl += int(p.bl)
			a += int(p.a)
			n++
		}
		if n == 0 {
			continue
		}
		reduced = append(reduced, pix{uint8(r / n), uint8(g / n), uint8(bl / n), uint8(a / n)})
	}
	return build(reduced)
}

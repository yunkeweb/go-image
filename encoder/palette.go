package encoder

import (
	"image"
	"image/color"
	"image/draw"
)

func QuantizePaletted(src *image.NRGBA, limit int) *image.Paletted {
	b := src.Bounds()
	pal := MedianCutPalette(src, limit)
	p := image.NewPaletted(b, pal)
	draw.Draw(p, b, src, b.Min, draw.Src)
	return p
}

func MedianCutPalette(src *image.NRGBA, limit int) color.Palette {
	if limit < 2 {
		limit = 2
	}
	b := src.Bounds()
	type pix struct{ r, g, bl, a uint8 }
	seen := map[uint32]pix{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
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
	if len(pts) <= limit {
		pal := make(color.Palette, 0, len(pts)+1)
		pal = append(pal, color.NRGBA{A: 0})
		for _, p := range pts {
			pal = append(pal, color.NRGBA{R: p.r, G: p.g, B: p.bl, A: p.a})
		}
		return pal
	}
	boxes := [][]pix{pts}
	for len(boxes) < limit {
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
	pal := color.Palette{color.NRGBA{A: 0}}
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
		pal = append(pal, color.NRGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(bl / n), A: uint8(a / n)})
	}
	return pal
}

package goimage

import (
	"image"
	"image/color"
	"math"
)

func (img *Image) Greyscale() *Image {
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		y := uint8((int(c.R)*299 + int(c.G)*587 + int(c.B)*114) / 1000)
		return color.NRGBA{R: y, G: y, B: y, A: c.A}
	})
}

func (img *Image) Invert() *Image {
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A}
	})
}

func (img *Image) Brightness(level int) *Image {
	delta := int(math.Round(float64(level) * 2.55))
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{
			R: clamp8(int(c.R) + delta),
			G: clamp8(int(c.G) + delta),
			B: clamp8(int(c.B) + delta),
			A: c.A,
		}
	})
}

func (img *Image) Contrast(level int) *Image {
	// PHP GD IMG_FILTER_CONTRAST: factor around (100-level)/100 with a tan curve.
	c := clampInt(level, -100, 100)
	factor := (259.0 * (float64(c) + 255)) / (255 * (259 - float64(c)))
	return img.mapPixels(func(px color.NRGBA) color.NRGBA {
		adj := func(v uint8) uint8 {
			x := factor*(float64(v)-128) + 128
			return clamp8(int(math.Round(x)))
		}
		return color.NRGBA{R: adj(px.R), G: adj(px.G), B: adj(px.B), A: px.A}
	})
}

func (img *Image) Gamma(gamma float64) *Image {
	if gamma <= 0 {
		return img.setErr(wrap(ErrInput, "gamma must be > 0"))
	}
	inv := 1 / gamma
	lut := [256]uint8{}
	for i := 0; i < 256; i++ {
		lut[i] = clamp8(int(math.Round(math.Pow(float64(i)/255, inv) * 255)))
	}
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{R: lut[c.R], G: lut[c.G], B: lut[c.B], A: c.A}
	})
}

func (img *Image) Colorize(red, green, blue int) *Image {
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{
			R: clamp8(int(c.R) + red),
			G: clamp8(int(c.G) + green),
			B: clamp8(int(c.B) + blue),
			A: c.A,
		}
	})
}

func (img *Image) Pixelate(size int) *Image {
	if size < 1 {
		size = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		b := n.Bounds()
		dst := image.NewNRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y += size {
			for x := b.Min.X; x < b.Max.X; x += size {
				x2 := min(x+size, b.Max.X)
				y2 := min(y+size, b.Max.Y)
				var r, g, bl, a, cnt int
				for yy := y; yy < y2; yy++ {
					for xx := x; xx < x2; xx++ {
						c := n.NRGBAAt(xx, yy)
						r += int(c.R)
						g += int(c.G)
						bl += int(c.B)
						a += int(c.A)
						cnt++
					}
				}
				avg := color.NRGBA{R: uint8(r / cnt), G: uint8(g / cnt), B: uint8(bl / cnt), A: uint8(a / cnt)}
				for yy := y; yy < y2; yy++ {
					for xx := x; xx < x2; xx++ {
						dst.SetNRGBA(xx, yy, avg)
					}
				}
			}
		}
		return dst, nil
	})
}

func (img *Image) Blur(amount int) *Image {
	if amount < 1 {
		amount = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		out := n
		for i := 0; i < amount; i++ {
			out = gaussian3(out)
		}
		return out, nil
	})
}

func gaussian3(src *image.NRGBA) *image.NRGBA {
	// Separable approximation of GD's IMG_FILTER_GAUSSIAN_BLUR (3x3).
	k := []float64{1, 2, 1}
	tmp := convolve1D(src, k, true)
	return convolve1D(tmp, k, false)
}

func convolve1D(src *image.NRGBA, k []float64, horizontal bool) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	sum := 0.0
	for _, v := range k {
		sum += v
	}
	r := len(k) / 2
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var rr, gg, bb, aa float64
			for i, kv := range k {
				sx, sy := x, y
				if horizontal {
					sx = x + i - r
				} else {
					sy = y + i - r
				}
				c := pixelAt(src, clampInt(sx, b.Min.X, b.Max.X-1), clampInt(sy, b.Min.Y, b.Max.Y-1))
				rr += float64(c.R) * kv
				gg += float64(c.G) * kv
				bb += float64(c.B) * kv
				aa += float64(c.A) * kv
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: clamp8(int(math.Round(rr / sum))),
				G: clamp8(int(math.Round(gg / sum))),
				B: clamp8(int(math.Round(bb / sum))),
				A: clamp8(int(math.Round(aa / sum))),
			})
		}
	}
	return dst
}

func (img *Image) Sharpen(amount int) *Image {
	if amount < 0 {
		amount = 0
	}
	minV := 0.0
	if amount >= 10 {
		minV = float64(amount) * -0.01
	}
	maxV := float64(amount) * -0.025
	abs := ((4*minV + 4*maxV) * -1) + 1
	kernel := [3][3]float64{
		{minV, maxV, minV},
		{maxV, abs, maxV},
		{minV, maxV, minV},
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return convolve3(n, kernel), nil
	})
}

func convolve3(src *image.NRGBA, k [3][3]float64) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var rr, gg, bb, aa float64
			for ky := 0; ky < 3; ky++ {
				for kx := 0; kx < 3; kx++ {
					c := pixelAt(src, clampInt(x+kx-1, b.Min.X, b.Max.X-1), clampInt(y+ky-1, b.Min.Y, b.Max.Y-1))
					w := k[ky][kx]
					rr += float64(c.R) * w
					gg += float64(c.G) * w
					bb += float64(c.B) * w
					aa += float64(c.A) * w
				}
			}
			origA := src.NRGBAAt(x, y).A
			dst.SetNRGBA(x, y, color.NRGBA{
				R: clamp8(int(math.Round(rr))),
				G: clamp8(int(math.Round(gg))),
				B: clamp8(int(math.Round(bb))),
				A: origA,
			})
		}
	}
	return dst
}

func (img *Image) BlendTransparency(col any) *Image {
	if img.fail() {
		return img
	}
	var bg Color
	var err error
	if col == nil {
		bg = img.BlendingColor()
	} else {
		bg, err = ParseColor(col)
		if err != nil {
			return img.setErr(err)
		}
	}
	return img.mapPixels(func(c color.NRGBA) color.NRGBA {
		return overNRGBA(bg.NRGBA(), c)
	})
}

func (img *Image) ReduceColors(limit int, background any) *Image {
	if img.fail() {
		return img
	}
	if limit < 1 {
		limit = 1
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorTransparent
	}
	_ = bg
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		pal := medianCutPalette(n, limit)
		b := n.Bounds()
		dst := image.NewNRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				dst.Set(x, y, pal.Convert(n.NRGBAAt(x, y)))
			}
		}
		return dst, nil
	})
}

func (img *Image) mapPixels(fn func(color.NRGBA) color.NRGBA) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		b := n.Bounds()
		dst := image.NewNRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				dst.SetNRGBA(x, y, fn(n.NRGBAAt(x, y)))
			}
		}
		return dst, nil
	})
}

func (img *Image) RemoveAnimation(position any) *Image {
	if img.fail() {
		return img
	}
	idx, err := parsePercentOrIndex(position, len(img.frames))
	if err != nil {
		return img.setErr(err)
	}
	if len(img.frames) == 0 {
		return img
	}
	img.frames = []Frame{img.frames[idx].clone()}
	img.loops = 0
	return img
}

func (img *Image) SliceAnimation(offset int, length int) *Image {
	if img.fail() {
		return img
	}
	n := len(img.frames)
	if offset < 0 {
		offset = 0
	}
	if offset >= n {
		return img.setErr(wrap(ErrAnimation, "slice offset out of range"))
	}
	end := n
	if length > 0 {
		end = min(n, offset+length)
	}
	cp := make([]Frame, 0, end-offset)
	for i := offset; i < end; i++ {
		cp = append(cp, img.frames[i].clone())
	}
	img.frames = cp
	return img
}



package modifier

import (
	"image"
	"image/color"
	"math"

	intcolor "github.com/yunkeweb/go-image/internal/color"
	"github.com/yunkeweb/go-image/internal/errs"
	"github.com/yunkeweb/go-image/internal/pool"
)

// MapPixels applies fn to every pixel of n in place.
func MapPixels(n *image.NRGBA, fn func(color.NRGBA) color.NRGBA) *image.NRGBA {
	b := n.Bounds()
	dst := pool.AcquireRect(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetNRGBA(x, y, fn(n.NRGBAAt(x, y)))
		}
	}
	return dst
}

// Greyscale converts n to luma and keeps alpha.
func Greyscale(n *image.NRGBA) *image.NRGBA {
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		y := uint8((int(c.R)*299 + int(c.G)*587 + int(c.B)*114) / 1000)
		return color.NRGBA{R: y, G: y, B: y, A: c.A}
	})
}

// Invert inverts RGB channels and keeps alpha.
func Invert(n *image.NRGBA) *image.NRGBA {
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A}
	})
}

// Brightness adjusts luma by level percent.
func Brightness(n *image.NRGBA, level int) *image.NRGBA {
	delta := int(math.Round(float64(level) * 2.55))
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{
			R: intcolor.Clamp8(int(c.R) + delta),
			G: intcolor.Clamp8(int(c.G) + delta),
			B: intcolor.Clamp8(int(c.B) + delta),
			A: c.A,
		}
	})
}

// Contrast adjusts contrast by level percent.
func Contrast(n *image.NRGBA, level int) *image.NRGBA {
	c := intcolor.ClampInt(level, -100, 100)
	factor := (259.0 * (float64(c) + 255)) / (255 * (259 - float64(c)))
	return MapPixels(n, func(px color.NRGBA) color.NRGBA {
		adj := func(v uint8) uint8 {
			x := factor*(float64(v)-128) + 128
			return intcolor.Clamp8(int(math.Round(x)))
		}
		return color.NRGBA{R: adj(px.R), G: adj(px.G), B: adj(px.B), A: px.A}
	})
}

// Gamma applies a gamma curve. Non-positive gamma is an error.
func Gamma(n *image.NRGBA, gamma float64) (*image.NRGBA, error) {
	if gamma <= 0 {
		return nil, errs.Wrap(errs.ErrInput, "gamma must be > 0")
	}
	inv := 1 / gamma
	lut := [256]uint8{}
	for i := 0; i < 256; i++ {
		lut[i] = intcolor.Clamp8(int(math.Round(math.Pow(float64(i)/255, inv) * 255)))
	}
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{R: lut[c.R], G: lut[c.G], B: lut[c.B], A: c.A}
	}), nil
}

// Colorize tints RGB channels by signed percent deltas.
func Colorize(n *image.NRGBA, red, green, blue int) *image.NRGBA {
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{
			R: intcolor.Clamp8(int(c.R) + red),
			G: intcolor.Clamp8(int(c.G) + green),
			B: intcolor.Clamp8(int(c.B) + blue),
			A: c.A,
		}
	})
}

// Pixelate mosaics n with square cells of the given size.
func Pixelate(n *image.NRGBA, size int) *image.NRGBA {
	if size < 1 {
		size = 1
	}
	b := n.Bounds()
	dst := pool.AcquireRect(b)
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
	return dst
}

// Blur applies a box blur of the given radius.
func Blur(n *image.NRGBA, amount int) *image.NRGBA {
	if amount < 1 {
		amount = 1
	}
	out := n
	for i := 0; i < amount; i++ {
		next := gaussian3(out)
		if out != n {
			pool.Release(out)
		}
		out = next
	}
	return out
}

func gaussian3(src *image.NRGBA) *image.NRGBA {
	k := []float64{1, 2, 1}
	tmp := convolve1D(src, k, true)
	out := convolve1D(tmp, k, false)
	if tmp != src {
		pool.Release(tmp)
	}
	return out
}

func convolve1D(src *image.NRGBA, k []float64, horizontal bool) *image.NRGBA {
	b := src.Bounds()
	dst := pool.AcquireRect(b)
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
				c := pool.PixelAt(src, intcolor.ClampInt(sx, b.Min.X, b.Max.X-1), intcolor.ClampInt(sy, b.Min.Y, b.Max.Y-1))
				rr += float64(c.R) * kv
				gg += float64(c.G) * kv
				bb += float64(c.B) * kv
				aa += float64(c.A) * kv
			}
			dst.SetNRGBA(x, y, color.NRGBA{
				R: intcolor.Clamp8(int(math.Round(rr / sum))),
				G: intcolor.Clamp8(int(math.Round(gg / sum))),
				B: intcolor.Clamp8(int(math.Round(bb / sum))),
				A: intcolor.Clamp8(int(math.Round(aa / sum))),
			})
		}
	}
	return dst
}

// Sharpen applies an unsharp-mask of the given amount.
func Sharpen(n *image.NRGBA, amount int) *image.NRGBA {
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
	return convolve3(n, kernel)
}

func convolve3(src *image.NRGBA, k [3][3]float64) *image.NRGBA {
	b := src.Bounds()
	dst := pool.AcquireRect(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var rr, gg, bb float64
			for ky := 0; ky < 3; ky++ {
				for kx := 0; kx < 3; kx++ {
					c := pool.PixelAt(src, intcolor.ClampInt(x+kx-1, b.Min.X, b.Max.X-1), intcolor.ClampInt(y+ky-1, b.Min.Y, b.Max.Y-1))
					w := k[ky][kx]
					rr += float64(c.R) * w
					gg += float64(c.G) * w
					bb += float64(c.B) * w
				}
			}
			origA := src.NRGBAAt(x, y).A
			dst.SetNRGBA(x, y, color.NRGBA{
				R: intcolor.Clamp8(int(math.Round(rr))),
				G: intcolor.Clamp8(int(math.Round(gg))),
				B: intcolor.Clamp8(int(math.Round(bb))),
				A: origA,
			})
		}
	}
	return dst
}

// BlendTransparency composites n over bg.
func BlendTransparency(n *image.NRGBA, bg color.NRGBA) *image.NRGBA {
	return MapPixels(n, func(c color.NRGBA) color.NRGBA {
		return intcolor.OverNRGBA(bg, c)
	})
}

// ReduceColors maps n onto pal.
func ReduceColors(n *image.NRGBA, pal color.Palette) *image.NRGBA {
	b := n.Bounds()
	dst := pool.AcquireRect(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, pal.Convert(n.NRGBAAt(x, y)))
		}
	}
	return dst
}

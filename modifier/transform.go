package modifier

import (
	"image"
	"image/color"
	"math"

	"github.com/yunkeweb/go-image/internal/pool"
)

func Flip(n *image.NRGBA) *image.NRGBA {
	b := n.Bounds()
	dst := pool.AcquireRect(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		sy := b.Max.Y - 1 - (y - b.Min.Y)
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.SetNRGBA(x, y, n.NRGBAAt(x, sy))
		}
	}
	return dst
}

func Flop(n *image.NRGBA) *image.NRGBA {
	b := n.Bounds()
	dst := pool.AcquireRect(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			sx := b.Max.X - 1 - (x - b.Min.X)
			dst.SetNRGBA(x, y, n.NRGBAAt(sx, y))
		}
	}
	return dst
}

func Rotate(src *image.NRGBA, angleDeg float64, bg color.NRGBA) *image.NRGBA {
	angleDeg = math.Mod(angleDeg, 360)
	if angleDeg < 0 {
		angleDeg += 360
	}
	if angleDeg == 0 {
		return pool.Clone(src)
	}
	return rotateNRGBA(src, angleDeg, bg)
}

func rotateNRGBA(src *image.NRGBA, angleDeg float64, bg color.NRGBA) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if almost90(angleDeg) {
		return rotate90(src)
	}
	if almost90(angleDeg - 180) {
		return rotate180(src)
	}
	if almost90(angleDeg - 270) {
		return rotate270(src)
	}
	rad := angleDeg * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	cx := float64(w) / 2
	cy := float64(h) / 2
	corners := [][2]float64{
		{0, 0}, {float64(w), 0}, {float64(w), float64(h)}, {0, float64(h)},
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, c := range corners {
		x := (c[0] - cx)
		y := (c[1] - cy)
		rx := x*cos + y*sin
		ry := -x*sin + y*cos
		if rx < minX {
			minX = rx
		}
		if ry < minY {
			minY = ry
		}
		if rx > maxX {
			maxX = rx
		}
		if ry > maxY {
			maxY = ry
		}
	}
	nw := int(math.Ceil(maxX - minX))
	nh := int(math.Ceil(maxY - minY))
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := pool.Blank(nw, nh, bg)
	ncx := float64(nw) / 2
	ncy := float64(nh) / 2
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			dx := float64(x) + 0.5 - ncx
			dy := float64(y) + 0.5 - ncy
			sx := dx*cos - dy*sin + cx
			sy := dx*sin + dy*cos + cy
			dst.SetNRGBA(x, y, sampleBilinear(src, sx, sy, bg))
		}
	}
	return dst
}

func almost90(v float64) bool {
	v = math.Mod(v, 360)
	if v < 0 {
		v += 360
	}
	return math.Abs(v-90) < 0.001
}

func rotate90(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := pool.Acquire(h, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.SetNRGBA(y, w-1-x, src.NRGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate180(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := pool.Acquire(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.SetNRGBA(w-1-x, h-1-y, src.NRGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate270(src *image.NRGBA) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := pool.Acquire(h, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.SetNRGBA(h-1-y, x, src.NRGBAAt(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func sampleBilinear(src *image.NRGBA, fx, fy float64, bg color.NRGBA) color.NRGBA {
	b := src.Bounds()
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1, y1 := x0+1, y0+1
	tx := fx - float64(x0)
	ty := fy - float64(y0)
	c00 := atOr(src, b, x0, y0, bg)
	c10 := atOr(src, b, x1, y0, bg)
	c01 := atOr(src, b, x0, y1, bg)
	c11 := atOr(src, b, x1, y1, bg)
	return lerpNRGBA(lerpNRGBA(c00, c10, tx), lerpNRGBA(c01, c11, tx), ty)
}

func atOr(src *image.NRGBA, b image.Rectangle, x, y int, bg color.NRGBA) color.NRGBA {
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return bg
	}
	return src.NRGBAAt(x, y)
}

func lerpNRGBA(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: lerpU8(a.R, b.R, t),
		G: lerpU8(a.G, b.G, t),
		B: lerpU8(a.B, b.B, t),
		A: lerpU8(a.A, b.A, t),
	}
}

func lerpU8(a, b uint8, t float64) uint8 {
	v := float64(a)*(1-t) + float64(b)*t
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}

// OrientPixels applies EXIF orientation 2–8. Orientation 1 is a no-op.
func OrientPixels(n *image.NRGBA, orient int) *image.NRGBA {
	switch orient {
	case 2:
		return Flop(n)
	case 3:
		return Rotate(n, 180, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	case 4:
		r := Rotate(n, 180, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		out := Flop(r)
		if r != n {
			pool.Release(r)
		}
		return out
	case 5:
		r := Rotate(n, 270, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		out := Flop(r)
		if r != n {
			pool.Release(r)
		}
		return out
	case 6:
		return Rotate(n, 270, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	case 7:
		r := Rotate(n, 90, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		out := Flop(r)
		if r != n {
			pool.Release(r)
		}
		return out
	case 8:
		return Rotate(n, 90, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	default:
		return n
	}
}

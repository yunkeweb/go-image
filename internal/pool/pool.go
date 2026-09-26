package pool

import (
	"image"
	"image/color"
	"image/draw"
	"sync"
)

// pixPool reuses NRGBA backing stores so chained modifiers (Cover, Sharpen, …)
// do not retain every intermediate Pix buffer until GC.
var pixPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 64*64*4)
		return buf
	},
}

func Acquire(w, h int) *image.NRGBA {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	n := w * h * 4
	buf := pixPool.Get().([]byte)
	if cap(buf) < n {
		buf = make([]byte, n)
	} else {
		buf = buf[:n]
		clear(buf)
	}
	return &image.NRGBA{
		Pix:    buf,
		Stride: w * 4,
		Rect:   image.Rect(0, 0, w, h),
	}
}

func AcquireRect(r image.Rectangle) *image.NRGBA {
	if r.Empty() {
		return Acquire(1, 1)
	}
	img := Acquire(r.Dx(), r.Dy())
	img.Rect = r
	return img
}

// MaxPooledPix is the largest Pix backing store returned to pixPool.
const MaxPooledPix = 16 << 20 // 16 MiB

func Release(img *image.NRGBA) {
	if img == nil || img.Pix == nil {
		return
	}
	pix := img.Pix
	img.Pix = nil
	if cap(pix) > MaxPooledPix {
		return
	}
	pixPool.Put(pix[:0])
}

func Clone(n *image.NRGBA) *image.NRGBA {
	if n == nil {
		return nil
	}
	b := n.Bounds()
	dst := AcquireRect(b)
	copy(dst.Pix, n.Pix)
	return dst
}

func Rebase(src *image.NRGBA) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	if b.Min.X == 0 && b.Min.Y == 0 {
		return src
	}
	dst := Acquire(b.Dx(), b.Dy())
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

// AsNRGBA copies src into a library-owned NRGBA buffer with draw.Draw.
func AsNRGBA(src image.Image) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := Acquire(b.Dx(), b.Dy())
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func FillRect(dst *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	r = r.Intersect(dst.Bounds())
	if r.Empty() {
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			dst.SetNRGBA(x, y, c)
		}
	}
}

func Blank(w, h int, c color.NRGBA) *image.NRGBA {
	dst := Acquire(w, h)
	FillRect(dst, dst.Bounds(), c)
	return dst
}

func HasTransparency(n *image.NRGBA) bool {
	if n == nil {
		return false
	}
	b := n.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if n.NRGBAAt(x, y).A < 255 {
				return true
			}
		}
	}
	return false
}

func PixelAt(n *image.NRGBA, x, y int) color.NRGBA {
	b := n.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return color.NRGBA{}
	}
	return n.NRGBAAt(x, y)
}

// GetForTest returns a buffer from the pool (tests only).
func GetForTest() []byte { return pixPool.Get().([]byte) }

// PutForTest returns a buffer to the pool (tests only).
func PutForTest(buf []byte) { pixPool.Put(buf[:0]) }

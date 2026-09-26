package pool

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"github.com/yunkeweb/go-image/internal/errs"
)

// pixPool reuses NRGBA backing stores so chained modifiers (Cover, Sharpen, …)
// do not retain every intermediate Pix buffer until GC.
var pixPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 64*64*4)
		return buf
	},
}

// MaxPooledPix is the largest Pix backing store returned to pixPool.
const MaxPooledPix = 16 << 20 // 16 MiB

// MaxAcquireBytes is the largest Pix allocation Acquire will perform.
const MaxAcquireBytes = 1 << 30 // 1 GiB

// PixBytes returns the NRGBA backing-store length for a w×h image.
// It rejects non-positive sizes, integer overflow of w*h*4, and
// allocations larger than MaxAcquireBytes.
func PixBytes(w, h int) (int, error) {
	if w < 1 || h < 1 {
		return 0, errs.ErrInvalidDimensions
	}
	if w > 0 && h > math.MaxInt/w {
		return 0, errs.ErrInvalidDimensions
	}
	area := w * h
	if area > math.MaxInt/4 {
		return 0, errs.ErrInvalidDimensions
	}
	n := area * 4
	if n > MaxAcquireBytes {
		return 0, errs.ErrInvalidDimensions
	}
	return n, nil
}

// Acquire returns a zeroed NRGBA of size w×h whose Pix may come from pixPool.
// Invalid or overflowing dimensions return nil; callers should treat that as
// ErrInvalidDimensions.
func Acquire(w, h int) *image.NRGBA {
	n, err := PixBytes(w, h)
	if err != nil {
		return nil
	}
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

// AcquireRect returns an NRGBA whose Bounds equal r. Empty rectangles yield a
// 1×1 buffer with Rect set to r. Overflow returns nil.
func AcquireRect(r image.Rectangle) *image.NRGBA {
	if r.Empty() {
		img := Acquire(1, 1)
		if img != nil {
			img.Rect = r
		}
		return img
	}
	img := Acquire(r.Dx(), r.Dy())
	if img == nil {
		return nil
	}
	img.Rect = r
	return img
}

// Release returns img.Pix to pixPool when cap(Pix) <= MaxPooledPix.
// Larger buffers are dropped for GC. img.Pix is nilled so the image cannot
// alias a recycled buffer.
func Release(img *image.NRGBA) {
	if img == nil || img.Pix == nil {
		return
	}
	pix := img.Pix
	img.Pix = nil
	putPix(pix)
}

func putPix(pix []byte) {
	if cap(pix) > MaxPooledPix {
		return
	}
	pixPool.Put(pix[:0])
}

// Clone copies src into a newly acquired NRGBA with the same Bounds.
// Pixels are copied row-by-row via PixOffset so non-zero Min, extra Stride,
// subimages, and row padding are preserved. The destination never shares
// src.Pix. Nil src or an allocation failure returns nil.
func Clone(src *image.NRGBA) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := AcquireRect(b)
	if dst == nil {
		return nil
	}
	rowBytes := b.Dx() * 4
	if rowBytes < 1 {
		return dst
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		srcOff := src.PixOffset(b.Min.X, y)
		dstOff := dst.PixOffset(b.Min.X, y)
		if srcOff < 0 || dstOff < 0 || srcOff+rowBytes > len(src.Pix) || dstOff+rowBytes > len(dst.Pix) {
			continue
		}
		copy(dst.Pix[dstOff:dstOff+rowBytes], src.Pix[srcOff:srcOff+rowBytes])
	}
	return dst
}

// Rebase copies src onto a (0,0) origin when Bounds.Min is not the origin.
func Rebase(src *image.NRGBA) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	if b.Min.X == 0 && b.Min.Y == 0 {
		return src
	}
	dst := Acquire(b.Dx(), b.Dy())
	if dst == nil {
		return nil
	}
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
	if dst == nil {
		return nil
	}
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

// FillRect paints r inside dst with c.
func FillRect(dst *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	if dst == nil {
		return
	}
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

// Blank returns a w×h NRGBA filled with c. Invalid sizes return nil.
func Blank(w, h int, c color.NRGBA) *image.NRGBA {
	dst := Acquire(w, h)
	if dst == nil {
		return nil
	}
	FillRect(dst, dst.Bounds(), c)
	return dst
}

// HasTransparency reports whether any pixel in n has A < 255.
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

// PixelAt returns the NRGBA at (x, y), or a zero color when out of bounds.
func PixelAt(n *image.NRGBA, x, y int) color.NRGBA {
	if n == nil {
		return color.NRGBA{}
	}
	b := n.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return color.NRGBA{}
	}
	return n.NRGBAAt(x, y)
}

// GetForTest returns a buffer from the pool (tests only).
func GetForTest() []byte { return pixPool.Get().([]byte) }

// PutForTest returns a buffer to the pool (tests only). Buffers larger than
// MaxPooledPix are dropped.
func PutForTest(buf []byte) {
	if buf == nil {
		return
	}
	putPix(buf)
}

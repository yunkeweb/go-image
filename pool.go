package goimage

import (
	"image"
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

func acquireNRGBA(w, h int) *image.NRGBA {
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

func acquireNRGBARect(r image.Rectangle) *image.NRGBA {
	if r.Empty() {
		return acquireNRGBA(1, 1)
	}
	img := acquireNRGBA(r.Dx(), r.Dy())
	img.Rect = r
	return img
}

func releaseNRGBA(img *image.NRGBA) {
	if img == nil || img.Pix == nil {
		return
	}
	pix := img.Pix[:0]
	img.Pix = nil
	pixPool.Put(pix)
}

func cloneNRGBA(n *image.NRGBA) *image.NRGBA {
	if n == nil {
		return nil
	}
	b := n.Bounds()
	dst := acquireNRGBARect(b)
	copy(dst.Pix, n.Pix)
	return dst
}

func rebaseNRGBA(src *image.NRGBA) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	if b.Min.X == 0 && b.Min.Y == 0 {
		return src
	}
	dst := acquireNRGBA(b.Dx(), b.Dy())
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

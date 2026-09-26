package goimage

import (
	"image"
	"image/color"
	"image/draw"
)

// Frame is one animation frame. Delay is in seconds (PHP Frame::delay).
type Frame struct {
	Img        *image.NRGBA
	Delay      float64
	Dispose    int
	OffsetLeft int
	OffsetTop  int
}

func (f Frame) Size() Size {
	if f.Img == nil {
		return Size{}
	}
	b := f.Img.Bounds()
	return Size{Width: b.Dx(), Height: b.Dy()}
}

func (f Frame) clone() Frame {
	out := f
	if f.Img != nil {
		b := f.Img.Bounds()
		cp := image.NewNRGBA(b)
		draw.Draw(cp, b, f.Img, b.Min, draw.Src)
		out.Img = cp
	}
	return out
}

func asNRGBA(src image.Image) *image.NRGBA {
	if src == nil {
		return nil
	}
	if n, ok := src.(*image.NRGBA); ok {
		return n
	}
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func newBlank(w, h int, bg Color) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	fillRect(dst, dst.Bounds(), bg.NRGBA())
	return dst
}

func fillRect(dst *image.NRGBA, r image.Rectangle, c color.NRGBA) {
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

package goimage

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
)

// Frame is one animation frame. Delay is in seconds.
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
		out.Img = cloneNRGBA(f.Img)
	}
	return out
}

// asNRGBA copies src into a library-owned NRGBA buffer with draw.Draw.
// JPEG YCbCr, Paletted, RGBA, NRGBA, and other image.Image values never
// share the caller's Pix slice and never panic on a concrete type assertion.
func asNRGBA(src image.Image) *image.NRGBA {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	dst := acquireNRGBA(b.Dx(), b.Dy())
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func newBlank(w, h int, bg Color) *image.NRGBA {
	dst := acquireNRGBA(w, h)
	fillRect(dst, dst.Bounds(), bg.NRGBA())
	return dst
}

func (img *Image) replaceAll(fn func(*image.NRGBA) (*image.NRGBA, error)) *Image {
	return img.eachFrame(func(f *Frame) error {
		old := f.Img
		n, err := fn(old)
		if err != nil {
			return err
		}
		if n != old {
			releaseNRGBA(old)
		}
		f.Img = n
		return nil
	})
}

func (img *Image) replaceAllGeometry(fn func(*image.NRGBA) (*image.NRGBA, error)) *Image {
	img = img.replaceAll(fn)
	if img.fail() {
		return img
	}
	img.resetGIFFrameLayout()
	return img
}

// resetGIFFrameLayout rebases each frame after Crop/Resize.
// Opaque full-canvas frames keep DisposalNone (or Previous) so players
// do not flash a background between frames. Transparent or region-cropped
// frames reset DisposalBackground and sit at origin (0,0).
func (img *Image) resetGIFFrameLayout() {
	if img == nil {
		return
	}
	cw, ch := img.Width(), img.Height()
	for i := range img.frames {
		f := &img.frames[i]
		if f.Img == nil {
			continue
		}
		rebased := rebaseNRGBA(f.Img)
		if rebased != f.Img {
			releaseNRGBA(f.Img)
			f.Img = rebased
		}
		f.OffsetLeft = 0
		f.OffsetTop = 0
		if hasTransparency(f.Img) || !frameCoversCanvas(f, cw, ch) {
			f.Dispose = int(gif.DisposalBackground)
			continue
		}
		if f.Dispose == 0 || f.Dispose == int(gif.DisposalBackground) {
			f.Dispose = int(gif.DisposalNone)
		}
	}
}

func frameCoversCanvas(f *Frame, canvasW, canvasH int) bool {
	if f == nil || f.Img == nil || canvasW < 1 || canvasH < 1 {
		return false
	}
	b := f.Img.Bounds()
	return f.OffsetLeft == 0 && f.OffsetTop == 0 &&
		b.Min.X == 0 && b.Min.Y == 0 &&
		b.Dx() >= canvasW && b.Dy() >= canvasH
}

func gifDisposalForFrame(f *Frame, canvasW, canvasH int) byte {
	if f == nil || f.Img == nil {
		return gif.DisposalNone
	}
	opaqueFull := !hasTransparency(f.Img) && frameCoversCanvas(f, canvasW, canvasH)
	if opaqueFull {
		if f.Dispose == int(gif.DisposalPrevious) {
			return gif.DisposalPrevious
		}
		return gif.DisposalNone
	}
	if f.Dispose == int(gif.DisposalPrevious) {
		return gif.DisposalPrevious
	}
	return gif.DisposalBackground
}

func (img *Image) releaseFrames(except int) {
	if img == nil {
		return
	}
	for i := range img.frames {
		if i == except {
			continue
		}
		releaseNRGBA(img.frames[i].Img)
		img.frames[i].Img = nil
	}
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

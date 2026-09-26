package modifier

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"

	"github.com/yunkeweb/go-image/internal/pool"
)

// AnimFrame is one composited animation frame on the GIF logical canvas.
type AnimFrame struct {
	Img        *image.NRGBA
	Delay      float64
	Dispose    int
	OffsetLeft int
	OffsetTop  int
}

// FrameCoversCanvas reports whether f paints every pixel of a w×h canvas.
func FrameCoversCanvas(f AnimFrame, canvasW, canvasH int) bool {
	if f.Img == nil || canvasW < 1 || canvasH < 1 {
		return false
	}
	b := f.Img.Bounds()
	return f.OffsetLeft == 0 && f.OffsetTop == 0 &&
		b.Min.X == 0 && b.Min.Y == 0 &&
		b.Dx() >= canvasW && b.Dy() >= canvasH
}

// GIFDisposal chooses an encode-time GIF disposal method from a composited frame.
func GIFDisposal(f AnimFrame, canvasW, canvasH int) byte {
	if f.Img == nil {
		return gif.DisposalNone
	}
	opaqueFull := !pool.HasTransparency(f.Img) && FrameCoversCanvas(f, canvasW, canvasH)
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

// ResetGIFLayout rebases every frame onto a (0,0) origin after geometry changes.
func ResetGIFLayout(frames []AnimFrame, canvasW, canvasH int) {
	for i := range frames {
		f := &frames[i]
		if f.Img == nil {
			continue
		}
		rebased := pool.Rebase(f.Img)
		if rebased != f.Img {
			pool.Release(f.Img)
			f.Img = rebased
		}
		f.OffsetLeft = 0
		f.OffsetTop = 0
		if pool.HasTransparency(f.Img) || !FrameCoversCanvas(*f, canvasW, canvasH) {
			f.Dispose = int(gif.DisposalBackground)
			continue
		}
		if f.Dispose == 0 || f.Dispose == int(gif.DisposalBackground) {
			f.Dispose = int(gif.DisposalNone)
		}
	}
}

// gifCanvasSize returns the GIF logical screen. Config.Width/Height are used
// when both are >= 1; otherwise the union of frame Bounds is used.
func gifCanvasSize(g *gif.GIF) (w, h int) {
	if g == nil {
		return 1, 1
	}
	if g.Config.Width >= 1 && g.Config.Height >= 1 {
		return g.Config.Width, g.Config.Height
	}
	for _, p := range g.Image {
		if p == nil {
			continue
		}
		b := p.Bounds()
		if b.Max.X > w {
			w = b.Max.X
		}
		if b.Max.Y > h {
			h = b.Max.Y
		}
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

func gifBackground(g *gif.GIF) color.NRGBA {
	if g == nil {
		return color.NRGBA{}
	}
	pal, ok := g.Config.ColorModel.(color.Palette)
	if !ok || int(g.BackgroundIndex) >= len(pal) {
		return color.NRGBA{}
	}
	r, g8, b, a := pal[g.BackgroundIndex].RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g8 >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func applyPaletted(canvas *image.NRGBA, pal *image.Paletted) {
	if canvas == nil || pal == nil {
		return
	}
	pb := pal.Bounds().Intersect(canvas.Bounds())
	for y := pb.Min.Y; y < pb.Max.Y; y++ {
		for x := pb.Min.X; x < pb.Max.X; x++ {
			r, g8, b, a := pal.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			canvas.SetNRGBA(x, y, color.NRGBA{R: uint8(r >> 8), G: uint8(g8 >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
		}
	}
}

// CompositeGIF expands every GIF frame onto the logical canvas using an
// explicit snapshot → apply → restore disposal state machine.
//
// Each returned AnimFrame is the fully composited picture the viewer sees
// after that frame is applied and before disposal runs. Delay and the original
// disposal method are preserved. Offsets are zero because pixels already sit
// on the logical canvas.
func CompositeGIF(g *gif.GIF) []AnimFrame {
	if g == nil || len(g.Image) == 0 {
		return nil
	}
	w, h := gifCanvasSize(g)
	canvas := pool.Acquire(w, h)
	if canvas == nil {
		return nil
	}
	bg := gifBackground(g)
	pool.FillRect(canvas, canvas.Bounds(), bg)

	frames := make([]AnimFrame, 0, len(g.Image))
	for i, pal := range g.Image {
		delay := 0.0
		if i < len(g.Delay) {
			delay = float64(g.Delay[i]) / 100
		}
		dispose := int(gif.DisposalNone)
		if i < len(g.Disposal) {
			dispose = int(g.Disposal[i])
		}

		// snapshot: canvas as it exists before this frame is applied.
		var snapshot *image.NRGBA
		if dispose == int(gif.DisposalPrevious) {
			snapshot = pool.Clone(canvas)
		}

		// apply: paint non-transparent paletted pixels onto the canvas.
		applyPaletted(canvas, pal)

		cloned := pool.Clone(canvas)
		if cloned == nil {
			pool.Release(snapshot)
			continue
		}
		frames = append(frames, AnimFrame{
			Img:        cloned,
			Delay:      delay,
			Dispose:    dispose,
			OffsetLeft: 0,
			OffsetTop:  0,
		})

		// restore: honour the frame's disposal method for the next iteration.
		switch dispose {
		case int(gif.DisposalBackground):
			if pal != nil {
				pool.FillRect(canvas, pal.Bounds().Intersect(canvas.Bounds()), bg)
			}
		case int(gif.DisposalPrevious):
			if snapshot != nil {
				draw.Draw(canvas, canvas.Bounds(), snapshot, snapshot.Bounds().Min, draw.Src)
			}
		}
		pool.Release(snapshot)
	}
	pool.Release(canvas)
	return frames
}

// DelayToGIF converts a delay in seconds to GIF centiseconds.
func DelayToGIF(seconds float64) int {
	if seconds <= 0 {
		return 0
	}
	n := int(seconds * 100)
	if n < 1 {
		n = 1
	}
	return n
}

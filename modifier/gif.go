package modifier

import (
	"image"
	"image/color"
	"image/gif"

	"github.com/yunkeweb/go-image/internal/pool"
)

// AnimFrame is one composited animation frame.
type AnimFrame struct {
	Img        *image.NRGBA
	Delay      float64
	Dispose    int
	OffsetLeft int
	OffsetTop  int
}

func FrameCoversCanvas(f AnimFrame, canvasW, canvasH int) bool {
	if f.Img == nil || canvasW < 1 || canvasH < 1 {
		return false
	}
	b := f.Img.Bounds()
	return f.OffsetLeft == 0 && f.OffsetTop == 0 &&
		b.Min.X == 0 && b.Min.Y == 0 &&
		b.Dx() >= canvasW && b.Dy() >= canvasH
}

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

func CompositeGIF(g *gif.GIF) []AnimFrame {
	if g == nil || len(g.Image) == 0 {
		return nil
	}
	w, h := 0, 0
	for _, p := range g.Image {
		if p.Bounds().Max.X > w {
			w = p.Bounds().Max.X
		}
		if p.Bounds().Max.Y > h {
			h = p.Bounds().Max.Y
		}
	}
	canvas := pool.Acquire(w, h)
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
		pb := pal.Bounds()
		snapshot := make([]color.NRGBA, pb.Dx()*pb.Dy())
		si := 0
		for y := pb.Min.Y; y < pb.Max.Y; y++ {
			for x := pb.Min.X; x < pb.Max.X; x++ {
				snapshot[si] = canvas.NRGBAAt(x, y)
				si++
				r, g8, b, a := pal.At(x, y).RGBA()
				if a == 0 {
					continue
				}
				canvas.SetNRGBA(x, y, color.NRGBA{R: uint8(r >> 8), G: uint8(g8 >> 8), B: uint8(b >> 8), A: uint8(a >> 8)})
			}
		}
		cloned := pool.Clone(canvas)
		storedDispose := int(gif.DisposalNone)
		if pool.HasTransparency(cloned) {
			storedDispose = int(gif.DisposalBackground)
		} else if dispose == int(gif.DisposalPrevious) {
			storedDispose = dispose
		}
		frames = append(frames, AnimFrame{
			Img:        cloned,
			Delay:      delay,
			Dispose:    storedDispose,
			OffsetLeft: 0,
			OffsetTop:  0,
		})
		if dispose == int(gif.DisposalBackground) {
			pool.FillRect(canvas, pb, color.NRGBA{})
		} else if dispose == int(gif.DisposalPrevious) {
			si = 0
			for y := pb.Min.Y; y < pb.Max.Y; y++ {
				for x := pb.Min.X; x < pb.Max.X; x++ {
					canvas.SetNRGBA(x, y, snapshot[si])
					si++
				}
			}
		}
	}
	pool.Release(canvas)
	return frames
}

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

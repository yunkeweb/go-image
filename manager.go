package goimage

import (
	"image"
	"image/color"
	"image/gif"
	"io"
)

// Manager holds a Config reused across Open, Decode, and New.
type Manager struct {
	cfg Config
}

func (m *Manager) Driver() string { return "go" }

func (m *Manager) Config() Config { return m.cfg }

func newCanvas(width, height int, cfg Config) *Image {
	if width < 1 || height < 1 {
		return failed(wrap(ErrGeometry, "width and height must be >= 1"))
	}
	img := newImage([]Frame{{Img: newBlank(width, height, ColorTransparent)}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img
}

// New creates a transparent canvas using the manager's config.
func (m *Manager) New(width, height int) *Image {
	return newCanvas(width, height, m.cfg)
}

// Create is an alias of Manager.New.
func (m *Manager) Create(width, height int) *Image { return m.New(width, height) }

// Open decodes an image from a filesystem path.
func (m *Manager) Open(path string) *Image {
	return result(decodeFile(path, m.cfg))
}

// Decode decodes an image from r.
func (m *Manager) Decode(r io.Reader) *Image {
	return result(decodeReader(r, m.cfg))
}

// DecodeBytes decodes an image from encoded bytes.
func (m *Manager) DecodeBytes(data []byte) *Image {
	return result(decodeBytes(data, "", m.cfg))
}

// DecodeDataURI decodes a data URI.
func (m *Manager) DecodeDataURI(uri string) *Image {
	return result(decodeDataURI(uri, m.cfg))
}

// FromImage wraps a standard-library image.Image.
func (m *Manager) FromImage(src image.Image) *Image {
	return result(fromStdImage(src, m.cfg))
}

// Animate builds a multi-frame image using the manager's config.
func (m *Manager) Animate(init func(*Animation)) *Image {
	return buildAnimation(m.cfg, init)
}

func buildAnimation(cfg Config, init func(*Animation)) *Image {
	a := &Animation{cfg: cfg}
	if init != nil {
		init(a)
	}
	if a.err != nil {
		return failed(a.err)
	}
	if len(a.frames) == 0 {
		return failed(wrap(ErrAnimation, "animation has no frames"))
	}
	img := newImage(a.frames, cfg)
	img.loops = a.loops
	img.origin = Origin{MediaType: "image/gif"}
	return img
}

// Animation collects frames for Animate.
type Animation struct {
	cfg    Config
	frames []Frame
	loops  int
	err    error
}

// Add appends src as one or more frames with the given delay in seconds.
func (a *Animation) Add(src *Image, delaySeconds float64) *Animation {
	if a.err != nil {
		return a
	}
	if src == nil || src.Err() != nil {
		if src != nil {
			a.err = src.Err()
		} else {
			a.err = wrap(ErrAnimation, "nil frame")
		}
		return a
	}
	for _, f := range src.frames {
		f.Delay = delaySeconds
		a.frames = append(a.frames, f.clone())
	}
	return a
}

// AddImage is an alias of Add.
func (a *Animation) AddImage(src *Image, delaySeconds float64) *Animation {
	return a.Add(src, delaySeconds)
}

// AddFile decodes path and appends it as a frame.
func (a *Animation) AddFile(path string, delaySeconds float64) *Animation {
	if a.err != nil {
		return a
	}
	img, err := decodeFile(path, a.cfg)
	if err != nil {
		a.err = err
		return a
	}
	return a.Add(img, delaySeconds)
}

func (a *Animation) SetLoops(n int) *Animation {
	a.loops = n
	return a
}

func (a *Animation) Loops(n int) *Animation { return a.SetLoops(n) }

func imageFromGIF(g *gif.GIF, cfg Config, origin Origin) *Image {
	if g == nil || len(g.Image) == 0 {
		return failed(wrap(ErrDecoder, "empty gif"))
	}
	if !cfg.DecodeAnimation || len(g.Image) == 1 {
		img := newImage([]Frame{{Img: asNRGBA(g.Image[0])}}, cfg)
		img.origin = origin
		img.loops = g.LoopCount
		return img
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
	canvas := acquireNRGBA(w, h)
	frames := make([]Frame, 0, len(g.Image))
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
		// Snapshot a full independent frame at origin so later Crop/Resize can
		// reset Disposal/Rect without sharing the compositor canvas.
		cloned := cloneNRGBA(canvas)
		storedDispose := int(gif.DisposalNone)
		if hasTransparency(cloned) {
			storedDispose = int(gif.DisposalBackground)
		} else if dispose == int(gif.DisposalPrevious) {
			storedDispose = dispose
		}
		frames = append(frames, Frame{
			Img:        cloned,
			Delay:      delay,
			Dispose:    storedDispose,
			OffsetLeft: 0,
			OffsetTop:  0,
		})
		if dispose == int(gif.DisposalBackground) {
			fillRect(canvas, pb, color.NRGBA{})
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
	releaseNRGBA(canvas)
	img := newImage(frames, cfg)
	img.loops = g.LoopCount
	img.origin = origin
	return img
}

func delayToGIF(seconds float64) int {
	if seconds <= 0 {
		return 0
	}
	n := int(seconds * 100)
	if n < 1 {
		n = 1
	}
	return n
}

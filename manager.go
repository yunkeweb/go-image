package goimage

import (
	"image"
	"image/color"
	"image/gif"
)

// Manager is the PHP ImageManager equivalent. Go uses a single stdlib driver.
type Manager struct {
	cfg Config
}

// New constructs a manager. Options map to PHP Config constructor flags.
func New(opts ...Option) *Manager {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return &Manager{cfg: cfg}
}

func (m *Manager) Driver() string { return "go" }

func (m *Manager) Config() Config { return m.cfg }

// Create a transparent canvas of the given size.
func (m *Manager) Create(width, height int) *Image {
	if width < 1 || height < 1 {
		return failed(wrap(ErrGeometry, "width and height must be >= 1"))
	}
	img := newImage([]Frame{{Img: newBlank(width, height, ColorTransparent)}}, m.cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img
}

// Read decodes a path, []byte, io.Reader, data URI, base64 string, or *Image.
func (m *Manager) Read(input any) *Image {
	img, err := decodeInput(input, m.cfg)
	if err != nil {
		return failed(err)
	}
	return img
}

// Animate builds a multi-frame image.
func (m *Manager) Animate(init func(*Animation)) *Image {
	a := &Animation{mgr: m}
	if init != nil {
		init(a)
	}
	if a.err != nil {
		return failed(a.err)
	}
	if len(a.frames) == 0 {
		return failed(wrap(ErrAnimation, "animation has no frames"))
	}
	img := newImage(a.frames, m.cfg)
	img.loops = a.loops
	img.origin = Origin{MediaType: "image/gif"}
	return img
}

// Animation is the PHP animation callback builder.
type Animation struct {
	mgr    *Manager
	frames []Frame
	loops  int
	err    error
}

func (a *Animation) Add(input any, delaySeconds float64) *Animation {
	if a.err != nil {
		return a
	}
	img := a.mgr.Read(input)
	if img.Err() != nil {
		a.err = img.Err()
		return a
	}
	for _, f := range img.frames {
		f.Delay = delaySeconds
		a.frames = append(a.frames, f.clone())
	}
	return a
}

func (a *Animation) AddImage(src *Image, delaySeconds float64) *Animation {
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

func (a *Animation) SetLoops(n int) *Animation {
	a.loops = n
	return a
}

func (a *Animation) Loops(n int) *Animation { return a.SetLoops(n) }

var defaultManager = New()

func Create(width, height int) *Image { return defaultManager.Create(width, height) }

func Read(input any) *Image { return defaultManager.Read(input) }

func Animate(init func(*Animation)) *Image { return defaultManager.Animate(init) }

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
	canvas := image.NewNRGBA(image.Rect(0, 0, w, h))
	frames := make([]Frame, 0, len(g.Image))
	for i, pal := range g.Image {
		delay := 0.0
		if i < len(g.Delay) {
			delay = float64(g.Delay[i]) / 100
		}
		dispose := 1
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
		frames = append(frames, Frame{Img: asNRGBA(canvas), Delay: delay, Dispose: dispose})
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

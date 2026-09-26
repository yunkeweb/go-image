package goimage

import (
	"image"
	"image/gif"
	"io"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-image/encoder"
	"github.com/yunkeweb/go-image/internal/pool"
	"github.com/yunkeweb/go-image/modifier"
)

func result(img *Image, err error) *Image {
	if err != nil {
		return failed(err)
	}
	return img
}

// New creates a transparent canvas of the given size.
// Options apply to this image only.
func New(width, height int, opts ...Option) *Image {
	return newCanvas(width, height, applyOptions(opts))
}

// Create is an alias of New.
func Create(width, height int, opts ...Option) *Image {
	return New(width, height, opts...)
}

// Open decodes an image from a filesystem path.
func Open(path string, opts ...Option) *Image {
	return result(decodeFile(path, applyOptions(opts)))
}

// Decode decodes an image from r.
func Decode(r io.Reader, opts ...Option) *Image {
	return result(decodeReader(r, applyOptions(opts)))
}

// DecodeBytes decodes an image from encoded bytes (JPEG, PNG, GIF, WebP, BMP, TIFF).
func DecodeBytes(data []byte, opts ...Option) *Image {
	return result(decodeBytes(data, "", applyOptions(opts)))
}

// DecodeDataURI decodes a `data:image/...;base64,...` URI.
func DecodeDataURI(uri string, opts ...Option) *Image {
	return result(decodeDataURI(uri, applyOptions(opts)))
}

// FromImage copies src into a library-owned NRGBA buffer with draw.Draw.
// YCbCr, Paletted, RGBA, NRGBA, and other image.Image values are safe: the
// original Pix slice is never retained.
func FromImage(src image.Image, opts ...Option) *Image {
	return result(fromStdImage(src, applyOptions(opts)))
}

// Animate builds a multi-frame GIF. Options apply to the resulting image.
func Animate(init func(*Animation), opts ...Option) *Image {
	return buildAnimation(applyOptions(opts), init)
}

func newCanvas(width, height int, cfg Config) *Image {
	if width < 1 || height < 1 {
		return failed(wrap(ErrGeometry, "width and height must be >= 1"))
	}
	img := newImage([]Frame{{Img: pool.Blank(width, height, ColorTransparent.NRGBA())}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img
}

func decodeFile(path string, cfg Config) (*Image, error) {
	data, err := encoder.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodeBytes(data, path, cfg)
}

func decodeReader(r io.Reader, cfg Config) (*Image, error) {
	data, err := encoder.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return decodeBytes(data, "", cfg)
}

func fromStdImage(src image.Image, cfg Config) (*Image, error) {
	if src == nil {
		return nil, wrap(ErrDecoder, "nil image")
	}
	img := newImage([]Frame{{Img: pool.AsNRGBA(src)}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img, nil
}

func decodeDataURI(s string, cfg Config) (*Image, error) {
	data, err := encoder.DecodeDataURIPayload(s)
	if err != nil {
		return nil, err
	}
	return decodeBytes(data, "", cfg)
}

func decodeBytes(data []byte, path string, cfg Config) (*Image, error) {
	if len(data) == 0 {
		return nil, wrap(ErrDecoder, "empty input")
	}
	origin := Origin{FilePath: path, MediaType: encoder.SniffMediaType(data)}

	if encoder.IsGIF(data) {
		g, err := encoder.DecodeGIF(data)
		if err != nil {
			return nil, err
		}
		img := imageFromGIF(g, cfg, origin)
		if img.Err() != nil {
			return nil, img.Err()
		}
		return img, nil
	}

	decoded, format, err := encoder.DecodeStill(data)
	if err != nil {
		return nil, wrap(ErrDecoder, "unable to decode input: %v", err)
	}
	if origin.MediaType == "application/octet-stream" && format != "" {
		if f, ferr := parseFormat(format); ferr == nil {
			origin.MediaType = f.MediaType()
		}
	}
	img := newImage([]Frame{{Img: pool.AsNRGBA(decoded)}}, cfg)
	img.origin = origin
	if format == "jpeg" {
		if exif, orient := encoder.ParseJPEGExif(data); exif != nil {
			img.exif = exif
			if cfg.AutoOrientation {
				applyOrientation(img, orient)
				img.markOrientationNormal()
			}
		}
	}
	return img, nil
}

func imageFromGIF(g *gif.GIF, cfg Config, origin Origin) *Image {
	if g == nil || len(g.Image) == 0 {
		return failed(wrap(ErrDecoder, "empty gif"))
	}
	if !cfg.DecodeAnimation || len(g.Image) == 1 {
		img := newImage([]Frame{{Img: pool.AsNRGBA(g.Image[0])}}, cfg)
		img.origin = origin
		img.loops = g.LoopCount
		return img
	}
	raw := modifier.CompositeGIF(g)
	frames := make([]Frame, len(raw))
	for i, f := range raw {
		frames[i] = Frame(f)
	}
	img := newImage(frames, cfg)
	img.loops = g.LoopCount
	img.origin = origin
	return img
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

func (a *Animation) AddImage(src *Image, delaySeconds float64) *Animation {
	return a.Add(src, delaySeconds)
}

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

func parsePercentOrIndex(v any, total int) (int, error) {
	switch t := v.(type) {
	case int:
		if t < 0 {
			t = 0
		}
		if total == 0 {
			return 0, nil
		}
		if t >= total {
			return total - 1, nil
		}
		return t, nil
	case string:
		s := strings.TrimSpace(t)
		s = strings.TrimSuffix(s, "%")
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, wrap(ErrInput, "invalid animation position %q", t)
		}
		if total <= 1 {
			return 0, nil
		}
		idx := int(float64(total-1) * f / 100)
		return clampInt(idx, 0, total-1), nil
	default:
		return 0, wrap(ErrInput, "invalid animation position %T", v)
	}
}

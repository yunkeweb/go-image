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

func configOrFailed(opts []Option) (Config, *Image) {
	cfg, err := applyOptionsChecked(opts)
	if err != nil {
		return cfg, failed(err)
	}
	return cfg, nil
}

// New creates a transparent canvas of the given size.
// Options apply to this image only. Width and height must be >= 1;
// overflow-sized canvases set ErrInvalidDimensions.
// Error handling: Check img.Err() after call chain.
func New(width, height int, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return newCanvas(width, height, cfg)
}

// Create is an alias of New.
//
// Deprecated: Use New instead.
// Error handling: Check img.Err() after call chain.
func Create(width, height int, opts ...Option) *Image {
	return New(width, height, opts...)
}

// Open decodes an image from a filesystem path.
// The first failure is stored on the returned Image and retrieved with Err().
// Error handling: Check img.Err() after call chain.
func Open(path string, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return result(decodeFile(path, cfg))
}

// Decode decodes an image from r.
// Error handling: Check img.Err() after call chain.
func Decode(r io.Reader, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return result(decodeReader(r, cfg))
}

// DecodeBytes decodes an image from encoded bytes (JPEG, PNG, GIF, WebP, BMP, TIFF).
// Error handling: Check img.Err() after call chain.
func DecodeBytes(data []byte, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return result(decodeBytes(data, "", cfg))
}

// DecodeDataURI decodes a `data:image/...` URI.
// The media type must be image/*. `;base64` is a case-insensitive flag
// parameter. The payload is percent-decoded. Non-image types, unknown
// parameters, empty payloads, and illegal encoding set ErrDecoder.
// Error handling: Check img.Err() after call chain.
func DecodeDataURI(uri string, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return result(decodeDataURI(uri, cfg))
}

// FromImage copies src into a library-owned NRGBA buffer with draw.Draw.
// YCbCr, Paletted, RGBA, NRGBA, and other image.Image values are safe: the
// original Pix slice is never retained.
// Error handling: Check img.Err() after call chain.
func FromImage(src image.Image, opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return result(fromStdImage(src, cfg))
}

// Animate builds a multi-frame GIF. Options apply to the resulting image.
// Error handling: Check img.Err() after call chain.
func Animate(init func(*Animation), opts ...Option) *Image {
	cfg, fail := configOrFailed(opts)
	if fail != nil {
		return fail
	}
	return buildAnimation(cfg, init)
}

func newCanvas(width, height int, cfg Config) *Image {
	if width < 1 || height < 1 {
		return failed(wrap(ErrGeometry, "width and height must be >= 1"))
	}
	if err := checkImageLimits(cfg, width, height, 1); err != nil {
		return failed(err)
	}
	if _, err := pool.PixBytes(width, height); err != nil {
		return failed(wrap(ErrInvalidDimensions, "invalid dimensions"))
	}
	blank := pool.Blank(width, height, ColorTransparent.NRGBA())
	if blank == nil {
		return failed(wrap(ErrInvalidDimensions, "invalid dimensions"))
	}
	img := newImage([]Frame{{img: blank}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img
}

func decodeFile(path string, cfg Config) (*Image, error) {
	data, err := encoder.ReadFileLimited(path, cfg.Limits.MaxInputBytes)
	if err != nil {
		return nil, err
	}
	return decodeBytes(data, path, cfg)
}

func decodeReader(r io.Reader, cfg Config) (*Image, error) {
	data, err := encoder.ReadAllLimited(r, cfg.Limits.MaxInputBytes)
	if err != nil {
		return nil, err
	}
	return decodeBytes(data, "", cfg)
}

func fromStdImage(src image.Image, cfg Config) (*Image, error) {
	if src == nil {
		return nil, wrap(ErrDecoder, "nil image")
	}
	b := src.Bounds()
	if err := checkImageLimits(cfg, b.Dx(), b.Dy(), 1); err != nil {
		return nil, err
	}
	n := pool.AsNRGBA(src)
	if n == nil {
		return nil, wrap(ErrInvalidDimensions, "invalid dimensions")
	}
	img := newImage([]Frame{{img: n}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img, nil
}

func decodeDataURI(s string, cfg Config) (*Image, error) {
	if err := checkInputBytes(cfg, int64(len(s))); err != nil {
		return nil, err
	}
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
	if err := checkInputBytes(cfg, int64(len(data))); err != nil {
		return nil, err
	}
	origin := Origin{FilePath: path, MediaType: encoder.SniffMediaType(data)}

	conf, _, err := encoder.DecodeConfig(data)
	if err != nil {
		return nil, wrap(ErrDecoder, "unable to decode input: %v", err)
	}
	frames := 1
	if encoder.IsGIF(data) {
		n, cerr := encoder.CountGIFFrames(data)
		if cerr != nil {
			return nil, cerr
		}
		if n < 1 {
			return nil, wrap(ErrDecoder, "unable to decode gif: no frames")
		}
		frames = n
	}
	if err := checkImageLimits(cfg, conf.Width, conf.Height, frames); err != nil {
		return nil, err
	}

	if encoder.IsGIF(data) {
		g, err := encoder.DecodeGIF(data)
		if err != nil {
			return nil, err
		}
		if err := checkImageLimits(cfg, conf.Width, conf.Height, len(g.Image)); err != nil {
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
	converted := pool.AsNRGBA(decoded)
	if converted == nil {
		return nil, wrap(ErrInvalidDimensions, "invalid dimensions")
	}
	img := newImage([]Frame{{img: converted}}, cfg)
	img.origin = origin
	if format == "jpeg" {
		if exif, orient := encoder.ParseJPEGExif(data); exif != nil {
			img.exif = cloneExif(exif)
			if cfg.AutoOrientation {
				applyOrientation(img, orient)
				img.clearOrientation()
			}
		}
	}
	return img, nil
}

func imageFromGIF(g *gif.GIF, cfg Config, origin Origin) *Image {
	if g == nil || len(g.Image) == 0 {
		return failed(wrap(ErrDecoder, "empty gif"))
	}
	raw := modifier.CompositeGIF(g)
	if len(raw) == 0 {
		return failed(wrap(ErrDecoder, "empty gif"))
	}
	if !cfg.DecodeAnimation && len(raw) > 1 {
		for i := 1; i < len(raw); i++ {
			pool.Release(raw[i].Img)
		}
		raw = raw[:1]
	}
	frames := make([]Frame, len(raw))
	for i, f := range raw {
		frames[i] = frameFromAnim(f)
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

// Animation collects frames for Animate. It is not safe for concurrent use.
type Animation struct {
	cfg    Config
	frames []Frame
	loops  int
	err    error
}

// Add appends a clone of src's frames with the given delay in seconds.
// Error handling: Check img.Err() after call chain.
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
//
// Deprecated: Use Add instead.
// Error handling: Check img.Err() after call chain.
func (a *Animation) AddImage(src *Image, delaySeconds float64) *Animation {
	return a.Add(src, delaySeconds)
}

// AddFile decodes path and appends it as a frame.
// Error handling: Check img.Err() after call chain.
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

// SetLoops stores the Netscape loop count (0 means loop forever).
// Error handling: Check img.Err() after call chain.
func (a *Animation) SetLoops(n int) *Animation {
	a.loops = n
	return a
}

// Loops is an alias of SetLoops.
//
// Deprecated: Use SetLoops instead.
// Error handling: Check img.Err() after call chain.
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

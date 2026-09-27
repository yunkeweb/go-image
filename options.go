package goimage

import (
	"strings"

	"github.com/yunkeweb/go-image/modifier"
)

// Config holds decode and encode defaults for a single Open/Decode/New/Animate call.
// Values are copied into each Image; mutating a Config after the call does not
// affect in-flight images. Config itself is safe to copy across goroutines.
type Config struct {
	AutoOrientation bool
	DecodeAnimation bool
	BlendingColor   any
	Strip           bool
	Limits          Limits
}

// Limits caps decode resource use. A zero field means that check is skipped.
// DefaultConfig leaves every field at zero so existing callers are unchanged.
type Limits struct {
	MaxInputBytes int64
	MaxWidth      int
	MaxHeight     int
	MaxPixels     int64
	MaxFrames     int
}

var defaultConfig = Config{
	AutoOrientation: true,
	DecodeAnimation: true,
	BlendingColor:   "ffffff",
	Strip:           false,
}

// DefaultConfig returns a copy of the package decode defaults.
// AutoOrientation and DecodeAnimation are true; BlendingColor is "ffffff".
// Limits are all zero (unlimited). Callers mutate the returned value and pass
// it with WithConfig; the package default itself is immutable and safe for
// concurrent New/Open/Decode.
func DefaultConfig() Config {
	return defaultConfig
}

// Option is a functional option applied to a copy of DefaultConfig.
type Option func(*Config)

// WithAutoOrientation enables or disables JPEG EXIF orientation correction.
func WithAutoOrientation(v bool) Option {
	return func(c *Config) { c.AutoOrientation = v }
}

// WithDecodeAnimation enables or disables multi-frame GIF compositing.
func WithDecodeAnimation(v bool) Option {
	return func(c *Config) { c.DecodeAnimation = v }
}

// WithBlendingColor sets the color used when flattening transparency for JPEG/BMP.
// Invalid colors are reported on the Image via Err() at New/Open/Decode time.
func WithBlendingColor(color any) Option {
	return func(c *Config) { c.BlendingColor = color }
}

// WithStrip drops ICC profile bytes before encoding when true.
func WithStrip(v bool) Option {
	return func(c *Config) { c.Strip = v }
}

// WithConfig replaces the working Config with c. Later options still overlay fields.
func WithConfig(c Config) Option {
	return func(dst *Config) { *dst = c }
}

// WithLimits sets decode resource caps for Open, Decode, DecodeBytes,
// DecodeDataURI, New, and FromImage. Zero fields are unlimited.
func WithLimits(l Limits) Option {
	return func(c *Config) { c.Limits = l }
}

// Anchor is a 9-point pivot used by Cover, Crop, Place, and related helpers.
// String literals such as "center" still convert; prefer the exported constants.
type Anchor = modifier.Anchor

const (
	AnchorTopLeft     = modifier.AnchorTopLeft
	AnchorTop         = modifier.AnchorTop
	AnchorTopRight    = modifier.AnchorTopRight
	AnchorLeft        = modifier.AnchorLeft
	AnchorCenter      = modifier.AnchorCenter
	AnchorRight       = modifier.AnchorRight
	AnchorBottomLeft  = modifier.AnchorBottomLeft
	AnchorBottom      = modifier.AnchorBottom
	AnchorBottomRight = modifier.AnchorBottomRight
)

// ParseAnchor canonicalizes a 9-point position name or a documented synonym.
// Unknown values return ErrGeometry; they are never treated as top-left.
func ParseAnchor(s string) (Anchor, error) {
	return modifier.ParseAnchor(s)
}

func resolveAnchor(raw Anchor, fallback Anchor) (Anchor, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return fallback, nil
	}
	return ParseAnchor(string(raw))
}

type geometrySettings struct {
	anchor     Anchor
	background any
	offsetX    int
	offsetY    int
	opacity    int
}

// GeometryOption configures Cover, Contain, Pad, Crop, Fit, and canvas helpers.
type GeometryOption func(*geometrySettings)

// PlaceOption configures Place. WithOffset and WithOpacity are the usual options.
type PlaceOption = GeometryOption

// WithAnchor sets the 9-point pivot. Untyped string literals such as "center"
// still compile; prefer AnchorCenter, AnchorTopLeft, and the other constants.
// Unknown names are stored and reported as Image.Err() when the modifier runs.
func WithAnchor(anchor Anchor) GeometryOption {
	return func(s *geometrySettings) {
		if strings.TrimSpace(string(anchor)) != "" {
			s.anchor = anchor
		}
	}
}

// WithBackground sets the fill color for new canvas pixels.
// Invalid colors are reported as Image.Err() when the modifier runs.
func WithBackground(color any) GeometryOption {
	return func(s *geometrySettings) {
		if color != nil {
			s.background = color
		}
	}
}

// WithOffset shifts the crop origin or Place overlay after the anchor is applied.
func WithOffset(x, y int) GeometryOption {
	return func(s *geometrySettings) {
		s.offsetX = x
		s.offsetY = y
	}
}

// WithOpacity sets Place overlay opacity in the inclusive range 0–100.
// Values outside that range set Image.Err(); they are not clamped.
func WithOpacity(opacity int) PlaceOption {
	return func(s *geometrySettings) {
		s.opacity = opacity
	}
}

func applyGeometryOptions(base geometrySettings, opts []GeometryOption) geometrySettings {
	for _, o := range opts {
		if o != nil {
			o(&base)
		}
	}
	return base
}

// EncodeOptions controls format-specific encoding.
// Quality defaults to 80 when zero. Callers pass at most one EncodeOptions
// value; extra values are an error rather than silently ignored.
//
// Support matrix (unset/zero values are ignored):
//
//	JPEG: Quality (1–100, default 80)
//	PNG, GIF, BMP, TIFF, WebP: no extra options
//
// Progressive, Indexed, Interlaced, Bitdepth, and Quality on non-JPEG
// formats return ErrNotSupported. WebP is lossless VP8L only.
type EncodeOptions struct {
	Quality     int  // JPEG 1–100; 0 means default 80. Unsupported on other formats.
	Progressive bool // unsupported; JPEG stdlib writes baseline
	Indexed     bool // unsupported; PNG palette encoding is not implemented
	Interlaced  bool // unsupported
	Bitdepth    int  // unsupported
}

func (o EncodeOptions) qualityOrDefault() int {
	if o.Quality <= 0 {
		return 80
	}
	if o.Quality > 100 {
		return 100
	}
	return o.Quality
}

func (o EncodeOptions) validate(format Format) error {
	switch format {
	case FormatJPEG:
		if o.Progressive {
			return wrap(ErrNotSupported, "JPEG progressive encoding is not supported")
		}
		if o.Indexed {
			return wrap(ErrNotSupported, "JPEG does not support Indexed")
		}
		if o.Interlaced {
			return wrap(ErrNotSupported, "JPEG does not support Interlaced")
		}
		if o.Bitdepth != 0 {
			return wrap(ErrNotSupported, "JPEG does not support Bitdepth")
		}
		return nil
	case FormatWEBP:
		if o.Quality != 0 {
			return wrap(ErrNotSupported, "WebP encoding is lossless VP8L; Quality is not supported")
		}
	default:
		if o.Quality != 0 {
			return wrap(ErrNotSupported, "%s encoding does not support Quality", format)
		}
	}
	if o.Progressive {
		return wrap(ErrNotSupported, "%s encoding does not support Progressive", format)
	}
	if o.Indexed {
		return wrap(ErrNotSupported, "%s encoding does not support Indexed", format)
	}
	if o.Interlaced {
		return wrap(ErrNotSupported, "%s encoding does not support Interlaced", format)
	}
	if o.Bitdepth != 0 {
		return wrap(ErrNotSupported, "%s encoding does not support Bitdepth", format)
	}
	return nil
}

func applyOptions(opts []Option) Config {
	cfg := DefaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return cfg
}

func applyOptionsChecked(opts []Option) (Config, error) {
	cfg := applyOptions(opts)
	if _, err := ParseColor(cfg.BlendingColor); err != nil {
		return cfg, err
	}
	return cfg, nil
}

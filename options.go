package goimage

import "strings"

// Config holds decode and encode defaults for a single Open/Decode/New/Animate call.
// Values are copied into each Image; mutating a Config after the call does not
// affect in-flight images. Config itself is safe to copy across goroutines.
type Config struct {
	AutoOrientation bool
	DecodeAnimation bool
	BlendingColor   any
	Strip           bool
}

var defaultConfig = Config{
	AutoOrientation: true,
	DecodeAnimation: true,
	BlendingColor:   "ffffff",
	Strip:           false,
}

// DefaultConfig returns a copy of the package decode defaults.
// AutoOrientation and DecodeAnimation are true; BlendingColor is "ffffff".
// Callers mutate the returned value and pass it with WithConfig; the package
// default itself is immutable and safe for concurrent New/Open/Decode.
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

type geometrySettings struct {
	anchor     string
	background any
	offsetX    int
	offsetY    int
}

// GeometryOption configures Cover, Contain, Pad, Crop, Fit, and canvas helpers.
type GeometryOption func(*geometrySettings)

// WithAnchor sets the 9-point pivot (center, top-left, bottom-right, …).
func WithAnchor(anchor string) GeometryOption {
	return func(s *geometrySettings) {
		if strings.TrimSpace(anchor) != "" {
			s.anchor = anchor
		}
	}
}

// WithBackground sets the fill color for new canvas pixels.
func WithBackground(color any) GeometryOption {
	return func(s *geometrySettings) {
		if color != nil {
			s.background = color
		}
	}
}

// WithOffset shifts the crop origin after the anchor is applied.
func WithOffset(x, y int) GeometryOption {
	return func(s *geometrySettings) {
		s.offsetX = x
		s.offsetY = y
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
type EncodeOptions struct {
	Quality     int  // JPEG/WebP 0–100; default 80
	Progressive bool // JPEG (best-effort; stdlib writes baseline)
	Indexed     bool // PNG palette
	Interlaced  bool // PNG/GIF
	Bitdepth    int  // PNG
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

func applyOptions(opts []Option) Config {
	cfg := DefaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return cfg
}

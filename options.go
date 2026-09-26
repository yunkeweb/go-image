package goimage

import "strings"

// Config holds decode and encode defaults for a Manager or a single call.
type Config struct {
	AutoOrientation bool
	DecodeAnimation bool
	BlendingColor   any
	Strip           bool
}

func defaultConfig() Config {
	return Config{
		AutoOrientation: true,
		DecodeAnimation: true,
		BlendingColor:   "ffffff",
		Strip:           false,
	}
}

// Option is a functional option applied to Config.
type Option func(*Config)

func WithAutoOrientation(v bool) Option {
	return func(c *Config) { c.AutoOrientation = v }
}

func WithDecodeAnimation(v bool) Option {
	return func(c *Config) { c.DecodeAnimation = v }
}

func WithBlendingColor(color any) Option {
	return func(c *Config) { c.BlendingColor = color }
}

func WithStrip(v bool) Option {
	return func(c *Config) { c.Strip = v }
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
	cfg := defaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return cfg
}

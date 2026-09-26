package goimage

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

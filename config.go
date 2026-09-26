package goimage

// Config mirrors PHP Intervention\Image\Config.
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

// Option mutates Manager configuration.
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

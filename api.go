package goimage

import (
	"image"
	"io"
)

func applyOptions(opts []Option) Config {
	cfg := defaultConfig()
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	return cfg
}

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

// FromImage wraps a standard-library image.Image as an *Image.
func FromImage(src image.Image, opts ...Option) *Image {
	return result(fromStdImage(src, applyOptions(opts)))
}

// Animate builds a multi-frame GIF. Options apply to the resulting image.
func Animate(init func(*Animation), opts ...Option) *Image {
	return buildAnimation(applyOptions(opts), init)
}

// NewManager returns a Manager that reuses the same Config for Open, Decode, and New.
func NewManager(opts ...Option) *Manager {
	return &Manager{cfg: applyOptions(opts)}
}

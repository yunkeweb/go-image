package goimage

import (
	"encoding/base64"
	"testing"

	"github.com/yunkeweb/go-image/encoder"
)

func fuzzLimits() Limits {
	return Limits{
		MaxInputBytes: 1 << 16,
		MaxWidth:      64,
		MaxHeight:     64,
		MaxPixels:     64 * 64,
		MaxFrames:     4,
	}
}

func FuzzDecodeBytes(f *testing.F) {
	if data, err := New(2, 2).Fill("#112233").ToPNG().Result(); err == nil {
		f.Add(data)
	}
	if data, err := New(2, 2).Fill("#445566").ToJPEG(70).Result(); err == nil {
		f.Add(data)
	}
	f.Add([]byte("GIF89a"))
	f.Add([]byte{0xff, 0xd8, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		img := DecodeBytes(data, WithLimits(fuzzLimits()))
		_ = img.Err()
	})
}

func FuzzDecodeDataURI(f *testing.F) {
	enc := New(2, 2).Fill("#99aabb").ToPNG()
	f.Add(enc.ToDataURI())
	f.Add("data:image/png;base64,not-base64")
	f.Add("data:text/plain;base64,aaaa")
	f.Add("data:image/png,hello%zz")
	f.Add("data:image/png;BASE64," + base64.RawStdEncoding.EncodeToString([]byte("x")))
	f.Fuzz(func(t *testing.T, uri string) {
		img := DecodeDataURI(uri, WithLimits(Limits{
			MaxInputBytes: 1 << 16,
			MaxWidth:      32,
			MaxHeight:     32,
			MaxPixels:     32 * 32,
			MaxFrames:     2,
		}))
		_ = img.Err()
	})
}

func FuzzParseJPEGExif(f *testing.F) {
	if data, err := New(2, 2).Fill("#abcdef").ToJPEG(80).Result(); err == nil {
		f.Add(data)
	}
	f.Add([]byte{0xff, 0xd8, 0xff, 0xe1, 0x00, 0x10, 'E', 'x', 'i', 'f', 0, 0})
	f.Add([]byte{0xff, 0xd8})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = encoder.ParseJPEGExif(data)
	})
}

func FuzzDecodeGIF(f *testing.F) {
	a := New(2, 2).Fill("#ff0000")
	b := New(2, 2).Fill("#00ff00")
	anim := Animate(func(an *Animation) {
		an.Add(a, 0.1).Add(b, 0.1).SetLoops(1)
	})
	if data, err := anim.ToGIF().Result(); err == nil {
		f.Add(data)
	}
	f.Add([]byte("GIF89a"))
	f.Add([]byte("GIF87a"))
	f.Fuzz(func(t *testing.T, data []byte) {
		img := DecodeBytes(data, WithLimits(fuzzLimits()))
		_ = img.Err()
	})
}

func FuzzDecodeWebP(f *testing.F) {
	if data, err := New(2, 2).Fill("#123456").ToWebP().Result(); err == nil {
		f.Add(data)
	}
	f.Add([]byte("RIFF....WEBP"))
	f.Fuzz(func(t *testing.T, data []byte) {
		img := DecodeBytes(data, WithLimits(fuzzLimits()))
		_ = img.Err()
	})
}

func FuzzParseColor(f *testing.F) {
	f.Add("#fff")
	f.Add("#11223344")
	f.Add("rgb(1,2,3)")
	f.Add("rgba(255,0,0,0.5)")
	f.Add("Tomato")
	f.Add("transparent")
	f.Add("")
	f.Add("#gggggg")
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ParseColor(s)
	})
}

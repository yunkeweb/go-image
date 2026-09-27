package goimage

import "testing"

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
		img := DecodeBytes(data, WithLimits(Limits{
			MaxInputBytes: 1 << 16,
			MaxWidth:      64,
			MaxHeight:     64,
			MaxPixels:     64 * 64,
			MaxFrames:     4,
		}))
		_ = img.Err()
	})
}

func FuzzDecodeDataURI(f *testing.F) {
	enc := New(2, 2).Fill("#99aabb").ToPNG()
	f.Add(enc.ToDataURI())
	f.Add("data:image/png;base64,not-base64")
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

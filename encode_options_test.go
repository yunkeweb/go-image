package goimage

import (
	"errors"
	"testing"
)

func TestEncodeOptionsJPEGQuality(t *testing.T) {
	img := New(4, 4).Fill("#445566")
	low := img.Encode(FormatJPEG, EncodeOptions{Quality: 10})
	high := img.Encode(FormatJPEG, EncodeOptions{Quality: 95})
	if low.Err() != nil || high.Err() != nil {
		t.Fatalf("jpeg: %v %v", low.Err(), high.Err())
	}
	if low.Size() == 0 || high.Size() == 0 {
		t.Fatal("empty jpeg")
	}
	def := img.ToJPEG()
	if def.Err() != nil {
		t.Fatal(def.Err())
	}
}

func TestEncodeOptionsUnsupported(t *testing.T) {
	img := New(3, 3).Fill("#112233")
	cases := []struct {
		name string
		enc  EncodedImage
	}{
		{"jpeg progressive", img.Encode(FormatJPEG, EncodeOptions{Progressive: true})},
		{"png quality", img.ToPNG(EncodeOptions{Quality: 80})},
		{"png indexed", img.ToPNG(EncodeOptions{Indexed: true})},
		{"png interlaced", img.ToPNG(EncodeOptions{Interlaced: true})},
		{"png bitdepth", img.ToPNG(EncodeOptions{Bitdepth: 8})},
		{"webp quality", img.ToWebP(EncodeOptions{Quality: 80})},
		{"gif quality", img.ToGIF(EncodeOptions{Quality: 50})},
		{"bmp progressive", img.ToBMP(EncodeOptions{Progressive: true})},
		{"tiff interlaced", img.ToTIFF(EncodeOptions{Interlaced: true})},
	}
	for _, tc := range cases {
		if tc.enc.Err() == nil || !errors.Is(tc.enc.Err(), ErrNotSupported) {
			t.Fatalf("%s: %v", tc.name, tc.enc.Err())
		}
		if uri := tc.enc.ToDataURI(); uri != "" {
			t.Fatalf("%s data uri %q", tc.name, uri)
		}
	}
}

func TestEncodeOptionsZeroValuesOK(t *testing.T) {
	img := New(2, 2).Fill("#abcdef")
	png := img.ToPNG(EncodeOptions{Indexed: false})
	if png.Err() != nil {
		t.Fatal(png.Err())
	}
	webp := img.ToWebP()
	if webp.Err() != nil {
		t.Fatal(webp.Err())
	}
	jpeg := img.Encode(FormatJPEG, EncodeOptions{})
	if jpeg.Err() != nil {
		t.Fatal(jpeg.Err())
	}
}

package goimage

import "testing"

func BenchmarkDecodePNG(b *testing.B) {
	data, err := New(64, 64).Fill("#336699").ToPNG().Result()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := DecodeBytes(data)
		if img.Err() != nil {
			b.Fatal(img.Err())
		}
	}
}

func BenchmarkResize(b *testing.B) {
	src := New(128, 96).Fill("#224466")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := src.Clone().Resize(64, 48)
		if img.Err() != nil {
			b.Fatal(img.Err())
		}
	}
}

func BenchmarkClone(b *testing.B) {
	src := New(64, 64).Fill("#101010")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = src.Clone()
	}
}

func BenchmarkEncodeJPEG(b *testing.B) {
	img := New(64, 64).Fill("#778899")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc := img.ToJPEG(80)
		if enc.Err() != nil {
			b.Fatal(enc.Err())
		}
	}
}

func BenchmarkEncodeWebP(b *testing.B) {
	img := New(32, 32).Fill("#aabbcc")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc := img.ToWebP()
		if enc.Err() != nil {
			b.Fatal(enc.Err())
		}
	}
}

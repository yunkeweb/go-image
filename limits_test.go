package goimage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func pngIHDR(w, h uint32) []byte {
	sig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], w)
	binary.BigEndian.PutUint32(ihdr[4:8], h)
	ihdr[8] = 8
	ihdr[9] = 2
	chunkType := []byte("IHDR")
	crc := crc32.ChecksumIEEE(append(append([]byte{}, chunkType...), ihdr...))
	out := append([]byte{}, sig...)
	out = binary.BigEndian.AppendUint32(out, 13)
	out = append(out, chunkType...)
	out = append(out, ihdr...)
	out = binary.BigEndian.AppendUint32(out, crc)
	return out
}

func TestLimitsMaxInputBytes(t *testing.T) {
	data, err := New(4, 4).Fill("#336699").ToPNG().Result()
	if err != nil {
		t.Fatal(err)
	}
	lim := Limits{MaxInputBytes: 8}
	img := DecodeBytes(data, WithLimits(lim))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("bytes: %v", img.Err())
	}
	img = Decode(bytes.NewReader(data), WithLimits(lim))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("reader: %v", img.Err())
	}
}

func TestLimitsMaxWidthHeightPixels(t *testing.T) {
	header := pngIHDR(400, 300)
	img := DecodeBytes(header, WithLimits(Limits{MaxWidth: 100}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("width: %v", img.Err())
	}
	img = DecodeBytes(header, WithLimits(Limits{MaxHeight: 50}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("height: %v", img.Err())
	}
	img = DecodeBytes(header, WithLimits(Limits{MaxPixels: 1000}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("pixels: %v", img.Err())
	}
}

func TestLimitsDoNotAllocateHugeBuffer(t *testing.T) {
	header := pngIHDR(30000, 30000)
	img := DecodeBytes(header, WithLimits(Limits{MaxPixels: 1 << 20}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("huge: %v", img.Err())
	}
}

func TestLimitsMaxFrames(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{LoopCount: 0}
	for i := 0; i < 4; i++ {
		p := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
		g.Image = append(g.Image, p)
		g.Delay = append(g.Delay, 10)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	img := DecodeBytes(buf.Bytes(), WithLimits(Limits{MaxFrames: 2}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("frames: %v", img.Err())
	}
	ok := DecodeBytes(buf.Bytes(), WithLimits(Limits{MaxFrames: 8}))
	if ok.Err() != nil {
		t.Fatal(ok.Err())
	}
	if ok.Count() != 4 {
		t.Fatalf("count %d", ok.Count())
	}
}

func TestLimitsZeroMeansUnlimited(t *testing.T) {
	data, err := New(8, 8).Fill("#000000").ToPNG().Result()
	if err != nil {
		t.Fatal(err)
	}
	img := DecodeBytes(data, WithLimits(Limits{}))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
}

func TestNewRespectsLimits(t *testing.T) {
	img := New(100, 20, WithLimits(Limits{MaxWidth: 50}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("new: %v", img.Err())
	}
}

package encoder

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"

	"golang.org/x/image/webp"
)

func opaqueNRGBA(w, h int) *image.NRGBA {
	n := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 17),
				G: uint8(y * 13),
				B: 0x66,
				A: 255,
			})
		}
	}
	return n
}

func alphaNRGBA(w, h int) *image.NRGBA {
	n := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n.SetNRGBA(x, y, color.NRGBA{
				R: 0xff,
				G: uint8(x * 20),
				B: uint8(y * 20),
				A: uint8((x + y + 1) * 17),
			})
		}
	}
	return n
}

func toNRGBA(im image.Image) *image.NRGBA {
	b := im.Bounds()
	n := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(n, n.Bounds(), im, b.Min, draw.Src)
	return n
}

func assertRoundTrip(t *testing.T, src *image.NRGBA) {
	t.Helper()
	data, err := EncodeWebP(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 20 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		t.Fatalf("bad riff header %q", data[:min(12, len(data))])
	}
	payload := binaryU32(data[16:])
	if payload%2 == 1 && data[len(data)-1] != 0 {
		t.Fatal("odd VP8L chunk missing RIFF pad byte")
	}

	lib, format, err := DecodeStill(data)
	if err != nil {
		t.Fatalf("self decode: %v", err)
	}
	if format != "webp" {
		t.Fatalf("format %q", format)
	}
	got := toNRGBA(lib)
	assertPixels(t, src, got)

	std, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("x/image/webp decode: %v", err)
	}
	assertPixels(t, src, toNRGBA(std))
}

func binaryU32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func assertPixels(t *testing.T, want, got *image.NRGBA) {
	t.Helper()
	wb, gb := want.Bounds(), got.Bounds()
	if wb.Dx() != gb.Dx() || wb.Dy() != gb.Dy() {
		t.Fatalf("size want %dx%d got %dx%d", wb.Dx(), wb.Dy(), gb.Dx(), gb.Dy())
	}
	for y := 0; y < wb.Dy(); y++ {
		for x := 0; x < wb.Dx(); x++ {
			a := want.NRGBAAt(wb.Min.X+x, wb.Min.Y+y)
			b := got.NRGBAAt(gb.Min.X+x, gb.Min.Y+y)
			if a != b {
				t.Fatalf("pixel %d,%d want %+v got %+v", x, y, a, b)
			}
		}
	}
}

func TestWebPRoundTripOpaque(t *testing.T) {
	assertRoundTrip(t, opaqueNRGBA(8, 8))
}

func TestWebPRoundTripAlpha(t *testing.T) {
	assertRoundTrip(t, alphaNRGBA(6, 5))
}

func TestWebPRoundTripSinglePixel(t *testing.T) {
	n := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	n.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	assertRoundTrip(t, n)
}

func TestWebPRoundTripNonZeroBounds(t *testing.T) {
	full := image.NewNRGBA(image.Rect(0, 0, 20, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 20; x++ {
			full.SetNRGBA(x, y, color.NRGBA{R: 1, G: 2, B: 3, A: 255})
		}
	}
	for y := 4; y < 12; y++ {
		for x := 3; x < 11; x++ {
			full.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 8),
				G: uint8(y * 9),
				B: 0xaa,
				A: 200,
			})
		}
	}
	sub := full.SubImage(image.Rect(3, 4, 11, 12)).(*image.NRGBA)
	assertRoundTrip(t, sub)
}

func TestWebPRoundTripLarger(t *testing.T) {
	assertRoundTrip(t, opaqueNRGBA(64, 48))
}

func TestWebPCorruptAndTruncated(t *testing.T) {
	src := opaqueNRGBA(4, 4)
	ok, err := EncodeWebP(src)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := DecodeStill(ok[:12]); err == nil {
		t.Fatal("expected truncated decode error")
	}
	if _, err := webp.Decode(bytes.NewReader(ok[:8])); err == nil {
		t.Fatal("expected x/image/webp truncated error")
	}
	if _, _, err := DecodeStill([]byte("RIFF....WEBP")); err == nil {
		t.Fatal("expected invalid payload error")
	}
	if _, err := EncodeWebP(nil); err == nil {
		t.Fatal("expected nil source error")
	}
}

func FuzzVP8LRoundTrip(f *testing.F) {
	f.Add(uint8(2), uint8(2), []byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, w8, h8 uint8, pix []byte) {
		w, h := int(w8%8+1), int(h8%8+1)
		src := image.NewNRGBA(image.Rect(0, 0, w, h))
		for i := 0; i < w*h; i++ {
			off := i * 4
			c := color.NRGBA{A: 255}
			if off+3 < len(pix) {
				c = color.NRGBA{R: pix[off], G: pix[off+1], B: pix[off+2], A: pix[off+3]}
			}
			src.SetNRGBA(i%w, i/w, c)
		}
		data, err := EncodeWebP(src)
		if err != nil {
			t.Fatal(err)
		}
		std, err := webp.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		assertPixels(t, src, toNRGBA(std))
	})
}

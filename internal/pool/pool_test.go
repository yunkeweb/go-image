package pool

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestReleaseDropsHugePix(t *testing.T) {
	n := Acquire(2, 2)
	n.Pix = make([]byte, MaxPooledPix+64)
	n.Stride = 8
	n.Rect = image.Rect(0, 0, 2, 2)
	Release(n)
	got := GetForTest()
	if cap(got) > MaxPooledPix {
		t.Fatalf("pool retained cap=%d", cap(got))
	}
	if cap(got) <= MaxPooledPix {
		PutForTest(got)
	}
}

func TestPutForTestDropsHugePix(t *testing.T) {
	huge := make([]byte, MaxPooledPix+8)
	PutForTest(huge)
	got := GetForTest()
	if cap(got) > MaxPooledPix {
		t.Fatalf("PutForTest retained cap=%d", cap(got))
	}
	PutForTest(got)
}

func TestPixBytesRejectsOverflow(t *testing.T) {
	if _, err := PixBytes(0, 10); err == nil {
		t.Fatal("expected error for zero width")
	}
	if _, err := PixBytes(-3, 10); err == nil {
		t.Fatal("expected error for negative width")
	}
	if _, err := PixBytes(math.MaxInt/2, math.MaxInt/2); err == nil {
		t.Fatal("expected overflow error")
	}
	if Acquire(-1, 8) != nil {
		t.Fatal("Acquire should return nil for negative size")
	}
	if Acquire(math.MaxInt/2, math.MaxInt/2) != nil {
		t.Fatal("Acquire should return nil for overflow")
	}
}

func TestCloneSubimageAndStride(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 6, 4))
	src.Stride = 6*4 + 12 // extra padding
	src.Pix = make([]byte, src.Stride*4)
	for y := 0; y < 4; y++ {
		for x := 0; x < 6; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x + 1), G: uint8(y + 2), B: 9, A: 255})
		}
	}
	sub := src.SubImage(image.Rect(1, 1, 4, 3)).(*image.NRGBA)
	cloned := Clone(sub)
	if cloned == nil {
		t.Fatal("clone")
	}
	if cloned.Bounds() != sub.Bounds() {
		t.Fatalf("bounds %v want %v", cloned.Bounds(), sub.Bounds())
	}
	if &cloned.Pix[0] == &sub.Pix[0] {
		t.Fatal("clone shares backing array")
	}
	for y := 1; y < 3; y++ {
		for x := 1; x < 4; x++ {
			got := cloned.NRGBAAt(x, y)
			want := src.NRGBAAt(x, y)
			if got != want {
				t.Fatalf("pixel %d,%d got %+v want %+v", x, y, got, want)
			}
		}
	}
	cloned.SetNRGBA(1, 1, color.NRGBA{B: 255, A: 255})
	if src.NRGBAAt(1, 1).B == 255 {
		t.Fatal("mutating clone changed source")
	}
}

func TestCloneNonZeroBounds(t *testing.T) {
	src := AcquireRect(image.Rect(10, 20, 12, 23))
	if src == nil {
		t.Fatal("acquire")
	}
	src.SetNRGBA(10, 20, color.NRGBA{R: 7, A: 255})
	src.SetNRGBA(11, 22, color.NRGBA{G: 8, A: 255})
	cloned := Clone(src)
	if cloned.Bounds() != src.Bounds() {
		t.Fatalf("bounds %v", cloned.Bounds())
	}
	if cloned.NRGBAAt(10, 20) != (color.NRGBA{R: 7, A: 255}) {
		t.Fatalf("min %+v", cloned.NRGBAAt(10, 20))
	}
	if cloned.NRGBAAt(11, 22) != (color.NRGBA{G: 8, A: 255}) {
		t.Fatalf("max %+v", cloned.NRGBAAt(11, 22))
	}
}

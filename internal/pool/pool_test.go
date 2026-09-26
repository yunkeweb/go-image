package pool

import (
	"image"
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

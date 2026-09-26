package encoder

import (
	"image"
	"image/color"
	"image/gif"
	"math/rand"
	"testing"

	"github.com/yunkeweb/go-image/internal/pool"
)

func TestMedianCutPaletteCap256(t *testing.T) {
	src := pool.Acquire(40, 40)
	rng := rand.New(rand.NewSource(1))
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			src.SetNRGBA(x, y, color.NRGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}
	pal := MedianCutPalette(src, 256)
	if len(pal) > 256 {
		t.Fatalf("palette len %d", len(pal))
	}
	p := QuantizePaletted(src, 256)
	if len(p.Palette) > 256 {
		t.Fatalf("paletted len %d", len(p.Palette))
	}
	g := &gif.GIF{
		Image:     []*image.Paletted{p},
		Delay:     []int{10},
		Disposal:  []byte{gif.DisposalNone},
		Config:    image.Config{Width: 40, Height: 40, ColorModel: p.Palette},
		LoopCount: 0,
	}
	if _, err := EncodeGIF([]GIFFrame{{Img: src, DelayCS: 10, Disposal: gif.DisposalNone}}, 0, 40, 40); err != nil {
		t.Fatal(err)
	}
	_ = g
}

func TestMedianCutPaletteTransparentSlot(t *testing.T) {
	src := pool.Acquire(8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 30), G: uint8(y * 30), B: 10, A: 255})
		}
	}
	src.SetNRGBA(0, 0, color.NRGBA{A: 0})
	pal := MedianCutPalette(src, 256)
	if len(pal) > 256 {
		t.Fatalf("palette len %d", len(pal))
	}
	found := false
	for _, c := range pal {
		_, _, _, a := c.RGBA()
		if a == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected transparent slot")
	}

	opaque := pool.Acquire(4, 4)
	pool.FillRect(opaque, opaque.Bounds(), color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	opal := MedianCutPalette(opaque, 256)
	for _, c := range opal {
		_, _, _, a := c.RGBA()
		if uint8(a>>8) == 0 && a == 0 {
			t.Fatal("opaque image reserved a transparent slot")
		}
	}
}

func TestMedianCutPalette257Boundary(t *testing.T) {
	src := pool.Acquire(17, 16) // 272 unique opaque colors
	i := 0
	for y := 0; y < 16; y++ {
		for x := 0; x < 17; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(i), G: uint8(255 - i), B: uint8(i / 2), A: 255})
			i++
		}
	}
	pal := MedianCutPalette(src, 256)
	if len(pal) > 256 {
		t.Fatalf("palette len %d", len(pal))
	}
	src.SetNRGBA(0, 0, color.NRGBA{A: 0})
	pal = MedianCutPalette(src, 256)
	if len(pal) > 256 {
		t.Fatalf("transparent palette len %d", len(pal))
	}
}

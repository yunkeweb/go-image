package modifier

import (
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func TestCompositeGIFLogicalCanvasAndSubrect(t *testing.T) {
	pal := color.Palette{color.NRGBA{A: 0}, color.NRGBA{R: 255, A: 255}}
	p := image.NewPaletted(image.Rect(5, 5, 15, 15), pal)
	for y := 5; y < 15; y++ {
		for x := 5; x < 15; x++ {
			p.SetColorIndex(x, y, 1)
		}
	}
	g := &gif.GIF{
		Image:    []*image.Paletted{p},
		Delay:    []int{10},
		Disposal: []byte{gif.DisposalNone},
		Config:   image.Config{Width: 20, Height: 20, ColorModel: pal},
	}
	frames := CompositeGIF(g)
	if len(frames) != 1 {
		t.Fatalf("frames %d", len(frames))
	}
	if frames[0].Img.Bounds() != image.Rect(0, 0, 20, 20) {
		t.Fatalf("canvas %v", frames[0].Img.Bounds())
	}
	if frames[0].Delay != 0.1 {
		t.Fatalf("delay %v", frames[0].Delay)
	}
	if frames[0].Img.NRGBAAt(0, 0).A != 0 {
		t.Fatalf("outside %+v", frames[0].Img.NRGBAAt(0, 0))
	}
	if frames[0].Img.NRGBAAt(5, 5).R != 255 {
		t.Fatalf("inside %+v", frames[0].Img.NRGBAAt(5, 5))
	}
}

func TestCompositeGIFDisposalPrevious(t *testing.T) {
	pal := color.Palette{
		color.NRGBA{A: 0},
		color.NRGBA{R: 255, A: 255},
		color.NRGBA{G: 255, A: 255},
		color.NRGBA{B: 255, A: 255},
	}
	full := image.NewPaletted(image.Rect(0, 0, 8, 8), pal)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			full.SetColorIndex(x, y, 1)
		}
	}
	patch := image.NewPaletted(image.Rect(2, 2, 6, 6), pal)
	for y := 2; y < 6; y++ {
		for x := 2; x < 6; x++ {
			patch.SetColorIndex(x, y, 2)
		}
	}
	corner := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			corner.SetColorIndex(x, y, 3)
		}
	}
	g := &gif.GIF{
		Image:     []*image.Paletted{full, patch, corner},
		Delay:     []int{10, 10, 10},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalNone},
		Config:    image.Config{Width: 8, Height: 8, ColorModel: pal},
		LoopCount: 1,
	}
	frames := CompositeGIF(g)
	if len(frames) != 3 {
		t.Fatalf("frames %d", len(frames))
	}
	if frames[1].Img.NRGBAAt(3, 3).G != 255 {
		t.Fatalf("f1 %+v", frames[1].Img.NRGBAAt(3, 3))
	}
	if frames[2].Img.NRGBAAt(3, 3).R != 255 {
		t.Fatalf("f2 restored %+v", frames[2].Img.NRGBAAt(3, 3))
	}
	if frames[2].Img.NRGBAAt(0, 0).B != 255 {
		t.Fatalf("f2 overlay %+v", frames[2].Img.NRGBAAt(0, 0))
	}
	if frames[1].Dispose != int(gif.DisposalPrevious) {
		t.Fatalf("dispose %d", frames[1].Dispose)
	}
}

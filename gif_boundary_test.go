package goimage

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func mustEncodeGIF(t *testing.T, g *gif.GIF) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func gifPalettedRect(r image.Rectangle, pal color.Palette, idx uint8) *image.Paletted {
	p := image.NewPaletted(r, pal)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.SetColorIndex(x, y, idx)
		}
	}
	return p
}

func TestGIFDisposalVariants(t *testing.T) {
	pal := color.Palette{
		color.NRGBA{A: 0},
		color.NRGBA{R: 255, A: 255},
		color.NRGBA{G: 255, A: 255},
		color.NRGBA{B: 255, A: 255},
	}
	full := gifPalettedRect(image.Rect(0, 0, 8, 8), pal, 1)
	patch := gifPalettedRect(image.Rect(2, 2, 6, 6), pal, 2)
	g := &gif.GIF{
		Image:     []*image.Paletted{full, patch},
		Delay:     []int{10, 20},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalBackground},
		Config:    image.Config{Width: 8, Height: 8, ColorModel: pal},
		LoopCount: 4,
	}
	img := DecodeBytes(mustEncodeGIF(t, g))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Count() != 2 {
		t.Fatalf("count %d", img.Count())
	}
	if img.Loops() != 4 {
		t.Fatalf("loops %d", img.Loops())
	}
	frames := img.Frames()
	if frames[0].Delay != 0.1 || frames[1].Delay != 0.2 {
		t.Fatalf("delay %v %v", frames[0].Delay, frames[1].Delay)
	}

	prev := &gif.GIF{
		Image: []*image.Paletted{
			gifPalettedRect(image.Rect(0, 0, 8, 8), pal, 1),
			gifPalettedRect(image.Rect(2, 2, 6, 6), pal, 2),
			gifPalettedRect(image.Rect(0, 0, 2, 2), pal, 3),
		},
		Delay:     []int{5, 5, 5},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalNone},
		Config:    image.Config{Width: 8, Height: 8, ColorModel: pal},
		LoopCount: 1,
	}
	got := DecodeBytes(mustEncodeGIF(t, prev))
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	if got.Count() != 3 {
		t.Fatalf("prev count %d", got.Count())
	}
}

func TestGIFSubrectEmptyFrameAndRoundTrip(t *testing.T) {
	pal := color.Palette{
		color.NRGBA{A: 0},
		color.NRGBA{R: 255, A: 255},
	}
	sub := gifPalettedRect(image.Rect(4, 4, 12, 12), pal, 1)
	empty := gifPalettedRect(image.Rect(0, 0, 1, 1), pal, 0)
	g := &gif.GIF{
		Image:     []*image.Paletted{sub, empty},
		Delay:     []int{15, 25},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalNone},
		Config:    image.Config{Width: 16, Height: 16, ColorModel: pal},
		LoopCount: 2,
	}
	data := mustEncodeGIF(t, g)
	img := DecodeBytes(data)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 16 || img.Height() != 16 {
		t.Fatalf("canvas %dx%d", img.Width(), img.Height())
	}
	if img.PickColor(4, 4).R != 255 {
		t.Fatalf("subrect %+v", img.PickColor(4, 4))
	}
	enc := img.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	round := DecodeBytes(enc.Bytes())
	if round.Err() != nil {
		t.Fatal(round.Err())
	}
	if round.Count() != img.Count() {
		t.Fatalf("round-trip frames %d vs %d", round.Count(), img.Count())
	}
	if round.Loops() != img.Loops() {
		t.Fatalf("round-trip loops %d vs %d", round.Loops(), img.Loops())
	}
}

func TestGIFAnimatedGeometry(t *testing.T) {
	a := solid(8, 8, Color{R: 255, A: 255})
	b := solid(8, 8, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.Add(a, 0.1).Add(b, 0.2).SetLoops(5)
	})
	if anim.Err() != nil {
		t.Fatal(anim.Err())
	}
	resized := anim.Clone().Resize(4, 4)
	if resized.Err() != nil {
		t.Fatal(resized.Err())
	}
	if resized.Count() != 2 || resized.Width() != 4 || resized.Height() != 4 {
		t.Fatalf("resize %d %dx%d", resized.Count(), resized.Width(), resized.Height())
	}
	cropped := anim.Clone().Crop(3, 3)
	if cropped.Err() != nil {
		t.Fatal(cropped.Err())
	}
	if cropped.Count() != 2 || cropped.Width() != 3 {
		t.Fatalf("crop %d %d", cropped.Count(), cropped.Width())
	}
	rotated := anim.Clone().Rotate(90, "#000000")
	if rotated.Err() != nil {
		t.Fatal(rotated.Err())
	}
	if rotated.Count() != 2 {
		t.Fatalf("rotate %d", rotated.Count())
	}
}

func TestGIFCountFailureDoesNotDecodeAsStill(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{
		Image: []*image.Paletted{gifPalettedRect(image.Rect(0, 0, 4, 4), pal, 1)},
		Delay: []int{10},
	}
	data := mustEncodeGIF(t, g)
	truncated := data[:len(data)/2]
	if len(truncated) < 6 {
		t.Fatal("fixture too small")
	}
	img := DecodeBytes(truncated)
	if img.Err() == nil {
		t.Fatal("truncated GIF must fail")
	}
	if img.Count() != 0 {
		t.Fatalf("must not treat truncated GIF as a still image, count=%d", img.Count())
	}
	corrupt := append([]byte("GIF89a"), bytes.Repeat([]byte{0xff}, 40)...)
	bad := DecodeBytes(corrupt)
	if bad.Err() == nil {
		t.Fatal("corrupt GIF must fail")
	}
	if bad.Count() != 0 {
		t.Fatalf("must not fall back to still decode, count=%d", bad.Count())
	}
}

func TestGIFMaxFramesAndTotalPixels(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{LoopCount: 0}
	for i := 0; i < 4; i++ {
		g.Image = append(g.Image, gifPalettedRect(image.Rect(0, 0, 4, 4), pal, 1))
		g.Delay = append(g.Delay, 10)
	}
	data := mustEncodeGIF(t, g)
	img := DecodeBytes(data, WithLimits(Limits{MaxFrames: 2}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("frames: %v", img.Err())
	}
	img = DecodeBytes(data, WithLimits(Limits{MaxPixels: 4 * 4 * 2}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("pixels: %v", img.Err())
	}
}

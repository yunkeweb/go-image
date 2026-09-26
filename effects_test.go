package goimage

import "testing"

func TestGreyscaleInvertBrightness(t *testing.T) {
	img := solid(4, 4, Color{R: 200, G: 10, B: 10, A: 255}).Greyscale()
	c := img.PickColor(0, 0)
	if c.R != c.G || c.G != c.B {
		t.Fatalf("grey %+v", c)
	}
	inv := solid(2, 2, Color{R: 255, A: 255}).Invert()
	if inv.PickColor(0, 0).R != 0 {
		t.Fatalf("invert %+v", inv.PickColor(0, 0))
	}
	br := solid(2, 2, Color{R: 10, G: 10, B: 10, A: 255}).Brightness(100)
	if br.PickColor(0, 0).R <= 10 {
		t.Fatalf("brightness %+v", br.PickColor(0, 0))
	}
}

func TestContrastGammaColorize(t *testing.T) {
	img := solid(2, 2, Color{R: 100, G: 100, B: 100, A: 255}).Contrast(50)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	g := solid(2, 2, Color{R: 64, G: 64, B: 64, A: 255}).Gamma(2.2)
	if g.Err() != nil {
		t.Fatal(g.Err())
	}
	cz := solid(2, 2, Color{R: 10, G: 10, B: 10, A: 255}).Colorize(20, 0, 0)
	if cz.PickColor(0, 0).R != 30 {
		t.Fatalf("colorize %+v", cz.PickColor(0, 0))
	}
}

func TestFlipFlopRotate(t *testing.T) {
	img := Create(3, 2).Fill("#000000")
	img.DrawPixel(0, 0, "#ff0000")
	flipped := img.Clone().Flip()
	if flipped.PickColor(0, 1).R != 255 {
		t.Fatalf("flip %+v", flipped.PickColor(0, 1))
	}
	flopped := img.Clone().Flop()
	if flopped.PickColor(2, 0).R != 255 {
		t.Fatalf("flop %+v", flopped.PickColor(2, 0))
	}
	rot := img.Clone().Rotate(90, "#00ff00")
	if rot.Width() != 2 || rot.Height() != 3 {
		t.Fatalf("rotate size %dx%d", rot.Width(), rot.Height())
	}
}

func TestBlurSharpenPixelate(t *testing.T) {
	img := Create(16, 16).Fill("#ffffff")
	img.DrawPixel(8, 8, "#000000")
	b := img.Clone().Blur(2)
	if b.Err() != nil {
		t.Fatal(b.Err())
	}
	s := img.Clone().Sharpen(15)
	if s.Err() != nil {
		t.Fatal(s.Err())
	}
	p := img.Clone().Pixelate(4)
	if p.Err() != nil {
		t.Fatal(p.Err())
	}
}

func TestReduceColorsAndBlend(t *testing.T) {
	img := Create(8, 8).Fill("#112233")
	img.DrawPixel(0, 0, "#abcdef")
	q := img.Clone().ReduceColors(4, "transparent")
	if q.Err() != nil {
		t.Fatal(q.Err())
	}
	blend := Create(2, 2).Fill("transparent").BlendTransparency("#ff0000")
	if blend.PickColor(0, 0).R < 200 {
		t.Fatalf("blend %+v", blend.PickColor(0, 0))
	}
}

func TestProfileAndResolution(t *testing.T) {
	img := Create(2, 2).SetResolution(300, 300).SetProfile([]byte("icc"))
	x, y := img.Resolution()
	if x != 300 || y != 300 {
		t.Fatalf("res %v %v", x, y)
	}
	if string(img.Profile()) != "icc" {
		t.Fatal("profile")
	}
	img.RemoveProfile()
	if img.Profile() != nil {
		t.Fatal("profile not removed")
	}
}

package goimage

import "testing"

func TestResize(t *testing.T) {
	img := solid(40, 20, Color{R: 255, A: 255}).Resize(10, 5)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 10 || img.Height() != 5 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
}

func TestResizeAutoHeight(t *testing.T) {
	img := solid(400, 200, Color{R: 255, A: 255}).Resize(100)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 100 || img.Height() != 50 {
		t.Fatalf("want 100x50, got %dx%d", img.Width(), img.Height())
	}
	stretched := solid(400, 200, Color{R: 255, A: 255}).Resize(100, 80)
	if stretched.Width() != 100 || stretched.Height() != 80 {
		t.Fatalf("want 100x80, got %dx%d", stretched.Width(), stretched.Height())
	}
}

func TestScaleKeepsAspect(t *testing.T) {
	img := solid(100, 50, Color{G: 255, A: 255}).Scale(20)
	if img.Width() != 20 || img.Height() != 10 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
	boxed := solid(100, 50, Color{G: 255, A: 255}).Scale(20, 20)
	if boxed.Width() != 20 || boxed.Height() != 10 {
		t.Fatalf("boxed %dx%d", boxed.Width(), boxed.Height())
	}
}

func TestScaleDown(t *testing.T) {
	img := solid(10, 10, ColorWhite).ScaleDown(100, 100)
	if img.Width() != 10 || img.Height() != 10 {
		t.Fatalf("upsize leaked %dx%d", img.Width(), img.Height())
	}
}

func TestCover(t *testing.T) {
	img := solid(100, 50, Color{B: 255, A: 255}).Cover(20, 20, WithAnchor("center"))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 20 || img.Height() != 20 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
	fitted := solid(100, 50, Color{B: 255, A: 255}).Fit(20, 20)
	if fitted.Width() != 20 || fitted.Height() != 20 {
		t.Fatalf("fit %dx%d", fitted.Width(), fitted.Height())
	}
}

func TestContainAndPad(t *testing.T) {
	img := solid(40, 20, Color{R: 255, A: 255}).Contain(40, 40, WithBackground("#00ff00"), WithAnchor("center"))
	if img.Width() != 40 || img.Height() != 40 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
	c := img.PickColor(0, 0)
	if c.G < 200 {
		t.Fatalf("expected green pad, got %+v", c)
	}
	center := img.PickColor(20, 20)
	if center.R < 200 {
		t.Fatalf("expected red content, got %+v", center)
	}

	padded := solid(40, 20, Color{R: 255, A: 255}).Pad(80, 80, WithBackground("blue"))
	if padded.Width() != 80 || padded.Height() != 80 {
		t.Fatalf("pad %dx%d", padded.Width(), padded.Height())
	}
	if padded.PickColor(0, 0).B < 200 {
		t.Fatalf("pad bg %+v", padded.PickColor(0, 0))
	}
}

func TestCrop(t *testing.T) {
	img := Create(10, 10).Fill("#ffffff")
	img.DrawPixel(2, 2, "#ff0000")
	got := img.Crop(3, 3, WithBackground("#000000"), WithAnchor("top-left"))
	if got.Width() != 3 || got.Height() != 3 {
		t.Fatalf("%dx%d", got.Width(), got.Height())
	}
	if got.PickColor(2, 2).R != 255 {
		t.Fatalf("crop pixel %+v", got.PickColor(2, 2))
	}
	shifted := Create(10, 10).Fill("#ffffff")
	shifted.DrawPixel(4, 4, "#00ff00")
	got = shifted.Crop(3, 3, WithOffset(2, 2))
	if got.PickColor(2, 2).G != 255 {
		t.Fatalf("offset crop %+v", got.PickColor(2, 2))
	}
}

func TestResizeCanvas(t *testing.T) {
	img := solid(4, 4, Color{R: 255, A: 255}).ResizeCanvas(8, 8, WithBackground("#0000ff"), WithAnchor("center"))
	if img.Width() != 8 || img.Height() != 8 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
	if img.PickColor(0, 0).B < 200 {
		t.Fatalf("canvas bg %+v", img.PickColor(0, 0))
	}
	if img.PickColor(4, 4).R < 200 {
		t.Fatalf("content %+v", img.PickColor(4, 4))
	}
}

func TestTrim(t *testing.T) {
	img := Create(10, 10).Fill("#ffffff")
	img.DrawPixel(4, 4, "#000000")
	img.DrawPixel(5, 5, "#000000")
	img.Trim(10)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() > 8 || img.Height() > 8 {
		t.Fatalf("trim did not shrink: %dx%d", img.Width(), img.Height())
	}
}

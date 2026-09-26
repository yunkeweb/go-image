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

func TestScaleKeepsAspect(t *testing.T) {
	img := solid(100, 50, Color{G: 255, A: 255}).Scale(20, 0)
	if img.Width() != 20 || img.Height() != 10 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
}

func TestScaleDown(t *testing.T) {
	img := solid(10, 10, ColorWhite).ScaleDown(100, 100)
	if img.Width() != 10 || img.Height() != 10 {
		t.Fatalf("upsize leaked %dx%d", img.Width(), img.Height())
	}
}

func TestCover(t *testing.T) {
	img := solid(100, 50, Color{B: 255, A: 255}).Cover(20, 20, "center")
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 20 || img.Height() != 20 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
}

func TestContainAndPad(t *testing.T) {
	img := solid(40, 20, Color{R: 255, A: 255}).Contain(40, 40, "#00ff00", "center")
	if img.Width() != 40 || img.Height() != 40 {
		t.Fatalf("%dx%d", img.Width(), img.Height())
	}
	// corners should be background green
	c := img.PickColor(0, 0)
	if c.G < 200 {
		t.Fatalf("expected green pad, got %+v", c)
	}
	center := img.PickColor(20, 20)
	if center.R < 200 {
		t.Fatalf("expected red content, got %+v", center)
	}

	padded := solid(40, 20, Color{R: 255, A: 255}).Pad(80, 80, "blue", "center")
	if padded.Width() != 80 || padded.Height() != 80 {
		t.Fatalf("pad %dx%d", padded.Width(), padded.Height())
	}
	// Pad must not upscale the original 40x20 content.
	if padded.PickColor(0, 0).B < 200 {
		t.Fatalf("pad bg %+v", padded.PickColor(0, 0))
	}
}

func TestCrop(t *testing.T) {
	img := Create(10, 10).Fill("#ffffff")
	img.DrawPixel(2, 2, "#ff0000")
	got := img.Crop(3, 3, 0, 0, "#000000", "top-left")
	if got.Width() != 3 || got.Height() != 3 {
		t.Fatalf("%dx%d", got.Width(), got.Height())
	}
	if got.PickColor(2, 2).R != 255 {
		t.Fatalf("crop pixel %+v", got.PickColor(2, 2))
	}
}

func TestResizeCanvas(t *testing.T) {
	img := solid(4, 4, Color{R: 255, A: 255}).ResizeCanvas(8, 8, "#0000ff", "center")
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

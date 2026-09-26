package goimage

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"testing"

	"github.com/yunkeweb/go-image/internal/pool"
)

func assertNRGBA(t *testing.T, got Color, r, g, b, a uint8) {
	t.Helper()
	if got.R != r || got.G != g || got.B != b || got.A != a {
		t.Fatalf("pixel got rgba(%d,%d,%d,%d) want rgba(%d,%d,%d,%d)", got.R, got.G, got.B, got.A, r, g, b, a)
	}
}

func TestGreyscalePixelRGBA(t *testing.T) {
	// y = (R*299 + G*587 + B*114) / 1000
	img := Create(1, 1).Fill(Color{R: 255, A: 255}).Greyscale()
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	assertNRGBA(t, img.PickColor(0, 0), 76, 76, 76, 255)

	img = Create(1, 1).Fill(Color{R: 10, G: 20, B: 30, A: 200}).Greyscale()
	y := uint8((10*299 + 20*587 + 30*114) / 1000)
	assertNRGBA(t, img.PickColor(0, 0), y, y, y, 200)
}

func TestInvertPixelRGBA(t *testing.T) {
	img := Create(1, 1).Fill(Color{R: 10, G: 20, B: 30, A: 128}).Invert()
	assertNRGBA(t, img.PickColor(0, 0), 245, 235, 225, 128)
}

func TestBrightnessPixelRGBA(t *testing.T) {
	img := Create(1, 1).Fill(Color{R: 10, G: 10, B: 10, A: 255}).Brightness(100)
	assertNRGBA(t, img.PickColor(0, 0), 255, 255, 255, 255)
	img = Create(1, 1).Fill(Color{R: 100, G: 100, B: 100, A: 255}).Brightness(-20)
	assertNRGBA(t, img.PickColor(0, 0), 49, 49, 49, 255) // 100 + round(-20*2.55)=100-51
}

func TestColorizePixelRGBA(t *testing.T) {
	img := Create(1, 1).Fill(Color{R: 10, G: 20, B: 30, A: 255}).Colorize(5, -5, 10)
	assertNRGBA(t, img.PickColor(0, 0), 15, 15, 40, 255)
}

func TestRotate90PixelRGBA(t *testing.T) {
	img := Create(2, 1).Fill("#000000")
	img.DrawPixel(0, 0, Color{R: 255, A: 255})
	img.DrawPixel(1, 0, Color{B: 255, A: 255})
	rot := img.Rotate(90, "#00ff00")
	if rot.Width() != 1 || rot.Height() != 2 {
		t.Fatalf("size %dx%d", rot.Width(), rot.Height())
	}
	assertNRGBA(t, rot.PickColor(0, 0), 0, 0, 255, 255)
	assertNRGBA(t, rot.PickColor(0, 1), 255, 0, 0, 255)
}

func TestFlipFlopPixelRGBA(t *testing.T) {
	img := Create(2, 2).Fill("#000000")
	img.DrawPixel(0, 0, Color{R: 255, A: 255})
	flipped := img.Clone().Flip()
	assertNRGBA(t, flipped.PickColor(0, 1), 255, 0, 0, 255)
	assertNRGBA(t, flipped.PickColor(0, 0), 0, 0, 0, 255)
	flopped := img.Clone().Flop()
	assertNRGBA(t, flopped.PickColor(1, 0), 255, 0, 0, 255)
}

func TestBlurUniformStaysUniform(t *testing.T) {
	img := Create(5, 5).Fill(Color{R: 40, G: 80, B: 120, A: 255}).Blur(2)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			assertNRGBA(t, img.PickColor(x, y), 40, 80, 120, 255)
		}
	}
}

func TestBlurSpreadsCenterPixel(t *testing.T) {
	img := Create(3, 3).Fill(Color{A: 255})
	img.DrawPixel(1, 1, Color{R: 255, A: 255})
	out := img.Blur(1)
	center := out.PickColor(1, 1)
	neighbor := out.PickColor(0, 1)
	if center.R == 0 {
		t.Fatalf("center should remain non-zero, got %+v", center)
	}
	if neighbor.R == 0 || neighbor.R > center.R {
		t.Fatalf("neighbor should receive blur, got %+v center %+v", neighbor, center)
	}
}

func TestOrientateResetsExifAndPixels(t *testing.T) {
	img := Create(2, 1).Fill(Color{R: 255, A: 255})
	img.DrawPixel(0, 0, Color{G: 255, A: 255})
	img.SetExif(map[string]any{"Orientation": 2, "IFD0.Orientation": 2})
	img.Orientate()
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.ExifQuery("Orientation") != nil {
		t.Fatalf("orientation %v", img.ExifQuery("Orientation"))
	}
	if _, ok := img.Exif()["IFD0.Orientation"]; ok {
		t.Fatalf("IFD0.Orientation %v", img.ExifQuery("IFD0.Orientation"))
	}
	assertNRGBA(t, img.PickColor(1, 0), 0, 255, 0, 255)
	enc := img.ToJPEG(85)
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	if bytes.Contains(enc.Bytes(), []byte("Exif")) {
		t.Fatal("JPEG should omit EXIF orientation header")
	}
}

func TestResizeZeroDimensions(t *testing.T) {
	img := Create(8, 8).Resize(0, 0).Cover(0, 0)
	if img.Err() == nil || !errors.Is(img.Err(), ErrInvalidDimensions) {
		t.Fatalf("want ErrInvalidDimensions, got %v", img.Err())
	}
	neg := Create(8, 8).Resize(-4)
	if neg.Err() == nil || !errors.Is(neg.Err(), ErrInvalidDimensions) {
		t.Fatalf("want ErrInvalidDimensions for negative, got %v", neg.Err())
	}
	zeroH := Create(8, 8).Resize(8, 0)
	if zeroH.Err() == nil || !errors.Is(zeroH.Err(), ErrInvalidDimensions) {
		t.Fatalf("want ErrInvalidDimensions for zero height, got %v", zeroH.Err())
	}
}

func TestToJPEGQualityShortcut(t *testing.T) {
	src := solid(6, 6, Color{R: 200, G: 30, B: 40, A: 255})
	enc := src.ToJPEG(85)
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	if enc.MimeType() != "image/jpeg" {
		t.Fatal(enc.MimeType())
	}
	got := DecodeBytes(enc.Bytes())
	c := got.PickColor(3, 3)
	if c.R < 150 || c.B > 80 {
		t.Fatalf("jpeg pixel %+v", c)
	}
}

func TestToJPEGVariadicDefault80(t *testing.T) {
	src := solid(8, 8, Color{R: 200, G: 30, B: 40, A: 255})
	def := src.ToJPEG()
	q80 := src.Encode(FormatJPEG, EncodeOptions{Quality: 80})
	q95 := src.ToJPEG(95)
	if def.Err() != nil || q80.Err() != nil || q95.Err() != nil {
		t.Fatalf("encode err default=%v q80=%v q95=%v", def.Err(), q80.Err(), q95.Err())
	}
	if !bytes.Equal(def.Bytes(), q80.Bytes()) {
		t.Fatal("ToJPEG() should match quality 80")
	}
	if bytes.Equal(q95.Bytes(), def.Bytes()) {
		t.Fatal("ToJPEG(95) should differ from default 80")
	}
}

func TestGIFResizeResetsDisposalAndRect(t *testing.T) {
	a := solid(8, 8, Color{R: 255, A: 255})
	b := solid(8, 8, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1)
	})
	anim.frames[0].OffsetLeft = 3
	anim.frames[0].OffsetTop = 2
	anim.frames[0].Dispose = int(gif.DisposalPrevious)
	resized := anim.Resize(4, 4)
	if resized.Err() != nil {
		t.Fatal(resized.Err())
	}
	for i, f := range resized.Frames() {
		if f.OffsetLeft != 0 || f.OffsetTop != 0 {
			t.Fatalf("frame %d offset %d,%d", i, f.OffsetLeft, f.OffsetTop)
		}
		want := int(gif.DisposalPrevious)
		if i != 0 {
			want = int(gif.DisposalNone)
		}
		if f.Dispose != want {
			t.Fatalf("frame %d dispose %d want %d", i, f.Dispose, want)
		}
		nimg := f.Image().(*image.NRGBA)
		if nimg.Rect.Min.X != 0 || nimg.Rect.Min.Y != 0 {
			t.Fatalf("frame %d rect min %v", i, nimg.Rect.Min)
		}
	}
	enc := resized.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	got := DecodeBytes(enc.Bytes())
	if got.Count() != 2 {
		t.Fatalf("frames %d", got.Count())
	}
	if got.PickColor(0, 0).R < 200 {
		t.Fatalf("frame0 %+v", got.PickColor(0, 0))
	}
	c1 := got.PickColorFrame(0, 0, 1)
	if c1.G < 200 {
		t.Fatalf("frame1 %+v", c1)
	}
}

func TestGIFOpaqueResizeKeepsPreviousDisposal(t *testing.T) {
	a := solid(8, 8, Color{R: 255, A: 255})
	b := solid(8, 8, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1)
	})
	anim.frames[0].Dispose = int(gif.DisposalPrevious)
	resized := anim.Resize(4, 4)
	if resized.Err() != nil {
		t.Fatal(resized.Err())
	}
	if resized.frames[0].Dispose != int(gif.DisposalPrevious) {
		t.Fatalf("frame0 dispose %d want Previous", resized.frames[0].Dispose)
	}
	if resized.frames[1].Dispose != int(gif.DisposalNone) {
		t.Fatalf("frame1 dispose %d want None", resized.frames[1].Dispose)
	}
}

func TestPoolReuseAfterChain(t *testing.T) {
	img := Create(16, 16).Fill(Color{R: 12, G: 34, B: 56, A: 255}).
		Cover(8, 8).
		Sharpen(8).
		Greyscale()
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	c := img.PickColor(0, 0)
	if c.R != c.G || c.G != c.B || c.A != 255 {
		t.Fatalf("chained pixel %+v", c)
	}
}

func TestMapPixelsKeepsAlpha(t *testing.T) {
	img := Create(1, 1).Fill(Color{R: 80, G: 80, B: 80, A: 90}).Invert()
	assertNRGBA(t, img.PickColor(0, 0), 175, 175, 175, 90)
}

func TestGIFOpaqueEncodeUsesDisposalNone(t *testing.T) {
	a := solid(6, 6, Color{R: 255, A: 255})
	b := solid(6, 6, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1)
	})
	enc := anim.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	g, err := gif.DecodeAll(bytes.NewReader(enc.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for i, d := range g.Disposal {
		if d != gif.DisposalNone {
			t.Fatalf("opaque frame %d disposal %d want None", i, d)
		}
	}
	got := DecodeBytes(enc.Bytes())
	if got.PickColor(0, 0).R < 200 {
		t.Fatalf("frame0 %+v", got.PickColor(0, 0))
	}
	if got.PickColorFrame(0, 0, 1).G < 200 {
		t.Fatalf("frame1 %+v", got.PickColorFrame(0, 0, 1))
	}
}

func TestGIFTransparentEncodeUsesDisposalBackground(t *testing.T) {
	a := Create(6, 6).Fill(Color{R: 255, A: 128})
	b := Create(6, 6).Fill(Color{G: 255, A: 128})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1)
	})
	enc := anim.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	g, err := gif.DecodeAll(bytes.NewReader(enc.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for i, d := range g.Disposal {
		if d != gif.DisposalBackground {
			t.Fatalf("transparent frame %d disposal %d want Background", i, d)
		}
	}
}

func TestGIFPartialFrameResetsDisposal(t *testing.T) {
	a := solid(8, 8, Color{R: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(a, 0.1)
	})
	anim.frames[1].OffsetLeft = 2
	anim.frames[1].OffsetTop = 2
	anim.frames[1].img = pool.Acquire(4, 4)
	pool.FillRect(anim.frames[1].img, anim.frames[1].img.Bounds(), Color{G: 255, A: 255}.NRGBA())
	anim.resetGIFFrameLayout()
	if anim.frames[1].OffsetLeft != 0 || anim.frames[1].OffsetTop != 0 {
		t.Fatalf("partial offset %d,%d", anim.frames[1].OffsetLeft, anim.frames[1].OffsetTop)
	}
	if anim.frames[1].Dispose != int(gif.DisposalBackground) {
		t.Fatalf("partial dispose %d", anim.frames[1].Dispose)
	}
	if anim.frames[0].Dispose != int(gif.DisposalNone) {
		t.Fatalf("opaque full dispose %d", anim.frames[0].Dispose)
	}
}

func TestPixelAtMatchesNativeNRGBA(t *testing.T) {
	img := Create(2, 2).Fill(Color{R: 1, G: 2, B: 3, A: 4})
	n, ok := img.Native().(*image.NRGBA)
	if !ok {
		t.Fatal("native")
	}
	got := n.NRGBAAt(0, 0)
	if got.R != 1 || got.G != 2 || got.B != 3 || got.A != 4 {
		t.Fatalf("%+v", got)
	}
}

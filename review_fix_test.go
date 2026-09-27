package goimage

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"sync"
	"testing"

	"github.com/yunkeweb/go-image/internal/pool"
)

func TestImageImplementsStdImage(t *testing.T) {
	var _ image.Image = (*Image)(nil)
	img := Create(3, 2).Fill(Color{R: 10, G: 20, B: 30, A: 255})
	if img.ColorModel() != color.NRGBAModel {
		t.Fatal("color model")
	}
	if img.Bounds() != image.Rect(0, 0, 3, 2) {
		t.Fatalf("bounds %v", img.Bounds())
	}
	c, ok := img.At(1, 1).(color.NRGBA)
	if !ok || c.R != 10 || c.G != 20 || c.B != 30 || c.A != 255 {
		t.Fatalf("at %+v", c)
	}
	empty, _ := img.At(-1, 0).(color.NRGBA)
	if empty != (color.NRGBA{}) {
		t.Fatalf("oob %+v", empty)
	}
	var nilImg *Image
	if nilImg.Bounds() != (image.Rectangle{}) {
		t.Fatal("nil bounds")
	}
	if _, ok := nilImg.At(0, 0).(color.NRGBA); !ok {
		t.Fatal("nil at")
	}
	failed := Create(0, 1)
	if failed.Bounds() != (image.Rectangle{}) {
		t.Fatal("failed bounds")
	}
	if _, ok := failed.At(0, 0).(color.NRGBA); !ok {
		t.Fatal("failed at")
	}
}

func TestDefaultConfigIsImmutableCopy(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.AutoOrientation || !cfg.DecodeAnimation {
		t.Fatal("package defaults")
	}
	cfg.AutoOrientation = false
	cfg.DecodeAnimation = false
	again := DefaultConfig()
	if !again.AutoOrientation || !again.DecodeAnimation {
		t.Fatal("DefaultConfig mutated package default")
	}
	img := Create(1, 1, WithConfig(cfg))
	if img.Config().AutoOrientation || img.Config().DecodeAnimation {
		t.Fatal("WithConfig")
	}
}

func TestDefaultConfigConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			img := New(2, 2)
			if img.Err() != nil {
				t.Error(img.Err())
				return
			}
			_ = DefaultConfig()
			enc := img.ToPNG()
			if enc.Err() != nil {
				t.Error(enc.Err())
				return
			}
			got := DecodeBytes(enc.Bytes())
			if got.Err() != nil {
				t.Error(got.Err())
			}
		}()
	}
	wg.Wait()
}

func TestAccessorsReturnCopies(t *testing.T) {
	img := Create(2, 2).Fill(Color{R: 1, G: 2, B: 3, A: 4})
	img.SetExif(map[string]any{
		"Orientation": 3,
		"blob":        []byte{1, 2, 3},
		"nested":      map[string]any{"k": []byte{9}},
	})
	img.SetProfile([]byte("icc"))

	n := img.Native().(*image.NRGBA)
	n.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	if img.PickColor(0, 0).R != 1 {
		t.Fatal("Native mutation leaked")
	}

	frames := img.Frames()
	frames[0].Delay = 9
	clonedFrame := frames[0].Image().(*image.NRGBA)
	clonedFrame.SetNRGBA(0, 0, color.NRGBA{G: 255, A: 255})
	if img.Frames()[0].Delay == 9 {
		t.Fatal("Frames slice mutation leaked")
	}
	if img.PickColor(0, 0).G == 255 {
		t.Fatal("Frame.Image mutation leaked")
	}

	ex := img.Exif()
	ex["Orientation"] = 8
	ex["blob"].([]byte)[0] = 99
	ex["nested"].(map[string]any)["k"].([]byte)[0] = 1
	if img.ExifQuery("Orientation").(int) != 3 {
		t.Fatal("Exif map mutation leaked")
	}
	if img.Exif()["blob"].([]byte)[0] != 1 {
		t.Fatal("Exif []byte mutation leaked")
	}

	p := img.Profile()
	p[0] = 'X'
	if string(img.Profile()) != "icc" {
		t.Fatal("Profile mutation leaked")
	}
}

func TestCloneDeepCopiesExif(t *testing.T) {
	img := Create(1, 1)
	img.SetExif(map[string]any{
		"blob":   []byte{1, 2},
		"slice":  []any{"a", []byte{3}},
		"nested": map[string]any{"inner": []byte{4, 5}},
	})
	cp := img.Clone()
	cp.Exif()["blob"].([]byte)[0] = 9
	img.UnsafeExif()["blob"].([]byte)[0] = 1
	cp.UnsafeExif()["blob"].([]byte)[0] = 7
	if img.Exif()["blob"].([]byte)[0] != 1 {
		t.Fatal("clone []byte shared")
	}
	cp.UnsafeExif()["nested"].(map[string]any)["inner"].([]byte)[0] = 8
	if img.Exif()["nested"].(map[string]any)["inner"].([]byte)[0] != 4 {
		t.Fatal("clone nested map shared")
	}
}

func TestFillAndFloodFillSplit(t *testing.T) {
	img := Create(8, 8).Fill("#ffffff")
	img.DrawRectangle(0, 0, func(d *Drawable) {
		d.Size(8, 8).SetBorder(1, "#000000")
	})
	img.FloodFill(4, 4, "#ff0000")
	c := img.PickColor(4, 4)
	if c.R < 200 {
		t.Fatalf("flood %+v", c)
	}
	border := img.PickColor(0, 0)
	if border.R > 50 {
		t.Fatalf("border %+v", border)
	}
}

func TestEncodeRejectsExtraOptions(t *testing.T) {
	src := Create(2, 2).Fill(Color{R: 1, A: 255})
	enc := src.Encode(FormatPNG, EncodeOptions{}, EncodeOptions{Quality: 10})
	if enc.Err() == nil {
		t.Fatal("expected extra EncodeOptions error")
	}
	enc = src.ToJPEG(80, 90)
	if enc.Err() == nil {
		t.Fatal("expected extra quality error")
	}
}

func TestOrientClearsOrientationKeys(t *testing.T) {
	img := Create(2, 1).Fill(Color{R: 255, A: 255})
	img.DrawPixel(0, 0, Color{G: 255, A: 255})
	img.SetExif(map[string]any{
		"Orientation":      uint16(2),
		"IFD0.Orientation": int64(2),
		"EXIF.Orientation": float64(2),
	})
	img.Orientate()
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.ExifQuery("Orientation") != nil {
		t.Fatalf("orientation still present %v", img.ExifQuery("Orientation"))
	}
	if _, ok := img.Exif()["IFD0.Orientation"]; ok {
		t.Fatal("IFD0.Orientation still present")
	}
	if img.PickColor(1, 0).G != 255 {
		t.Fatalf("pixels %+v", img.PickColor(1, 0))
	}
}

func encodeGIFBytes(t *testing.T, g *gif.GIF) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func palettedRect(r image.Rectangle, pal color.Palette, idx uint8) *image.Paletted {
	p := image.NewPaletted(r, pal)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			p.SetColorIndex(x, y, idx)
		}
	}
	return p
}

func TestGIFSingleFrameUsesLogicalCanvas(t *testing.T) {
	pal := color.Palette{color.NRGBA{A: 0}, color.NRGBA{R: 255, A: 255}}
	p := palettedRect(image.Rect(5, 5, 15, 15), pal, 1)
	g := &gif.GIF{
		Image:    []*image.Paletted{p},
		Delay:    []int{10},
		Disposal: []byte{gif.DisposalNone},
		Config:   image.Config{Width: 20, Height: 20, ColorModel: pal},
	}
	img := DecodeBytes(encodeGIFBytes(t, g))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 20 || img.Height() != 20 {
		t.Fatalf("size %dx%d", img.Width(), img.Height())
	}
	if img.PickColor(5, 5).R < 200 {
		t.Fatalf("subrect %+v", img.PickColor(5, 5))
	}
	outside := img.PickColor(0, 0)
	if outside.R > 50 {
		t.Fatalf("outside %+v", outside)
	}
}

func TestGIFDisposalStateMachine(t *testing.T) {
	pal := color.Palette{
		color.NRGBA{A: 0},
		color.NRGBA{R: 255, A: 255},
		color.NRGBA{G: 255, A: 255},
		color.NRGBA{B: 255, A: 255},
	}
	full := palettedRect(image.Rect(0, 0, 8, 8), pal, 1)
	patch := palettedRect(image.Rect(2, 2, 6, 6), pal, 2)
	prev := palettedRect(image.Rect(0, 0, 2, 2), pal, 3)

	noneGIF := &gif.GIF{
		Image:     []*image.Paletted{full, palettedRect(image.Rect(0, 0, 8, 8), pal, 2)},
		Delay:     []int{5, 7},
		Disposal:  []byte{gif.DisposalNone, gif.DisposalNone},
		Config:    image.Config{Width: 8, Height: 8, ColorModel: pal},
		LoopCount: 2,
	}
	img := DecodeBytes(encodeGIFBytes(t, noneGIF))
	if img.Count() != 2 || img.Loops() != 2 {
		t.Fatalf("none count/loops %d %d", img.Count(), img.Loops())
	}
	if img.PickColor(0, 0).R < 200 {
		t.Fatalf("none f0 %+v", img.PickColor(0, 0))
	}
	if img.PickColorFrame(0, 0, 1).G < 200 {
		t.Fatalf("none f1 %+v", img.PickColorFrame(0, 0, 1))
	}
	if img.Frames()[0].Delay != 0.05 {
		t.Fatalf("delay %v", img.Frames()[0].Delay)
	}

	bgGIF := &gif.GIF{
		Image:    []*image.Paletted{full, patch},
		Delay:    []int{10, 10},
		Disposal: []byte{gif.DisposalNone, gif.DisposalBackground},
		Config:   image.Config{Width: 8, Height: 8, ColorModel: pal},
	}
	img = DecodeBytes(encodeGIFBytes(t, bgGIF))
	if img.PickColorFrame(3, 3, 1).G < 200 {
		t.Fatalf("bg displayed %+v", img.PickColorFrame(3, 3, 1))
	}
	if img.PickColorFrame(0, 0, 1).R < 200 {
		t.Fatalf("bg keep red %+v", img.PickColorFrame(0, 0, 1))
	}

	prevGIF := &gif.GIF{
		Image:    []*image.Paletted{full, patch, prev},
		Delay:    []int{10, 10, 10},
		Disposal: []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalNone},
		Config:   image.Config{Width: 8, Height: 8, ColorModel: pal},
	}
	img = DecodeBytes(encodeGIFBytes(t, prevGIF))
	if img.Count() != 3 {
		t.Fatalf("prev count %d", img.Count())
	}
	if img.PickColorFrame(3, 3, 1).G < 200 {
		t.Fatalf("prev f1 %+v", img.PickColorFrame(3, 3, 1))
	}
	if img.PickColorFrame(3, 3, 2).R < 200 {
		t.Fatalf("prev restored %+v", img.PickColorFrame(3, 3, 2))
	}
	if img.PickColorFrame(0, 0, 2).B < 200 {
		t.Fatalf("prev f2 overlay %+v", img.PickColorFrame(0, 0, 2))
	}
}

func TestGIFGeometryRoundTrip(t *testing.T) {
	a := solid(12, 8, Color{R: 255, A: 255})
	b := solid(12, 8, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1).SetLoops(1)
	})
	for _, fn := range []func(*Image) *Image{
		func(img *Image) *Image { return img.Clone().Crop(6, 4) },
		func(img *Image) *Image { return img.Clone().Resize(6, 4) },
		func(img *Image) *Image { return img.Clone().Rotate(90, "#000000") },
	} {
		out := fn(anim)
		enc := out.ToGIF()
		if enc.Err() != nil {
			t.Fatal(enc.Err())
		}
		got := DecodeBytes(enc.Bytes())
		if got.Err() != nil {
			t.Fatal(got.Err())
		}
		if got.Count() != 2 {
			t.Fatalf("frames %d", got.Count())
		}
		if got.PickColor(0, 0).A == 0 {
			t.Fatal("empty frame after round trip")
		}
	}
}

func TestHighColorGIFEncodes(t *testing.T) {
	img := Create(40, 40)
	n := img.UnsafeNative()
	i := 0
	b := n.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			n.SetNRGBA(x, y, color.NRGBA{R: uint8(i), G: uint8(255 - i), B: uint8(x + y), A: 255})
			i++
		}
	}
	enc := img.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
}

func TestNewRejectsOverflowDimensions(t *testing.T) {
	img := New(1<<30, 8)
	if img.Err() == nil {
		t.Fatal("expected invalid dimensions")
	}
}

func TestFrameImageHidesInternalPointer(t *testing.T) {
	img := Create(2, 2).Fill(Color{R: 4, A: 255})
	f := img.Frames()[0]
	got := f.Image().(*image.NRGBA)
	if &got.Pix[0] == &img.UnsafeNative().Pix[0] {
		t.Fatal("Frame.Image shared Pix")
	}
	_ = pool.HasTransparency(got)
}

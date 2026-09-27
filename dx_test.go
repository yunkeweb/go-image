package goimage

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestParseAnchorConstantsAndSynonyms(t *testing.T) {
	cases := []struct {
		in   string
		want Anchor
	}{
		{"center", AnchorCenter},
		{"middle", AnchorCenter},
		{"top-left", AnchorTopLeft},
		{"left-top", AnchorTopLeft},
		{"bottom-right", AnchorBottomRight},
		{"right-bottom", AnchorBottomRight},
		{"TOP", AnchorTop},
		{"  bottom  ", AnchorBottom},
	}
	for _, tc := range cases {
		got, err := ParseAnchor(tc.in)
		if err != nil {
			t.Fatalf("ParseAnchor(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseAnchor(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestInvalidAnchorSetsErr(t *testing.T) {
	img := New(8, 8).Cover(8, 8, WithAnchor("not-an-anchor"))
	if img.Err() == nil {
		t.Fatal("expected invalid anchor error")
	}
	if !errors.Is(img.Err(), ErrGeometry) {
		t.Fatalf("want ErrGeometry, got %v", img.Err())
	}
	if !strings.Contains(img.Err().Error(), "invalid anchor") {
		t.Fatalf("message %q", img.Err())
	}

	placed := New(8, 8).Fill("#000000")
	mark := New(2, 2).Fill("#ff0000")
	placed.Place(mark, Anchor("centre"))
	if placed.Err() == nil || !errors.Is(placed.Err(), ErrGeometry) {
		t.Fatalf("place invalid anchor: %v", placed.Err())
	}
}

func TestPlaceOptionsOffsetAndOpacity(t *testing.T) {
	base := New(20, 20).Fill("#000000")
	mark := New(4, 4).Fill("#ff0000")
	base.Place(mark, AnchorBottomRight, WithOffset(0, 0), WithOpacity(100))
	if base.Err() != nil {
		t.Fatal(base.Err())
	}
	c := base.PickColor(19, 19)
	if c.R < 200 {
		t.Fatalf("place %+v", c)
	}

	bad := New(8, 8).Fill("#000000")
	bad.Place(mark, AnchorCenter, WithOpacity(101))
	if bad.Err() == nil || !errors.Is(bad.Err(), ErrInput) {
		t.Fatalf("want opacity error, got %v", bad.Err())
	}

	neg := New(8, 8).Fill("#000000")
	neg.Place(mark, AnchorCenter, WithOpacity(-1))
	if neg.Err() == nil {
		t.Fatal("expected negative opacity error")
	}

	std := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			std.SetNRGBA(x, y, color.NRGBA{B: 255, A: 255})
		}
	}
	canvas := New(6, 6).Fill("#000000")
	canvas.Place(std, AnchorTopLeft)
	if canvas.Err() != nil {
		t.Fatal(canvas.Err())
	}
	if canvas.PickColor(0, 0).B < 200 {
		t.Fatalf("stdlib overlay %+v", canvas.PickColor(0, 0))
	}
}

func TestInvalidColorDoesNotFallback(t *testing.T) {
	img := New(4, 4).Fill("not-a-color")
	if img.Err() == nil || !errors.Is(img.Err(), ErrColor) {
		t.Fatalf("fill: %v", img.Err())
	}

	rot := New(4, 4).Fill("#ffffff").Rotate(360, "not-a-color")
	if rot.Err() == nil || !errors.Is(rot.Err(), ErrColor) {
		t.Fatalf("rotate even at 0 deg: %v", rot.Err())
	}

	line := New(8, 8).Fill("#ffffff")
	line.DrawLine(func(d *Drawable) {
		d.Line(0, 0, 4, 0).SetBorder(1, "not-a-color")
	})
	if line.Err() == nil || !errors.Is(line.Err(), ErrColor) {
		t.Fatalf("draw line: %v", line.Err())
	}

	rect := New(8, 8).Fill("#ffffff")
	rect.DrawRectangle(0, 0, func(d *Drawable) {
		d.Size(4, 4).SetBackground("not-a-color")
	})
	if rect.Err() == nil || !errors.Is(rect.Err(), ErrColor) {
		t.Fatalf("draw rect: %v", rect.Err())
	}

	contain := New(8, 8).Fill("#ff0000").Contain(10, 10, WithBackground("not-a-color"))
	if contain.Err() == nil || !errors.Is(contain.Err(), ErrColor) {
		t.Fatalf("contain: %v", contain.Err())
	}

	txt := New(20, 20).Fill("#ffffff")
	txt.Text("Hi", 2, 10, func(f *Font) { f.Color("not-a-color") })
	if txt.Err() == nil || !errors.Is(txt.Err(), ErrColor) {
		t.Fatalf("text: %v", txt.Err())
	}

	blend := New(2, 2, WithBlendingColor("not-a-color"))
	if blend.Err() == nil || !errors.Is(blend.Err(), ErrColor) {
		t.Fatalf("new blending: %v", blend.Err())
	}
}

func TestEncodedImageErrorContract(t *testing.T) {
	failed := New(0, 1)
	enc := failed.ToPNG()
	if enc.Err() == nil {
		t.Fatal("expected encode error from failed image")
	}
	data, err := enc.Result()
	if data != nil || err == nil || !errors.Is(err, enc.Err()) {
		t.Fatalf("Result should return (nil, same Err): data=%v err=%v EncodedImage.Err=%v", data, err, enc.Err())
	}
	if enc.ToDataURI() != "" {
		t.Fatal("ToDataURI should be empty on error")
	}
	if enc.Bytes() != nil || enc.String() != "" || enc.Size() != 0 || enc.MimeType() != "" {
		t.Fatal("failed EncodedImage accessors must be empty")
	}
	n, err := enc.WriteTo(&bytes.Buffer{})
	if err == nil || n != 0 {
		t.Fatalf("WriteTo: n=%d err=%v", n, err)
	}
	if err := enc.Save("out.png"); err == nil {
		t.Fatal("Save should return delayed error")
	}

	ok := New(2, 2).Fill("#112233")
	png := ok.ToPNG()
	data, err = png.Result()
	if err != nil || len(data) == 0 {
		t.Fatalf("png result: %v len=%d", err, len(data))
	}
	var buf bytes.Buffer
	if _, err := png.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected written bytes")
	}

	jpeg := ok.ToJPEG(80)
	if jpeg.Err() != nil {
		t.Fatal(jpeg.Err())
	}
	if _, err := jpeg.Result(); err != nil {
		t.Fatal(err)
	}

	webp := ok.ToWebP()
	if webp.Err() != nil {
		t.Fatal(webp.Err())
	}

	avif := ok.ToAVIF()
	if avif.Err() == nil {
		t.Fatal("expected AVIF unsupported")
	}
	if _, err := avif.WriteTo(&bytes.Buffer{}); err == nil {
		t.Fatal("WriteTo should surface AVIF error")
	}
}

func TestWithAnchorTypedConstant(t *testing.T) {
	img := solid(100, 50, Color{B: 255, A: 255}).Cover(20, 20, WithAnchor(AnchorCenter))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 20 || img.Height() != 20 {
		t.Fatalf("size %dx%d", img.Width(), img.Height())
	}
}

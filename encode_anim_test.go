package goimage

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/image/webp"
)

func TestJPEGRoundTrip(t *testing.T) {
	src := solid(12, 8, Color{R: 200, G: 30, B: 40, A: 255})
	enc := src.ToJPEG(90)
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	if enc.MimeType() != "image/jpeg" {
		t.Fatal(enc.MimeType())
	}
	got := DecodeBytes(enc.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	c := got.PickColor(6, 4)
	if c.R < 150 || c.B > 80 {
		t.Fatalf("jpeg pixel %+v", c)
	}
}

func TestGIFStillAndAnimated(t *testing.T) {
	a := solid(6, 6, Color{R: 255, A: 255})
	b := solid(6, 6, Color{G: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.2).AddImage(b, 0.2).SetLoops(3)
	})
	if anim.Err() != nil {
		t.Fatal(anim.Err())
	}
	if !anim.IsAnimated() || anim.Count() != 2 || anim.Loops() != 3 {
		t.Fatalf("anim %+v %d %d", anim.IsAnimated(), anim.Count(), anim.Loops())
	}
	enc := anim.ToGIF()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	got := DecodeBytes(enc.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	if got.Count() != 2 {
		t.Fatalf("decoded frames %d", got.Count())
	}
}

func TestSliceAndRemoveAnimation(t *testing.T) {
	a := solid(4, 4, Color{R: 255, A: 255})
	b := solid(4, 4, Color{G: 255, A: 255})
	c := solid(4, 4, Color{B: 255, A: 255})
	anim := Animate(func(an *Animation) {
		an.AddImage(a, 0.1).AddImage(b, 0.1).AddImage(c, 0.1)
	})
	sliced := anim.Clone().SliceAnimation(1, 1)
	if sliced.Count() != 1 {
		t.Fatalf("slice %d", sliced.Count())
	}
	if sliced.PickColor(0, 0).G < 200 {
		t.Fatalf("slice color %+v", sliced.PickColor(0, 0))
	}
	removed := anim.Clone().RemoveAnimation("100%")
	if removed.IsAnimated() {
		t.Fatal("still animated")
	}
	if removed.PickColor(0, 0).B < 200 {
		t.Fatalf("remove %+v", removed.PickColor(0, 0))
	}
}

func TestBMPAndTIFF(t *testing.T) {
	src := solid(5, 5, Color{R: 1, G: 2, B: 3, A: 255})
	bmp := src.ToBMP()
	if bmp.Err() != nil {
		t.Fatal(bmp.Err())
	}
	got := DecodeBytes(bmp.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	tif := src.ToTIFF()
	if tif.Err() != nil {
		t.Fatal(tif.Err())
	}
	got = DecodeBytes(tif.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
}

func TestWebPEncodeDecode(t *testing.T) {
	src := solid(7, 5, Color{R: 10, G: 20, B: 30, A: 255})
	enc := src.ToWebP()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	if !bytes.HasPrefix(enc.Bytes(), []byte("RIFF")) {
		t.Fatal("missing RIFF")
	}
	im, err := webp.Decode(bytes.NewReader(enc.Bytes()))
	if err != nil {
		t.Fatalf("x/image/webp decode: %v", err)
	}
	got := FromImage(im)
	c := got.PickColor(2, 2)
	if c.R != 10 || c.G != 20 || c.B != 30 {
		t.Fatalf("webp pixel %+v", c)
	}
}

func TestUnsupportedFormats(t *testing.T) {
	src := solid(2, 2, ColorWhite)
	if src.ToAVIF().Err() == nil || src.ToHEIC().Err() == nil || src.ToJP2().Err() == nil {
		t.Fatal("expected not supported")
	}
}

func TestEncodeByExtensionAndPath(t *testing.T) {
	src := solid(3, 3, Color{R: 9, A: 255})
	enc := src.EncodeByExtension("png")
	if enc.MimeType() != "image/png" {
		t.Fatal(enc.MimeType())
	}
	enc = src.EncodeByPath("out.jpeg")
	if enc.MimeType() != "image/jpeg" {
		t.Fatal(enc.MimeType())
	}
	enc = src.EncodeByMediaType("image/gif")
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
}

func TestParseFormatAliases(t *testing.T) {
	f, err := parseFormat("JPG")
	if err != nil || f != FormatJPEG {
		t.Fatal(f, err)
	}
	f, err = parseFormat("image/x-png")
	if err != nil || f != FormatPNG {
		t.Fatal(f, err)
	}
}

func TestDefaultConfigAndWithConfig(t *testing.T) {
	if !DefaultConfig.AutoOrientation || !DefaultConfig.DecodeAnimation {
		t.Fatal("package defaults")
	}
	cfg := DefaultConfig
	cfg.AutoOrientation = false
	cfg.DecodeAnimation = false
	cfg.BlendingColor = "red"
	img := Create(1, 1, WithConfig(cfg))
	if img.Config().AutoOrientation || img.Config().DecodeAnimation {
		t.Fatal("options")
	}
	if img.BlendingColor().R != 255 {
		t.Fatalf("blend %+v", img.BlendingColor())
	}
}

func TestErrorMessages(t *testing.T) {
	err := DecodeBytes([]byte("nope")).Err()
	if err == nil || !strings.Contains(err.Error(), "decode") && !strings.Contains(err.Error(), "unable") {
		t.Fatalf("err %v", err)
	}
}

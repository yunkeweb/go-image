package goimage

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func solid(w, h int, c Color) *Image {
	img := Create(w, h).Fill(c)
	if img.Err() != nil {
		panic(img.Err())
	}
	return img
}

func TestCreateAndPickColor(t *testing.T) {
	img := Create(20, 10).Fill("#ff0000")
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 20 || img.Height() != 10 {
		t.Fatalf("size %dx%d", img.Width(), img.Height())
	}
	c := img.PickColor(0, 0)
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Fatalf("pixel %+v", c)
	}
}

func TestDelayedErrorStopsChain(t *testing.T) {
	img := Create(0, 10).Resize(50, 50).Greyscale()
	if img.Err() == nil {
		t.Fatal("expected delayed error")
	}
}

func TestReadPNGRoundTrip(t *testing.T) {
	src := solid(8, 8, Color{R: 0, G: 128, B: 255, A: 255})
	enc := src.ToPNG()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	got := DecodeBytes(enc.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	c := got.PickColor(3, 3)
	if c.R != 0 || c.G != 128 || c.B != 255 {
		t.Fatalf("got %+v", c)
	}
	if got.Origin().MediaType != "image/png" {
		t.Fatalf("media %s", got.Origin().MediaType)
	}
}

func TestReadFileAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "in.png")
	src := solid(4, 4, Color{R: 10, G: 20, B: 30, A: 255})
	if err := src.ToPNG().Save(path); err != nil {
		t.Fatal(err)
	}
	got := Open(path)
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
	if got.PickColor(0, 0).R != 10 {
		t.Fatalf("got %+v", got.PickColor(0, 0))
	}
	out := filepath.Join(dir, "out.jpg")
	if got.Save(out).Err() != nil {
		t.Fatal(got.Err())
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}

func TestReadDataURIAndBase64(t *testing.T) {
	src := solid(2, 2, ColorWhite)
	raw := src.ToPNG().Bytes()
	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
	a := DecodeDataURI(uri)
	if a.Err() != nil {
		t.Fatal(a.Err())
	}
	b := DecodeBytes(raw)
	if b.Err() != nil {
		t.Fatal(b.Err())
	}
	if a.Width() != 2 || b.Width() != 2 {
		t.Fatal("decode size")
	}
}

func TestReadGoImage(t *testing.T) {
	n := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	n.SetNRGBA(1, 1, color.NRGBA{R: 9, A: 255})
	img := FromImage(n)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.PickColor(1, 1).R != 9 {
		t.Fatal(img.PickColor(1, 1))
	}
}

func TestCloneIndependence(t *testing.T) {
	a := solid(4, 4, Color{R: 255, A: 255})
	b := a.Clone().Fill("#00ff00")
	if a.PickColor(0, 0).R != 255 {
		t.Fatal("clone mutated original")
	}
	if b.PickColor(0, 0).G != 255 {
		t.Fatal("clone fill failed")
	}
}

func TestCloneNil(t *testing.T) {
	var img *Image
	cp := img.Clone()
	if cp.Err() == nil {
		t.Fatal("expected error on nil Clone")
	}
}

func TestFromImageOwnedBuffer(t *testing.T) {
	ycbcr := image.NewYCbCr(image.Rect(0, 0, 8, 8), image.YCbCrSubsampleRatio420)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			ycbcr.Y[ycbcr.YOffset(x, y)] = 255
			ycbcr.Cb[ycbcr.COffset(x, y)] = 128
			ycbcr.Cr[ycbcr.COffset(x, y)] = 128
		}
	}
	fromY := FromImage(ycbcr)
	if fromY.Err() != nil {
		t.Fatal(fromY.Err())
	}
	c := fromY.PickColor(0, 0)
	if c.R < 250 || c.G < 250 || c.B < 250 || c.A != 255 {
		t.Fatalf("YCbCr white got %+v", c)
	}
	ycbcr.Y[ycbcr.YOffset(0, 0)] = 0
	if fromY.PickColor(0, 0).R < 250 {
		t.Fatal("FromImage shared YCbCr buffer")
	}

	n := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	n.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	fromN := FromImage(n)
	n.SetNRGBA(0, 0, color.NRGBA{R: 99, A: 255})
	if fromN.PickColor(0, 0).R != 10 {
		t.Fatal("FromImage shared NRGBA Pix")
	}

	rgba := image.NewRGBA(image.Rect(0, 0, 2, 2))
	rgba.SetRGBA(0, 0, color.RGBA{B: 200, A: 255})
	fromR := FromImage(rgba)
	if fromR.Err() != nil {
		t.Fatal(fromR.Err())
	}
	if fromR.PickColor(0, 0).B < 190 {
		t.Fatalf("RGBA %+v", fromR.PickColor(0, 0))
	}
	rgba.SetRGBA(0, 0, color.RGBA{R: 1, A: 255})
	if fromR.PickColor(0, 0).B < 190 {
		t.Fatal("FromImage shared RGBA Pix")
	}
}

func TestEncodedDataURI(t *testing.T) {
	enc := solid(1, 1, ColorBlack).ToPNG()
	uri := enc.ToDataURI()
	if !bytes.HasPrefix([]byte(uri), []byte("data:image/png;base64,")) {
		t.Fatalf("uri %s", uri)
	}
}

func TestEncodedWriteTo(t *testing.T) {
	enc := solid(2, 2, Color{R: 1, G: 2, B: 3, A: 255}).ToPNG()
	if enc.Err() != nil {
		t.Fatal(enc.Err())
	}
	var buf bytes.Buffer
	n, err := enc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(len(enc.Data)) || !bytes.Equal(buf.Bytes(), enc.Data) {
		t.Fatalf("n=%d len=%d", n, len(enc.Data))
	}
	if _, err := encodedErr(ErrNotSupported).WriteTo(&buf); err == nil {
		t.Fatal("expected delayed encode error")
	}
	if _, err := enc.WriteTo(nil); err == nil {
		t.Fatal("expected nil writer error")
	}
}

func TestBlendingColor(t *testing.T) {
	img := Create(2, 2).SetBlendingColor("#112233")
	c := img.BlendingColor()
	if c.R != 0x11 || c.G != 0x22 || c.B != 0x33 {
		t.Fatalf("%+v", c)
	}
}

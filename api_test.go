package goimage

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCanvasAndOptions(t *testing.T) {
	img := New(12, 8, WithBlendingColor("red"))
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 12 || img.Height() != 8 {
		t.Fatalf("size %dx%d", img.Width(), img.Height())
	}
	if img.BlendingColor().R != 255 {
		t.Fatalf("blend %+v", img.BlendingColor())
	}
	alias := Create(3, 3)
	if alias.Err() != nil || alias.Width() != 3 {
		t.Fatal(alias.Err())
	}
}

func TestOpenDecodeFromImage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "in.png")
	src := New(4, 4).Fill(Color{R: 10, G: 20, B: 30, A: 255})
	if err := src.ToPNG().Save(path); err != nil {
		t.Fatal(err)
	}
	opened := Open(path)
	if opened.Err() != nil {
		t.Fatal(opened.Err())
	}
	if opened.PickColor(0, 0).R != 10 {
		t.Fatalf("open %+v", opened.PickColor(0, 0))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoded := Decode(bytes.NewReader(data))
	if decoded.Err() != nil {
		t.Fatal(decoded.Err())
	}
	if decoded.PickColor(1, 1).G != 20 {
		t.Fatalf("decode %+v", decoded.PickColor(1, 1))
	}
	n := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	n.SetNRGBA(0, 0, color.NRGBA{B: 255, A: 255})
	wrapped := FromImage(n)
	if wrapped.PickColor(0, 0).B != 255 {
		t.Fatalf("from %+v", wrapped.PickColor(0, 0))
	}
}

func TestNewManagerReusesConfig(t *testing.T) {
	mgr := NewManager(WithAutoOrientation(false), WithDecodeAnimation(false), WithBlendingColor("#00ff00"))
	if mgr.Config().AutoOrientation || mgr.Config().DecodeAnimation {
		t.Fatal("options")
	}
	img := mgr.New(2, 2)
	if img.BlendingColor().G != 255 {
		t.Fatalf("blend %+v", img.BlendingColor())
	}
	enc := New(3, 3).Fill(Color{R: 1, A: 255}).ToPNG()
	got := mgr.DecodeBytes(enc.Bytes())
	if got.Err() != nil {
		t.Fatal(got.Err())
	}
}

func TestOpenMissingFile(t *testing.T) {
	img := Open("no-such-file-goimage-test.png").Cover(10, 10, "center")
	if img.Err() == nil {
		t.Fatal("expected delayed error")
	}
}

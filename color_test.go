package goimage

import "testing"

func TestParseHex(t *testing.T) {
	c, err := ParseColor("#f00")
	if err != nil {
		t.Fatal(err)
	}
	if c != (Color{R: 255, A: 255}) {
		t.Fatalf("got %+v", c)
	}
	c, err = ParseColor("00ff00")
	if err != nil {
		t.Fatal(err)
	}
	if c.G != 255 || c.R != 0 {
		t.Fatalf("got %+v", c)
	}
	c, err = ParseColor("#11223344")
	if err != nil {
		t.Fatal(err)
	}
	if c.A != 0x44 || c.R != 0x11 {
		t.Fatalf("got %+v", c)
	}
}

func TestParseHTMLNameAndTransparent(t *testing.T) {
	c, err := ParseColor("Tomato")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 99 || c.B != 71 {
		t.Fatalf("got %+v", c)
	}
	c, err = ParseColor("transparent")
	if err != nil {
		t.Fatal(err)
	}
	if !c.IsClear() {
		t.Fatalf("expected clear, got %+v", c)
	}
}

func TestParseRGBFunc(t *testing.T) {
	c, err := ParseColor("rgb(10, 20, 30)")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 10 || c.G != 20 || c.B != 30 || c.A != 255 {
		t.Fatalf("got %+v", c)
	}
	c, err = ParseColor("rgba(255, 0, 0, 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	if c.A < 120 || c.A > 130 {
		t.Fatalf("alpha=%d", c.A)
	}
	c, err = ParseColor("rgb(100%, 0%, 0%)")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 {
		t.Fatalf("got %+v", c)
	}
}

func TestColorConversions(t *testing.T) {
	red := Color{R: 255, A: 255}
	h, s, l := red.HSL()
	if h < 0 || h > 1 || s < 0.99 || l < 0.4 || l > 0.6 {
		// hue of pure red is 0
		if s < 0.99 {
			t.Fatalf("HSL %+v %+v %+v", h, s, l)
		}
	}
	if red.ToHex("#") != "#ff0000" {
		t.Fatalf("hex %s", red.ToHex("#"))
	}
	if !red.IsGreyscale() && red.R == red.G {
		t.Fatal("unexpected")
	}
	gray := Color{R: 10, G: 10, B: 10, A: 255}
	if !gray.IsGreyscale() {
		t.Fatal("expected greyscale")
	}
}

func TestInvalidColor(t *testing.T) {
	_, err := ParseColor("not-a-color")
	if err == nil {
		t.Fatal("expected error")
	}
}

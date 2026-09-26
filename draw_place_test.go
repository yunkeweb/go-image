package goimage

import "testing"

func TestDrawPixelRectangleCircle(t *testing.T) {
	img := Create(40, 40).Fill("#ffffff")
	img.DrawPixel(1, 1, "red")
	if img.PickColor(1, 1).R != 255 {
		t.Fatal("pixel")
	}
	img.DrawRectangle(5, 5, func(d *Drawable) {
		d.Size(10, 8).SetBackground("#0000ff")
	})
	if img.PickColor(6, 6).B < 200 {
		t.Fatalf("rect %+v", img.PickColor(6, 6))
	}
	img.DrawCircle(30, 30, func(d *Drawable) {
		d.SetRadius(5).SetBackground("#00ff00")
	})
	if img.PickColor(30, 30).G < 200 {
		t.Fatalf("circle %+v", img.PickColor(30, 30))
	}
}

func TestDrawLinePolygonBezier(t *testing.T) {
	img := Create(50, 50).Fill("#ffffff")
	img.DrawLine(func(d *Drawable) {
		d.Line(0, 0, 40, 0).SetBorder(2, "#ff0000")
	})
	if img.PickColor(10, 0).R < 200 {
		t.Fatalf("line %+v", img.PickColor(10, 0))
	}
	img.DrawPolygon(func(d *Drawable) {
		d.AddPoint(10, 10).AddPoint(20, 10).AddPoint(15, 20).SetBackground("#0000ff")
	})
	if img.PickColor(15, 12).B < 200 {
		t.Fatalf("poly %+v", img.PickColor(15, 12))
	}
	img.DrawBezier(func(d *Drawable) {
		d.AddPoint(0, 40).AddPoint(25, 20).AddPoint(49, 40).SetBorder(1, "#00aa00")
	})
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
}

func TestFloodFill(t *testing.T) {
	img := Create(8, 8).Fill("#ffffff")
	img.DrawRectangle(0, 0, func(d *Drawable) {
		d.Size(8, 8).SetBorder(1, "#000000")
	})
	img.Fill("#ff0000", 4, 4)
	c := img.PickColor(4, 4)
	if c.R < 200 {
		t.Fatalf("flood %+v", c)
	}
}

func TestPlace(t *testing.T) {
	base := Create(20, 20).Fill("#000000")
	mark := Create(4, 4).Fill("#ff0000")
	base.Place(mark, "bottom-right", 0, 0, 100)
	c := base.PickColor(19, 19)
	if c.R < 200 {
		t.Fatalf("place %+v", c)
	}
	base2 := Create(20, 20).Fill("#000000")
	base2.Place(mark, "center", 0, 0, 50)
	if base2.Err() != nil {
		t.Fatal(base2.Err())
	}
}

func TestText(t *testing.T) {
	img := Create(80, 30).Fill("#ffffff")
	img.Text("Hi", 5, 20, func(f *Font) {
		f.Color("#000000").Size(12)
	})
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	// At least one non-white pixel should appear.
	found := false
	for y := 0; y < img.Height(); y++ {
		for x := 0; x < img.Width(); x++ {
			c := img.PickColor(x, y)
			if c.R < 250 || c.G < 250 || c.B < 250 {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected drawn glyphs")
	}
}

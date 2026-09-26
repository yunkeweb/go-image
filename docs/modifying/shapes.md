# Shapes

Rectangle, ellipse, circle, polygon, line, and Bézier share a `func(*Drawable)` callback. Set size, fill, border, and points on the `Drawable`, then the library rasterizes into every frame.

## Overview

`(x, y)` for rectangle is the **top-left**. `(x, y)` for ellipse and circle is the **center**. Polygon and Bézier collect `AddPoint`. Line uses `Line(x1,y1,x2,y2)`. Fill comes from `SetBackground`; stroke from `SetBorder`.

## Signatures

```go
func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawPolygon(init func(*Drawable)) *Image
func (img *Image) DrawLine(init func(*Drawable)) *Image
func (img *Image) DrawBezier(init func(*Drawable)) *Image
```

## Drawable helpers

| Method | Type | Default | Description |
|--------|------|---------|-------------|
| `Size(w, h)` | `int, int` | `0, 0` | Rectangle / ellipse box |
| `SetRadius(r)` | `int` | `0` | Circle radius; ellipse uses it for both axes |
| `SetBackground(c)` | `any` | `nil` (no fill) | Fill color |
| `SetBorder(size, col)` | `int, any` | `0`, `nil` | Stroke width and color |
| `Line(x1,y1,x2,y2)` | `int×4` | `0` | Line endpoints |
| `AddPoint(x, y)` | `int, int` | empty | Polygon / Bézier vertex |

Line and Bézier default border to 1 px black when you omit `SetBorder`.

## Example: package-level card with rectangle and circle

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 180).Fill("#0f172a")
	img.DrawRectangle(20, 20, func(d *goimage.Drawable) {
		d.Size(80, 40).SetBackground("#22c55e").SetBorder(2, "#ffffff")
	})
	img.DrawCircle(240, 90, func(d *goimage.Drawable) {
		d.SetRadius(40).SetBackground("#38bdf8")
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("card.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: polygon, line, and Bézier

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(200, 200, goimage.WithBlendingColor("#ffffff")).
		Fill("#111827")
	img.DrawPolygon(func(d *goimage.Drawable) {
		d.AddPoint(100, 20).AddPoint(180, 160).AddPoint(20, 160).
			SetBackground("#f97316")
	})
	img.DrawLine(func(d *goimage.Drawable) {
		d.Line(20, 20, 180, 20).SetBorder(3, "#e2e8f0")
	})
	img.DrawBezier(func(d *goimage.Drawable) {
		d.AddPoint(20, 180).AddPoint(100, 80).AddPoint(180, 180).
			SetBorder(2, "#38bdf8")
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("shapes.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `init` may be `nil`; the shape then has zero size and is skipped.
- Ellipse radius comes from `Size` (`width/2`, `height/2`) unless `SetRadius` is set.
- For overlays of existing images see [Place](/modifying/place).

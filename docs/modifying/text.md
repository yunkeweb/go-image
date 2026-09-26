# Text

`Text` rasterizes a string at `(x, y)`. Pass `nil` for the built-in `basicfont` face, `func(*Font)` to configure a face, or a `*Font` value from `NewFont`. TTF/OTF files go through `opentype`.

## Overview

`(x, y)` is interpreted with `Align` (`left` / `center` / `right`) and `Valign` (`top` / `middle` / `bottom`). `Wrap` enables word wrap at a pixel width. `Stroke` outlines glyphs (width clamped 0–10). `Angle` rotates the run counter-clockwise.

Missing font files set delayed `ErrFont`.

## Signature

```go
func (img *Image) Text(text string, x, y int, fontInit any) *Image
func NewFont(filename ...string) *Font
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `text` | `string` | required | UTF-8 string |
| `x`, `y` | `int` | required | Anchor point |
| `fontInit` | `any` | `nil` → built-in bitmap | `nil`, `*Font`, `Font`, or `func(*Font)` |
| `Font.Size` | `float64` | `12` | Em size |
| `Font.Color` | `any` | `"000000"` | Fill |
| `Font.Align` | `string` | `"left"` | `left` / `center` / `right` |
| `Font.Valign` | `string` | `"bottom"` | `top` / `middle` / `bottom` |
| `Font.LineHeight` | `float64` | `1.25` | Multiplier for wrapped lines |
| `Font.Wrap` | `int` | `0` (off) | Wrap width in pixels |
| `Font.Stroke` | `any, int` | white, `0` | Outline color and width 0–10 |
| `Font.Angle` | `float64` | `0` | Degrees counter-clockwise |
| `Font.Filename` | `string` | empty | Path to TTF/OTF |

## Example: package-level built-in face

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 80).Fill("#0f172a")
	img.Text("go-image", 16, 50, nil)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("label.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: functional Font callback with TTF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(800, 450)
	img.Text("Summer sale", 400, 400, func(f *goimage.Font) {
		f.Filename("fonts/Inter-Bold.ttf").
			Size(36).
			Color("#ffffff").
			Align("center").
			Valign("bottom").
			Stroke("#000000", 2)
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("sale.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Without a font file, text uses `golang.org/x/image/font/basicfont` (fixed 7×13).
- `NewFont("face.ttf")` is equivalent to `NewFont().Filename("face.ttf")`.
- See [Dynamic Watermark](/cookbook/watermark) for timestamp overlays.

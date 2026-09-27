# Colors

Colors in go-image are 8-bit-per-channel sRGB with a straight alpha (`255` = opaque). `ParseColor` accepts hex, HTML names, `rgb()` / `rgba()` strings, `image/color.Color`, and `goimage.Color` values. Invalid input returns `ErrColor`.

## Overview

The library stores pixels as NRGBA. When a modifier needs a fill (canvas padding, rotate corners, flood fill, JPEG flatten), it runs the argument through `ParseColor`. `WithBackground` and `WithBlendingColor` take the same `any` set.

HTML names (`red`, `steelblue`, `transparent`, …) resolve through an internal table. Hex may be `#rgb`, `#rgba`, `#rrggbb`, or `#rrggbbaa`, with or without `#`.

## ParseColor

```go
func ParseColor(v any) (Color, error)

type Color struct {
	R, G, B, A uint8
}

var (
	ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
	ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
	ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)
```

## Parameters

| Input type | Example | Notes |
|------------|---------|-------|
| `string` hex | `"#38bdf8"`, `"fff"`, `"ff000080"` | 3/4/6/8 digits; `#` optional |
| `string` name | `"red"`, `"transparent"` | case-insensitive HTML names |
| `string` css | `"rgb(16, 185, 129)"`, `"rgba(0,0,0,0.5)"` | channels 0–255 or `%` |
| `goimage.Color` | `goimage.Color{R: 56, G: 189, B: 248, A: 255}` | copied as-is |
| `color.Color` | `color.NRGBA{…}` | 16-bit values shifted down to 8-bit |
| `nil` / unknown | | `ErrColor` |

`WithBlendingColor` default is `"ffffff"`. `ColorTransparent` is white with alpha 0 (GIF/PNG keep alpha; JPEG flattens against the blending color).

## Example: package-level parse and fill

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	c, err := goimage.ParseColor("#0ea5e9")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("R=%d G=%d B=%d A=%d", c.R, c.G, c.B, c.A)

	img := goimage.New(64, 64).Fill("steelblue")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("swatch.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: functional options for blending and canvas fill

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png",
		goimage.WithBlendingColor("#111827"), // used when JPEG flattens alpha
	).Pad(256, 256,
		goimage.WithBackground("transparent"),
		goimage.WithAnchor(goimage.AnchorCenter),
	)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("padded.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `mustColor` (used internally) falls back to white when parsing fails. Public `ParseColor` always returns the error.
- JPEG has no alpha. Encode flattens against `BlendingColor` (default white) unless you call `BlendTransparency` first.
- `PickColor` reads the owned NRGBA, so values match what modifiers wrote.

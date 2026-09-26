# Place

`Place` composites another `*Image` onto the current canvas. Position is a 9-point anchor. `opacity` is 0–100. Offsets are pixels inward from that anchor.

## Overview

Open the overlay with `Open` / `Decode` / `New`, then pass it as `src`. A nil overlay or an overlay that already failed sets `ErrInput` (or the overlay's error) on the destination. The overlay is drawn on every destination frame using its primary buffer.

This is the watermark primitive. See [Dynamic Watermark](/cookbook/watermark) for date stamps and [HTTP Handler](/cookbook/http) for streaming the result.

## Signature

```go
func (img *Image) Place(src *Image, position string, offsetX, offsetY, opacity int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `src` | `*Image` | required | Overlay. Must be non-nil and `Err() == nil` |
| `position` | `string` | required | `center`, `top`, `top-left`, `top-right`, `left`, `right`, `bottom`, `bottom-left`, `bottom-right` |
| `offsetX`, `offsetY` | `int` | required | Inward shift from the anchor, pixels |
| `opacity` | `int` | required | 0–100. `100` is opaque |

## Example: package-level logo watermark

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	logo := goimage.Open("logo.png")
	img := goimage.Open("photo.jpg").
		Cover(1200, 800).
		Place(logo, "bottom-right", 16, 16, 70)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("marked.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: generated badge with options on Open

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	badge := goimage.New(120, 32, goimage.WithBlendingColor("transparent")).
		Fill("#ef4444").
		Text("SALE", 60, 22, func(f *goimage.Font) {
			f.Size(14).Color("#ffffff").Align("center")
		})
	img := goimage.Open("product.jpg").
		Place(badge, "top-left", 12, 12, 100)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("sale.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Scale the overlay before `Place` if you need a smaller mark: `logo.Resize(120)`.
- Opacity `0` makes the overlay invisible; values are not clamped in the signature — keep them in 0–100.

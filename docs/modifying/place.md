# Place

`Place` composites another image onto the current canvas at a 9-point [Anchor](/getting-started/design). Offset and opacity are options, so callers cannot mix up a run of integers.

## Overview

Open the overlay with `Open` / `Decode` / `New`, or pass any `image.Image`. A nil overlay or an overlay `*Image` that already failed sets `ErrInput` (or the overlay's error) on the destination. The overlay is drawn on every destination frame.

Unknown anchors and opacity outside 0–100 set `Image.Err()`. They are never silently clamped or treated as `top-left`.

This is the watermark primitive. See [Dynamic Watermark](/cookbook/watermark) for date stamps and [HTTP Handler](/cookbook/http) for streaming the result.

## Signature

```go
func (img *Image) Place(src image.Image, position Anchor, options ...PlaceOption) *Image

func WithOffset(x, y int) GeometryOption
func WithOpacity(opacity int) PlaceOption
```

Error handling: Check `img.Err()` after the call chain.

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `src` | `image.Image` | required | Overlay. `*Image` must be non-nil and `Err() == nil` |
| `position` | `Anchor` | required | Prefer `AnchorBottomRight`, `AnchorCenter`, … String literals such as `"center"` still compile |
| `WithOffset` | `int, int` | `0, 0` | Inward shift from the anchor, pixels |
| `WithOpacity` | `int` | `100` | 0–100. Values outside that range set `Err()` |

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
		Place(logo, goimage.AnchorBottomRight, goimage.WithOffset(16, 16), goimage.WithOpacity(70))
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
		Place(badge, goimage.AnchorTopLeft, goimage.WithOffset(12, 12))
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
- Opacity `0` makes the overlay invisible. Omit `WithOpacity` for a fully opaque mark.

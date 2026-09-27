# Crop

`Crop` extracts a `width×height` rectangle. The default anchor is `top-left` (unlike Cover's `center`). Out-of-bounds areas are filled with `WithBackground`. `WithOffset` nudges the rectangle after the anchor is applied.

## Overview

Crop does not resample. Pixel size of the result is exactly `width×height`. When the rectangle hangs off the source, new pixels receive the background color (default white).

For “fill this aspect and resample”, use [Cover](/modifying/cover). For changing canvas size without cutting the picture, use [Resize Canvas](/modifying/canvas).

## Signature

```go
func (img *Image) Crop(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor Anchor) GeometryOption
func WithBackground(color any) GeometryOption
func WithOffset(x, y int) GeometryOption
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width`, `height` | `int` | required | Crop size, each `>= 1` |
| `WithAnchor` | `string` | `"top-left"` | 9-point origin of the rectangle |
| `WithBackground` | `any` | `"ffffff"` | Fill for pixels outside the source |
| `WithOffset` | `int, int` | `0, 0` | Added to the anchored origin |

## Example: package-level top-left crop

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Crop(400, 300)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("crop.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: centered crop with offset

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Crop(400, 300,
			goimage.WithAnchor("center"),
			goimage.WithOffset(8, 0), // 8px to the right of center
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("crop-center.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `Crop(0, 100)` sets `ErrInvalidDimensions`.
- Animated GIFs are cropped per frame and rebased to `(0, 0)`.
- See [Avatar Crop](/cookbook/avatar) for a square profile pipeline.

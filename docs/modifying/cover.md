# Cover / Fit

`Cover` fills a `width×height` box and crops overflow. It is the thumbnail operator: every output has the exact size you asked for. `Fit` is an alias of `Cover`. `CoverDown` uses the same crop, then resizes down only.

## Overview

The algorithm picks the largest source rectangle with the target aspect ratio, aligns it with a 9-point anchor (default `center`), then resamples to `width×height`. Extra pixels on the long axis are discarded.

`CoverDown` still crops to the target ratio, but the resample step never enlarges — useful when a 64×64 source must not become a blurry 400×400.

## Signatures

```go
func (img *Image) Cover(width, height int, opts ...GeometryOption) *Image
func (img *Image) CoverDown(width, height int, opts ...GeometryOption) *Image
func (img *Image) Fit(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor Anchor) GeometryOption
func WithOffset(x, y int) GeometryOption
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width`, `height` | `int` | required | Output size, each `>= 1` |
| `WithAnchor` | `string` | `"center"` | `center`, `top`, `top-left`, `top-right`, `left`, `right`, `bottom`, `bottom-left`, `bottom-right` |
| `WithOffset` | `int, int` | `0, 0` | Extra shift after the anchor (pixels) |

Invalid sizes set `ErrInvalidDimensions`.

## Example: package-level centered cover

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(400, 300)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: functional options — top anchor, no upscale

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg", goimage.WithAutoOrientation(true)).
		CoverDown(400, 400, goimage.WithAnchor("top"))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Faces near the top of a portrait often look better with `WithAnchor("top")`.
- `Fit` is identical to `Cover` in this library.
- For letterboxing without crop, use [Contain](/modifying/contain).

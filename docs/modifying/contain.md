# Contain / Pad

`Contain` scales the image to fit inside `width×height` and letterboxes the remainder with `WithBackground`. `Pad` is contain-down: it never enlarges, then pads to the box.

## Overview

The picture keeps its aspect ratio. Empty bands are filled with the background color (default white). `WithAnchor` chooses which edge the picture hugs inside the box (`center` by default).

Contain is the opposite of [Cover](/modifying/cover): Cover discards overflow, Contain adds canvas.

## Signatures

```go
func (img *Image) Contain(width, height int, opts ...GeometryOption) *Image
func (img *Image) Pad(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor Anchor) GeometryOption
func WithBackground(color any) GeometryOption
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width`, `height` | `int` | required | Canvas size, each `>= 1` |
| `WithAnchor` | `string` | `"center"` | 9-point pivot inside the box |
| `WithBackground` | `any` | `"ffffff"` | Letterbox fill; hex, name, or `Color` |

## Example: package-level contain on white

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Contain(800, 800)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("square.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: Pad with functional options

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").
		Pad(512, 256,
			goimage.WithAnchor(goimage.AnchorLeft),
			goimage.WithBackground("#0f172a"),
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("banner.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- JPEG flattens alpha against `BlendingColor`. Use PNG or WebP when the pad color is `transparent`.
- `Pad` on an image already larger than the box scales down first, then pads.

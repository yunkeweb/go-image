# Resize Canvas

`ResizeCanvas` changes the canvas size without resampling the picture. Extra space is filled with `WithBackground`. `ResizeCanvasRelative` adds (or subtracts) pixels to the current size.

## Overview

The existing pixels stay 1:1. A larger canvas reveals background around the image according to `WithAnchor` (default `center`). A smaller canvas clips the image the same way Crop would.

Relative values may be negative: `ResizeCanvasRelative(-20, -20)` shrinks both sides by 20 pixels.

## Signatures

```go
func (img *Image) ResizeCanvas(width, height int, opts ...GeometryOption) *Image
func (img *Image) ResizeCanvasRelative(width, height int, opts ...GeometryOption) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width`, `height` | `int` | required | Absolute canvas size (`ResizeCanvas`) or delta (`Relative`) |
| `WithAnchor` | `string` | `"center"` | Where the old picture sits on the new canvas |
| `WithBackground` | `any` | `"ffffff"` | Color of new pixels |

Absolute width/height must be `>= 1` after the relative delta is applied.

## Example: package-level matte

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		ResizeCanvas(840, 640, goimage.WithBackground("#111827"))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("matted.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: relative padding with options

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("icon.png").
		ResizeCanvasRelative(32, 32,
			goimage.WithAnchor("center"),
			goimage.WithBackground("transparent"),
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("padded-icon.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Relative shrink below 1×1 sets `ErrInvalidDimensions`.
- Prefer [Pad](/modifying/contain) when you also need to scale the picture into the box.

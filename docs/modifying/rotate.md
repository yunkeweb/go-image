# Rotate

`Rotate` turns the image **counter-clockwise** by `angle` degrees. New corners are filled with `background`. The canvas grows to the axis-aligned bounding box of the rotated picture.

## Overview

Positive angles rotate counter-clockwise (`90` stands the image on its left side). Interpolation is performed in NRGBA. Pass `"transparent"` as background when encoding PNG or WebP; JPEG will flatten against `BlendingColor`.

For EXIF-driven upright correction, use [Orient](/modifying/flip) instead of a manual `Rotate(90, ...)`.

## Signature

```go
func (img *Image) Rotate(angle float64, background any) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `angle` | `float64` | required | Degrees counter-clockwise. `0` is a no-op |
| `background` | `any` | required | Fill for new corners; hex, name, or `Color` |

## Example: package-level 90° rotate

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Rotate(90, "#000000")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("rotated.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: free angle with transparent corners

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png", goimage.WithBlendingColor("ffffff")).
		Rotate(15, "transparent")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("tilt.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `Rotate` on every GIF frame rebases layout to `(0, 0)`.
- Unparseable `background` falls back through `ParseColor`; prefer a valid hex string.

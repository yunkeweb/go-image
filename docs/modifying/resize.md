# Resize

`Resize` stretches the image to an exact size. Catmull-Rom resampling (`golang.org/x/image/draw`) writes a new owned NRGBA. Omit height to keep the original aspect ratio. `ResizeDown` is the same transform that never enlarges.

## Overview

Given source size `Sw×Sh` and target `Tw×Th`:

- `Resize(Tw, Th)` maps the full source onto `Tw×Th`. Aspect ratio may change.
- `Resize(Tw)` computes `Th = round(Tw * Sh / Sw)` so the ratio is preserved.
- `ResizeDown` leaves the image unchanged when it already fits inside the target.

Zero or negative sizes set delayed `ErrInvalidDimensions`. `0` is never “auto”.

## Signatures

```go
func (img *Image) Resize(width int, height ...int) *Image
func (img *Image) ResizeDown(width int, height ...int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width` | `int` | required | Target width, must be `>= 1` |
| `height` | `...int` | computed from aspect ratio | When passed, must be `>= 1` |

## Example: package-level width-only resize

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	// 400×200 source → Resize(100) yields 100×50
	img := goimage.Open("photo.jpg").Resize(100)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("wide.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: explicit box and resize-down

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		ResizeDown(1920, 1080) // no-op if already smaller
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("hd.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `Resize(400, 300)` stretches. Use [Scale](/modifying/scale) to fit inside a box, or [Cover](/modifying/cover) to fill and crop.
- Animated GIFs are resized frame by frame and rebased to `(0, 0)`.
- See [Error Handling](/concepts/errors) for `errors.Is(err, goimage.ErrInvalidDimensions)`.

# Scale

`Scale` fits the image inside a box while keeping aspect ratio. The result may be smaller than the box on one axis. `ScaleDown` never enlarges. Both resample with Catmull-Rom.

## Overview

Unlike `Resize(w, h)`, which stretches to both sides, `Scale(w, h)` chooses the larger scale factor that still fits. `Scale(w)` is width-only and matches `Resize(w)` for stills that keep ratio.

Use Scale when you need “max 800×600” without letterboxing. For letterboxing see [Contain](/modifying/contain). For fill-and-crop see [Cover](/modifying/cover).

## Signatures

```go
func (img *Image) Scale(width int, height ...int) *Image
func (img *Image) ScaleDown(width int, height ...int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `width` | `int` | required | Box width, `>= 1` |
| `height` | `...int` | computed from aspect ratio | Box height when passed, `>= 1` |

Zero or negative values set `ErrInvalidDimensions`.

## Example: package-level max-width scale

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Scale(800) // height follows ratio
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("scaled.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: fit inside a box, never enlarge

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("icon.png", goimage.WithAutoOrientation(true)).
		ScaleDown(256, 256) // small icons stay small
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("icon-256.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- The output size equals the box only when the source ratio matches `width:height`.
- Animated frames are scaled independently, then rebased to `(0, 0)`.

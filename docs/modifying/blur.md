# Blur

`Blur` runs a box blur of radius `amount`. `amount <= 0` is a no-op. Larger radii cost more CPU; blur after [Cover](/modifying/cover) when you only need a thumbnail.

## Overview

The filter averages a square neighborhood in NRGBA. It is a true box blur (not Gaussian). Multiple passes approximate a wider kernel: `Blur(2).Blur(2)` is heavier than a single `Blur(2)`.

## Signature

```go
func (img *Image) Blur(amount int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `amount` | `int` | required | Box-blur radius in pixels. `<= 0` is a no-op |

## Example: package-level soft blur

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Blur(4)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("soft.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: blur a downscaled backdrop

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(800, 450, goimage.WithAnchor(goimage.AnchorCenter)).
		Blur(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("backdrop.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Blur every GIF frame independently; disposal is recomputed on encode.
- For privacy redaction prefer [Pixelate](/modifying/pixelate) or a solid [Fill](/modifying/fill) rectangle.

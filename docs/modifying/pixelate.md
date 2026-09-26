# Pixelate

`Pixelate` replaces each `size×size` block with its average color. `size <= 1` is a no-op.

## Overview

The grid is aligned to the top-left. Remainder strips on the right and bottom use the same block size clipped to the image bounds. Alpha is averaged with the color.

## Signature

```go
func (img *Image) Pixelate(size int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `size` | `int` | required | Block size in pixels. `<= 1` leaves the image unchanged |

## Example: package-level mosaic

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Pixelate(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("mosaic.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: pixelate a Cover thumbnail

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, goimage.WithAnchor("center")).
		Pixelate(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("preview.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Large `size` on small images yields a handful of flat squares.
- This is not encryption; it is a visual effect.

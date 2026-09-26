# Trim

`Trim` shrinks the canvas to the bounding box of pixels that differ from the corner color. `tolerance` is a 0–100 color distance. Animated images return `ErrNotSupported`.

## Overview

The reference color is the pixel at `(0, 0)`. Every pixel whose distance to that color is within `tolerance` is treated as border and stripped. Useful after a screenshot or a scanned page with a uniform margin.

## Signature

```go
func (img *Image) Trim(tolerance int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `tolerance` | `int` | required | Color distance 0–100 against the corner pixel. `0` trims only exact matches |

## Example: package-level tight trim

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("scan.png").Trim(0)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("tight.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: tolerant trim after a white pad

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").
		Pad(400, 400, goimage.WithBackground("#ffffff")).
		Trim(8) // allow slight JPEG/PNG noise in the white margin
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("logo-trim.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- GIF animations cannot be trimmed (`ErrNotSupported`).
- A fully uniform image trims to a 1×1 (or empty-border) result depending on the matcher; check `Err()` and `Size()` after the call.

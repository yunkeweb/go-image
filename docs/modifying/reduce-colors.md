# Reduce Colors

`ReduceColors` quantizes the image to at most `limit` palette entries with a median-cut palette. `limit < 1` is treated as `1`. The `background` argument is accepted for API symmetry; quantization runs on the current NRGBA pixels.

## Overview

Median-cut splits the color volume until `limit` boxes remain, then each pixel maps to its box center. Use this before GIF encode when you want a smaller palette than the encoder's default, or to posterize a PNG.

## Signature

```go
func (img *Image) ReduceColors(limit int, background any) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `limit` | `int` | required | Palette size. Values `< 1` become `1` |
| `background` | `any` | unused for the cut | Reserved; pass `nil` or a color |

## Example: package-level 16-color posterize

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").ReduceColors(16, nil)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("poster.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: 256 colors then GIF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(320, 240, goimage.WithAnchor("center")).
		ReduceColors(256, "#ffffff")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToGIF().Save("mini.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- GIF encode already quantizes; an extra `ReduceColors(256, …)` is optional pre-shaping.
- Animation frames are quantized independently and may flicker if palettes diverge.

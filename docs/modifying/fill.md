# Fill

`Fill` paints every pixel of every frame. `FloodFill` replaces the 4-connected region of equal color starting at a seed.

## Overview

Whole-canvas fill is the usual way to paint `New` canvases. Flood fill walks 4-connected neighbors that match the seed pixel. Both paths write every animation frame.

## Signature

```go
func (img *Image) Fill(col any) *Image
func (img *Image) FloodFill(x, y int, col any) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `col` | `any` | required | Fill color |
| `x`, `y` | `int` | required for FloodFill | Seed pixel |

## Example: package-level canvas fill

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 180).Fill("#1e293b")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("bg.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: flood fill a region

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("mask.png").FloodFill(10, 10, "#22c55e")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("flood.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Flood fill stops at any color difference, including anti-aliased edges.
- Invalid `col` sets `ErrColor` and skips the paint.
- `Fill(c, x, y)` is no longer a flood-fill overload; use `FloodFill`.

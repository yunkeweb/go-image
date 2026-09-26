# Fill

`Fill` without extra coordinates paints the entire canvas. Two extra ints run a flood fill from that seed, replacing the connected region of equal color.

## Overview

Whole-canvas fill is the usual way to paint `New` canvases. Flood fill walks 4-connected neighbors that match the seed pixel. Both paths write every animation frame.

## Signature

```go
func (img *Image) Fill(col any, xy ...int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `col` | `any` | required | Fill color |
| `xy` | `...int` | omitted = whole canvas | When `len >= 2`, flood-fill from `(xy[0], xy[1])` |

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
	img := goimage.Open("mask.png").Fill("#22c55e", 10, 10)
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

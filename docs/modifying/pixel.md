# Pixel

`DrawPixel` writes one NRGBA pixel at `(x, y)` on every frame. Out-of-bounds coordinates are ignored by the drawer. Invalid colors set delayed `ErrColor`.

## Overview

This is the lowest-level draw call. For filled regions use [Fill](/modifying/fill) or [Shapes](/modifying/shapes). Coordinates are in the current canvas space after any geometry you already applied.

## Signature

```go
func (img *Image) DrawPixel(x, y int, col any) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `x`, `y` | `int` | required | Pixel coordinates, origin top-left |
| `col` | `any` | required | Hex, HTML name, or `Color` |

## Example: package-level marker

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(32, 32).Fill("#0f172a").DrawPixel(16, 16, "#38bdf8")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("dot.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: stamp a grid after Open

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(64, 64)
	for x := 0; x < 64; x += 8 {
		img.DrawPixel(x, 0, "#ef4444")
	}
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("grid.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `PickColor(x, y)` reads the same buffer `DrawPixel` writes.
- Drawing on a failed `Image` is a no-op; check `Err()` after the chain.

# Flip / Flop / Orient

`Flip` mirrors top ↔ bottom. `Flop` mirrors left ↔ right. `Orient` (alias `Orientate`) applies JPEG EXIF orientation 2–8, then stamps orientation `1` so a later JPEG encode cannot rotate twice.

## Overview

`Open` already runs EXIF orientation when `WithAutoOrientation(true)` (the default). Call `Orient` yourself when you decoded with auto-orientation off, or when you built the image via `FromImage`.

JPEG encode writes no EXIF APP1 segment. After `Orient`, viewers see the pixels as stored.

## Signatures

```go
func (img *Image) Flip() *Image
func (img *Image) Flop() *Image
func (img *Image) Orient() *Image
func (img *Image) Orientate() *Image // alias of Orient
```

## Parameters

These methods take no arguments.

| Method | Effect |
|--------|--------|
| `Flip` | Vertical mirror |
| `Flop` | Horizontal mirror |
| `Orient` | EXIF 2–8 → pixels, then tag = Normal |
| `Orientate` | Same as `Orient` |

## Example: package-level flop for a selfie

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("selfie.jpg").Flop()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("mirror.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: decode without auto-orient, then Orient

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(false)).
		Orient()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("upright.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- `Flip` then `Flop` equals a 180° rotate without growing the canvas.
- Missing EXIF is a no-op for `Orient`.

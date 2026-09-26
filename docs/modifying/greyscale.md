# Greyscale

`Greyscale` converts every pixel to luminance and writes that value into R, G, and B. Alpha is unchanged. The transform runs on each animation frame.

## Overview

Luminance follows a Rec. 709-style weighted sum of the sRGB channels. The result is still NRGBA, so later color modifiers (Colorize, Colorize-then-JPEG) keep working.

## Signature

```go
func (img *Image) Greyscale() *Image
```

## Parameters

This method takes no arguments.

| Name | Type | Default | Description |
|------|------|---------|-------------|
| — | — | — | Operates in place on owned NRGBA frames |

## Example: package-level greyscale JPEG

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Greyscale()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("gray.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: chain after Cover

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Cover(400, 300, goimage.WithAnchor("center")).
		Greyscale().
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("gray-thumb.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Already-grey images stay grey; the call is still a full pixel pass.
- Combine with [Contrast](/modifying/contrast) for a punchier black-and-white thumbnail.

# Invert

`Invert` subtracts each RGB channel from 255. Alpha is unchanged. Useful for dark-mode previews and film-negative looks.

## Overview

The mapping is `R' = 255 - R` (same for G and B). It is linear in stored sRGB bytes, not in linear light.

## Signature

```go
func (img *Image) Invert() *Image
```

## Parameters

This method takes no arguments.

## Example: package-level invert

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Invert()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("negative.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: invert a generated canvas

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(128, 128, goimage.WithBlendingColor("#000000")).
		Fill("#f8fafc").
		Invert()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("dark.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Applying Invert twice restores the original RGB.
- Transparent pixels stay transparent (`A` is not inverted).

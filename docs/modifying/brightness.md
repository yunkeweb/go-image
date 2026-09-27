# Brightness

`Brightness` adds a signed offset to every RGB channel. The level is clamped to `-100…100`. Alpha is unchanged.

## Overview

`level` is a percentage of 255: `+100` pushes channels toward white, `-100` toward black. Values outside the range are clamped before the pixel pass.

## Signature

```go
func (img *Image) Brightness(level int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `level` | `int` | required | `-100…100`. `0` is a no-op after clamp |

## Example: package-level lift

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Brightness(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("bright.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: dim a thumbnail after Cover

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, goimage.WithAnchor(goimage.AnchorCenter)).
		Brightness(-15)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("dim.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Pair with [Contrast](/modifying/contrast) rather than stacking large brightness values.
- Animated frames share the same level.

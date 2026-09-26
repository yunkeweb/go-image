# Contrast

`Contrast` scales each RGB channel away from or toward mid-grey. The level is clamped to `-100…100`.

## Overview

Positive values increase separation around 128. Negative values flatten toward grey. The operation is a per-channel affine map in sRGB bytes.

## Signature

```go
func (img *Image) Contrast(level int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `level` | `int` | required | `-100…100`. `0` is a no-op after clamp |

## Example: package-level punch

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Contrast(18)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("punch.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: greyscale then contrast

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Greyscale().
		Contrast(25)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("bw.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Extreme positive contrast clips highlights and shadows.
- Combine with [Brightness](/modifying/brightness) for exposure-style edits.

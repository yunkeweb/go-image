# Colorize

`Colorize` tints RGB channels by signed percentages. Each of `red`, `green`, `blue` is clamped to `-100…100`.

## Overview

Positive red pushes the image toward red; negative red pulls red out. The three axes are independent, so `Colorize(20, -10, 0)` warms the picture slightly.

## Signature

```go
func (img *Image) Colorize(red, green, blue int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `red` | `int` | required | `-100…100` red shift |
| `green` | `int` | required | `-100…100` green shift |
| `blue` | `int` | required | `-100…100` blue shift |

## Example: package-level warm tint

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Colorize(18, 4, -8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("warm.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: colorize after greyscale

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Greyscale().
		Colorize(30, 10, 0) // duotone-style warm grey
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("duo.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Values are clamped; `Colorize(500, 0, 0)` behaves like `100, 0, 0`.
- Alpha is unchanged.

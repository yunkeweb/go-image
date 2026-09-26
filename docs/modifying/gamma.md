# Gamma

`Gamma` applies a power curve to each RGB channel. `gamma` must be `> 0`. Values below 1 darken midtones; values above 1 lighten them.

## Overview

Each channel is mapped as `(c/255)^(1/gamma) * 255`. Alpha is unchanged. Invalid `gamma` (`<= 0`) returns a delayed error from the modifier.

## Signature

```go
func (img *Image) Gamma(gamma float64) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `gamma` | `float64` | required | Must be `> 0`. `1` leaves the image unchanged |

## Example: package-level lift midtones

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Gamma(1.2)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("gamma.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: darken after a canvas fill

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Cover(800, 450).
		Gamma(0.85)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("moody.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Gamma is a global curve; it does not protect highlights the way a tone mapper would.
- Prefer small steps around `1.0` for photographic work.

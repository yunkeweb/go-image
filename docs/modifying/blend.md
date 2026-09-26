# Blend Transparency

`BlendTransparency` composites every pixel over an opaque background. Pass `nil` to use the image's `BlendingColor` (default white). JPEG encode also flattens; call this when you want the flatten visible in PNG as well.

## Overview

The math is standard “over” compositing: `out = src.RGB * src.A + bg.RGB * (1 - src.A)`, then alpha becomes 255. Useful before algorithms that ignore alpha, or before sending a still to a client that cannot display transparency.

## Signature

```go
func (img *Image) BlendTransparency(col any) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `col` | `any` | image `BlendingColor` when `nil` | Hex, name, `Color`, or `nil` |

Invalid colors set delayed `ErrColor`.

## Example: package-level flatten on white

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").BlendTransparency("#ffffff")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(90).Save("logo.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: blending color from Open options

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png", goimage.WithBlendingColor("#0f172a")).
		BlendTransparency(nil) // uses the config color
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("on-slate.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- After this call, `PickColor` reports alpha 255 everywhere.
- GIF frames with partial coverage become opaque; encode may then pick `DisposalNone`.

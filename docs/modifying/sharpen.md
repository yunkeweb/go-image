# Sharpen

`Sharpen` applies an unsharp-mask style boost. `amount <= 0` is a no-op. Typical web thumbnails use `8`–`16` after a downsample.

## Overview

The filter emphasizes local contrast around edges. Too large an amount rings JPEG artifacts. Sharpen **after** geometry (`Cover`, `Resize`) so the kernel sees the final pixel grid.

## Signature

```go
func (img *Image) Sharpen(amount int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `amount` | `int` | required | Unsharp amount. `<= 0` is a no-op |

## Example: package-level thumbnail sharpen

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300).
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: stronger sharpen on greyscale

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Resize(1200).
		Greyscale().
		Sharpen(14)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("crisp.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Sharpen before JPEG encode; the quality setting then compresses the extra edge energy.
- Animated frames are sharpened independently.

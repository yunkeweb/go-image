# Getting started

## Install

```bash
go get github.com/yunkeweb/go-image
```

Requires **Go 1.22+**. The only extra module is official `golang.org/x/image`.

## First encode

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, "center").
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	enc := img.ToJPEG(85)
	if err := enc.Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Create a canvas

```go
img := goimage.New(640, 480).
	Fill("#1e293b").
	DrawCircle(320, 240, func(d *goimage.Drawable) {
		d.SetRadius(80).SetBackground("#38bdf8")
	})
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToPNG().Save("circle.png"); err != nil {
	log.Fatal(err)
}
```

## Functional options

Options apply to a single package-level call:

```go
img := goimage.Open("input.png",
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
```

Reuse one config with `NewManager`:

```go
mgr := goimage.NewManager(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
)
img := mgr.Open("input.png")
```

| Option | Default | Meaning |
|--------|---------|---------|
| `WithAutoOrientation` | `true` | Apply JPEG EXIF orientation 2–8 after decode |
| `WithDecodeAnimation` | `true` | Keep all GIF frames; `false` keeps the first frame |
| `WithBlendingColor` | `"ffffff"` | Color used when flattening transparency (JPEG) |
| `WithStrip` | `false` | Drop ICC profile bytes on encode |

## Typed inputs

| Function | Input |
|----------|--------|
| `Open` | filesystem path |
| `Decode` | `io.Reader` |
| `DecodeBytes` | encoded `[]byte` |
| `DecodeDataURI` | `data:image/...;base64,...` |
| `FromImage` | `image.Image` |
| `New` | blank canvas |

## Next

- [Error model](/guide/errors)
- [Supported formats](/guide/formats)
- [Design](/guide/design)

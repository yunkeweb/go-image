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
	img := goimage.Read("photo.jpg").
		Cover(400, 300, "center").
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	enc := img.ToJPEG(goimage.EncodeOptions{Quality: 85})
	if err := enc.Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Create a canvas

```go
img := goimage.Create(640, 480).
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

## Manager options

```go
mgr := goimage.New(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
img := mgr.Read("input.png")
```

| Option | Default | Meaning |
|--------|---------|---------|
| `WithAutoOrientation` | `true` | Apply JPEG EXIF orientation 2–8 after decode |
| `WithDecodeAnimation` | `true` | Keep all GIF frames; `false` keeps the first frame |
| `WithBlendingColor` | `"ffffff"` | Color used when flattening transparency (JPEG) |
| `WithStrip` | `false` | Drop ICC profile bytes on encode |

Package-level `Create`, `Read`, and `Animate` use a default manager with those defaults.

## Read inputs

`Read` accepts:

- file path (`string` that exists as a file)
- raw bytes (`[]byte`)
- `io.Reader`
- data URI (`data:image/png;base64,...`)
- Base64 string
- `*Image` (clone)
- `image.Image` (stdlib native)

## Next

- [Error model](/guide/errors)
- [Supported formats](/guide/formats)
- [PHP → Go mapping](/guide/migration)

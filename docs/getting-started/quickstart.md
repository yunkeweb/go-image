# Quickstart

This page walks through the three calls you will use most often: `Open` an existing file, `New` a blank canvas, and encode with `ToJPEG` / `ToPNG`. Every mutating method returns the same `*Image`, so chains stay short. Inspect `Err()` once at the end.

## Open, cover, and encode JPEG

`Cover` fills a `400×300` box and crops overflow around the center. `ToJPEG(85)` sets quality to 85; omitting the argument uses 80.

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300). // fill 400×300, crop overflow, Catmull-Rom resample
		Sharpen(8)       // unsharp mask; amount 8 is a light web thumbnail
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	enc := img.ToJPEG(85) // quality 85; ToJPEG() would use 80
	if err := enc.Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Create a canvas with functional options

`New` allocates a transparent NRGBA canvas. `WithBlendingColor` is stored on the image and used when a later encode flattens alpha (JPEG). Drawing uses a `func(*Drawable)` callback.

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(640, 480, goimage.WithBlendingColor("#0f172a")).
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
}
```

## Entry points

| Function | Input | Typical use |
|----------|--------|-------------|
| `Open(path, opts...)` | filesystem path | files on disk |
| `Decode(r, opts...)` | `io.Reader` | HTTP uploads, `os.File` |
| `DecodeBytes(data, opts...)` | encoded `[]byte` | caches, object storage |
| `DecodeDataURI(uri, opts...)` | `data:image/...;base64,...` | HTML embeds |
| `FromImage(src, opts...)` | `image.Image` | stdlib decoders |
| `New(w, h, opts...)` | blank canvas | generated graphics |
| `Animate(init, opts...)` | frame builder | GIF assembly |

Options apply to a **single** call. They copy the default `Config` (`AutoOrientation: true`, `DecodeAnimation: true`, `BlendingColor: "ffffff"`, `Strip: false`) and overlay the functions you pass.

```go
img := goimage.Open("input.png",
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
```

Reuse one config across many files with `WithConfig`:

```go
cfg := goimage.DefaultConfig()
cfg.DecodeAnimation = false // stills only
thumb := goimage.Open("photo.jpg", goimage.WithConfig(cfg)).Cover(200, 200)
```

## Geometry in one line

`Resize(400)` keeps aspect ratio. `Cover` takes `WithAnchor`. Zero or negative sizes set delayed `ErrInvalidDimensions`.

```go
wide := goimage.Open("photo.jpg").Resize(800) // height computed from ratio
box := goimage.Open("photo.jpg").Cover(400, 300, goimage.WithAnchor("top"))
```

## Next

- [Core Design](/getting-started/design) — why the API looks this way
- [Image](/concepts/image) — inspectors, `Clone`, owned buffers
- [Cookbook](/cookbook/) — avatars, watermarks, HTTP

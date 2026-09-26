# go-image

[English](README.md) | [简体中文](README_zh-CN.md)

A fluent image processing library for Go. Package-level functions, functional options, and a delayed error on `Image` keep chains short and concurrent-friendly.

Uses only the Go standard library and official `golang.org/x/image` packages. No CGO.

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, "center").
		Greyscale().
		Sharpen(10)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("out.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Install

```bash
go get github.com/yunkeweb/go-image
```

Requires Go 1.22+.

## Features

- **Open / Decode / New**: path, `io.Reader`, `[]byte`, `image.Image`, or a blank canvas
- **Animate** multi-frame GIFs
- **Geometry**: Resize, Scale, Cover, Contain, Pad, Crop, Trim, ResizeCanvas
- **Effects**: Greyscale, Invert, Brightness, Contrast, Gamma, Colorize, Blur, Sharpen, Pixelate, Rotate, Flip, Flop, Orient
- **Draw**: pixel, rectangle, ellipse, circle, polygon, line, bezier, flood fill
- **Place** watermarks with 9-point alignment and opacity
- **Text** with TTF/OTF files or the built-in bitmap face
- **Encode** JPEG, PNG, GIF (including animation), WebP (lossless VP8L), BMP, TIFF
- **sync.Pool** for NRGBA buffers, with a 16 MiB put cap for large images

## Package-level API

```go
canvas := goimage.New(800, 600)
photo := goimage.Open("input.png")
fromReader := goimage.Decode(r)
fromBytes := goimage.DecodeBytes(raw)
fromStd := goimage.FromImage(stdImg)

img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
```

| Function | Input |
|----------|--------|
| `Open` | filesystem path (`string`) |
| `Decode` | `io.Reader` |
| `DecodeBytes` | encoded `[]byte` |
| `DecodeDataURI` | `data:image/...;base64,...` |
| `FromImage` | `image.Image` |
| `New` | canvas width and height |
| `Animate` | frame builder callback |

Functional options (`WithAutoOrientation`, `WithDecodeAnimation`, `WithBlendingColor`, `WithStrip`) apply to a single call. Reuse them with `NewManager` when many images share one config.

```go
mgr := goimage.NewManager(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
)
canvas := mgr.New(800, 600)
photo := mgr.Open("input.png")
anim := mgr.Animate(func(a *goimage.Animation) {
	a.AddFile("frame1.png", 0.1).AddFile("frame2.png", 0.1).SetLoops(0)
})
```

## Delayed errors

Modifiers return `*Image` so chains stay fluent. The first failure is stored and later calls become no-ops until you inspect it:

```go
img := goimage.Open("missing.jpg").Cover(200, 200, "center")
if err := img.Err(); err != nil {
	log.Fatal(err)
}
```

Encode helpers return `EncodedImage`. Check `enc.Err()` or the error from `Save`.

## Formats

| Format | Decode | Encode |
|--------|--------|--------|
| JPEG | yes (EXIF Orient when auto-orientation is on) | yes (quality 1–100, default 80; `ToJPEG()` / `ToJPEG(95)`) |
| PNG | yes | yes |
| GIF | still + animated | still + animated |
| WebP | yes (`x/image/webp`) | lossless VP8L |
| BMP | yes | yes |
| TIFF | yes | yes |
| AVIF / HEIC / JPEG 2000 | delayed `ErrNotSupported` | delayed `ErrNotSupported` |

A width or height of `0` on resize helpers means “unspecified” (keep aspect ratio). Both `0` yields `ErrInvalidDimensions`.

## Documentation

- [Getting started](https://yunkeweb.github.io/go-image/guide/getting-started.html)
- [API reference](https://yunkeweb.github.io/go-image/api/manager.html)
- [Recipes](https://yunkeweb.github.io/go-image/recipes/)
- Local docs: `npm install && npm run docs:dev`

## Tests

```bash
go test -v ./...
```

## License

MIT. See [LICENSE](LICENSE).

## Acknowledgements

Thanks to [Oliver Vogel](https://intervention.io) and the [Intervention Image](https://github.com/Intervention/image) contributors. Their MIT-licensed work informed the feature set of this library.

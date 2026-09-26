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
		Cover(400, 300).
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
- **Geometry**: `Resize(400)` or `Resize(400, 300)`; Cover / Crop / Pad with `WithAnchor`
- **Effects**: Greyscale, Invert, Brightness, Contrast, Gamma, Colorize, Blur, Sharpen, Pixelate, Rotate, Flip, Flop, Orient
- **Draw**: pixel, rectangle, ellipse, circle, polygon, line, bezier, flood fill
- **Place** watermarks with 9-point alignment and opacity
- **Text** with TTF/OTF files or the built-in bitmap face
- **Encode** JPEG, PNG, GIF (including animation), WebP (lossless VP8L), BMP, TIFF; `WriteTo` for HTTP
- **Clone** deep-copies pixels for concurrent pipelines (`Image` is not concurrent-safe)
- **sync.Pool** for NRGBA buffers, with a 16 MiB put cap for large images
- **Subpackages**: `modifier/` (geometry, effects, drawing, GIF), `encoder/` (codecs), `internal/` (pool and color tables). Callers still import only `github.com/yunkeweb/go-image`

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

Functional options (`WithAutoOrientation`, `WithDecodeAnimation`, `WithBlendingColor`, `WithStrip`) apply to a single call. Shared settings live in `Config` / `DefaultConfig()` and attach with `WithConfig`:

```go
cfg := goimage.DefaultConfig()
cfg.DecodeAnimation = false
photo := goimage.Open("input.png", goimage.WithConfig(cfg))
canvas := goimage.New(800, 600, goimage.WithConfig(cfg))
anim := goimage.Animate(func(a *goimage.Animation) {
	a.AddFile("frame1.png", 0.1).AddFile("frame2.png", 0.1).SetLoops(0)
}, goimage.WithConfig(cfg))
```

## Delayed errors

Modifiers return `*Image` so chains stay fluent. The first failure is stored and later calls become no-ops until you inspect it:

```go
img := goimage.Open("missing.jpg").Cover(200, 200)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
```

Encode helpers return `EncodedImage`. Check `enc.Err()` or the error from `Save` / `WriteTo`. An `Image` is not safe for concurrent mutation; `Clone()` before sharing one source across goroutines.

```go
enc := img.ToJPEG(85)
w.Header().Set("Content-Type", enc.MimeType())
_, _ = enc.WriteTo(w) // http.ResponseWriter, Gin, or any io.Writer
```

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

`Resize(400)` keeps aspect ratio. `Resize(400, 300)` sets both sides. Zero or negative sizes yield delayed `ErrInvalidDimensions`. Cover, Crop, and Pad take `WithAnchor` / `WithBackground` / `WithOffset`.

## Documentation

- [Getting started](https://yunkeweb.github.io/go-image/getting-started/installation.html)
- [Modifying Images](https://yunkeweb.github.io/go-image/modifying/resize.html)
- [Cookbook](https://yunkeweb.github.io/go-image/cookbook/)
- Local docs: `npm install && npm run docs:dev`

## Tests

```bash
go test -v ./...
```

## License

MIT. See [LICENSE](LICENSE).

## Acknowledgements

Thanks to [Oliver Vogel](https://intervention.io) and the [Intervention Image](https://github.com/Intervention/image) contributors. Their MIT-licensed work informed the feature set of this library.

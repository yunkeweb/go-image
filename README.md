# go-image

[English](README.md) | [简体中文](README_zh-CN.md)

Fluent image processing for Go. This module is an idiomatic port of [Intervention Image](https://github.com/Intervention/image) (PHP) to `github.com/yunkeweb/go-image`.

It uses only the Go standard library and the official `golang.org/x/image` packages. PHP exceptions become a delayed error on `Image`, retrieved with `Err()`.

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New().Read("photo.jpg").
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

- **Read** file paths, `[]byte`, `io.Reader`, data URIs, Base64, and `image.Image`
- **Create** canvases and **animate** multi-frame GIFs
- **Geometry**: Resize, Scale, Cover, Contain, Pad, Crop, Trim, ResizeCanvas
- **Effects**: Greyscale, Invert, Brightness, Contrast, Gamma, Colorize, Blur, Sharpen, Pixelate, Rotate, Flip, Flop, Orient
- **Draw**: pixel, rectangle, ellipse, circle, polygon, line, bezier, flood fill
- **Place** watermarks with 9-point alignment and opacity
- **Text** with TTF/OTF files or the built-in bitmap face
- **Encode** JPEG, PNG, GIF (including animation), WebP (lossless VP8L), BMP, TIFF

## Delayed errors

Modifiers return `*Image` so chains stay fluent. The first failure is stored and later calls become no-ops until you inspect it:

```go
img := goimage.Read("missing.jpg").Cover(200, 200, "center")
if err := img.Err(); err != nil {
	log.Fatal(err)
}
```

Encode helpers return `EncodedImage`. Check `enc.Err()` or the error from `Save`.

## Manager

```go
mgr := goimage.New(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)

canvas := mgr.Create(800, 600)
photo := mgr.Read("input.png")
anim := mgr.Animate(func(a *goimage.Animation) {
	a.Add("frame1.png", 0.1).Add("frame2.png", 0.1).SetLoops(0)
})
```

Package-level `Create`, `Read`, and `Animate` use a default manager.

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

A width or height of `0` means “unspecified”, matching PHP `null` on resize helpers.

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

This project is a Go language port of [Intervention Image](https://github.com/Intervention/image) by [Oliver Vogel](https://intervention.io), originally licensed under the MIT License. Thank you to Oliver Vogel and the Intervention Image contributors.

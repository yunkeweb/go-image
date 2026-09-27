# Installation

go-image is a pure Go module. It compiles with the standard toolchain, needs no C compiler, and links only the Go standard library plus official `golang.org/x/image`.

## Requirements

| Item | Value |
|------|--------|
| Go | 1.22 or newer |
| CGO | Off. The library never calls into libvips, ImageMagick, or libwebp C bindings |
| Extra modules | `golang.org/x/image` (JPEG extras, WebP VP8L, BMP, TIFF, fonts, `draw` resampling) |

## Install

```bash
go get github.com/yunkeweb/go-image
```

Pin a release when you vendor production code:

```bash
go get github.com/yunkeweb/go-image@v0.3.0
```

## Import

Callers import one path. Public types (`Image`, `Color`, `Config`, `EncodedImage`) and every modifier live in package `goimage`:

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println(img.Width(), img.Height())
}
```

Internal packages (`modifier`, `encoder`, `internal/pool`) are implementation details. Application code stays on the root import.

## Verify the toolchain

```bash
go version          # go1.22 or newer
go env CGO_ENABLED  # 0 is fine; the module does not need CGO
```

## Next

- [Quickstart](/getting-started/quickstart) — open, cover, sharpen, encode
- [Core Design](/getting-started/design) — package-level API, functional options, `sync.Pool`

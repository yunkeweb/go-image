# Error Handling

go-image stores the first failure on the `Image` value so chains stay fluent. That model is safe for goroutines: each `*Image` owns its own error and pixel buffers. Sentinel kinds are inspected with `errors.Is`.

## Overview

Mutating methods return the same `*Image`. If `Err()` is already set, later modifiers are no-ops. Encode helpers return `EncodedImage`; check `enc.Err()`, `enc.Result()`, or the error from `Save` / `WriteTo`. Geometry that receives `0` or a negative size sets `ErrInvalidDimensions` (wrapped with `ErrGeometry`). Unknown anchors set `ErrGeometry`. Invalid color strings set `ErrColor` instead of falling back to white.

## Signature

```go
func (img *Image) Err() error
func (e EncodedImage) Err() error
func (e EncodedImage) Result() ([]byte, error)
```

## Sentinel errors

| Variable | Typical cause |
|----------|----------------|
| `ErrRuntime` | Nil image or unexpected state |
| `ErrDecoder` | Unreadable or empty input |
| `ErrEncoder` | Encode failure or nil writer |
| `ErrGeometry` | Invalid size |
| `ErrInvalidDimensions` | Width or height is `0` or negative |
| `ErrColor` | Unparseable color |
| `ErrInput` | Bad argument (animation index, nil watermark) |
| `ErrLimit` | Decode exceeded `Limits` (`MaxInputBytes`, size, pixels, or frames) |
| `ErrNotSupported` | AVIF / HEIC / JPEG 2000, trim on animation, or an unimplemented encode option |
| `ErrNotWritable` | Filesystem write failure |
| `ErrAnimation` | Empty animation builder or nil frame |
| `ErrFont` | Font file load failure |
| `ErrDriver` | Driver-level failure |

## Example: delayed decoder error on Open

```go
package main

import (
	"errors"
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("missing.jpg").Cover(200, 200)
	if err := img.Err(); err != nil {
		if errors.Is(err, goimage.ErrDecoder) {
			log.Fatal("could not decode:", err)
		}
		log.Fatal(err)
	}
}
```

## Example: ErrInvalidDimensions from a bad Resize

```go
package main

import (
	"errors"
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(64, 64).Fill("#334155").Resize(0) // 0 is never “auto”
	if err := img.Err(); err != nil {
		if errors.Is(err, goimage.ErrInvalidDimensions) {
			log.Fatal("width/height must be >= 1:", err)
		}
		log.Fatal(err)
	}
}
```

## Notes

- Check `Err()` once after the chain. Intermediate checks are optional.
- `EncodedImage.WriteTo` returns the sticky encode error without writing bytes.
- `errors.Is` works because root sentinels are the same values as `internal/errs`.

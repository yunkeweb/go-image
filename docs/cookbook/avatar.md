# Avatar Crop

User avatars are square, filled, and cropped toward the face. `Cover` with `WithAnchor(AnchorTop)` keeps foreheads and eyes; the default `center` is better for landscape product shots.

## Pipeline

1. `Open` with auto-orientation so phone JPEGs stand upright.
2. `Cover(256, 256, WithAnchor(AnchorTop))` — exact 256×256, crop overflow.
3. `Sharpen(8)` after the downsample.
4. Encode lossless WebP for supporting clients; JPEG quality 85 as fallback.

Zero sizes never mean “auto”: `Cover(0, 256)` sets `ErrInvalidDimensions`.

## Example: package-level square avatar

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg").
		Cover(256, 256).
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: functional options — top anchor + JPEG fallback

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg",
		goimage.WithAutoOrientation(true),
		goimage.WithStrip(true),
	).Cover(256, 256, goimage.WithAnchor(goimage.AnchorTop))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("avatar.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Use `CoverDown` when you must not upscale a 64×64 source to 256×256.
- `Clone()` before writing two formats if another goroutine still holds `img`.

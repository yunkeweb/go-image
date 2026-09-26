# Dynamic Watermark

Ecommerce images usually need a logo in the corner and a timestamp or SKU as text. Compose the logo with `Place`, then `Text` for the live string.

## Pipeline

1. `Open` the product photo; `Cover` to a catalog size.
2. `Open` the logo, `Resize` it so it stays ~10% of the width.
3. `Place(logo, "bottom-right", 16, 16, 70)`.
4. `Text` with a `func(*Font)` callback for color, stroke, and alignment.

`Place` requires `*Image`. A failed logo (missing file) sticks on the destination via `ErrInput` / decoder error — check `Err()` once.

## Example: package-level logo + caption

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	logo := goimage.Open("logo.png").Resize(120)
	img := goimage.Open("product.jpg").
		Cover(1200, 800).
		Place(logo, "bottom-right", 16, 16, 70)
	img.Text("ACME", 24, 40, func(f *goimage.Font) {
		f.Size(22).Color("#ffffff").Stroke("#000000", 2)
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("marked.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: timestamp badge built with New

```go
package main

import (
	"log"
	"time"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	stamp := time.Now().Format("2006-01-02 15:04")
	badge := goimage.New(220, 36, goimage.WithBlendingColor("transparent")).
		Fill("#111827cc")
	badge.Text(stamp, 110, 24, func(f *goimage.Font) {
		f.Size(13).Color("#e2e8f0").Align("center")
	})

	img := goimage.Open("product.jpg", goimage.WithAutoOrientation(true)).
		Place(badge, "top-left", 12, 12, 90)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("dated.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- Hex `#111827cc` is 8-digit RGBA (alpha `cc`). JPEG will flatten that badge; PNG/WebP keep it.
- Reuse one decoded `logo` across goroutines only after `Clone()`, or decode per job like the [WebP thumbnail](/cookbook/webp-thumbnails) recipe.

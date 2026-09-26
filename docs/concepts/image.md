# Image

`Image` is the fluent handle for one still or one animation. It owns a slice of NRGBA frames, a delayed error, EXIF (when present), loop count, and the `Config` captured at decode or `New` time. Mutating methods return the same pointer so chains compose.

## Overview

Decode always copies pixels into library-owned `*image.NRGBA` buffers. The original `image.Image` (YCbCr JPEG, paletted GIF, caller RGBA) is never retained. Geometry that changes size rebases every GIF frame to origin `(0, 0)` so disposal stays consistent.

An `Image` is not safe for concurrent mutation. Independent files belong in independent goroutines. One source shared across goroutines must be `Clone()`d first — `Clone` deep-copies every frame's `Pix`.

## Inspectors

```go
func (img *Image) Err() error
func (img *Image) Clone() *Image
func (img *Image) Native() image.Image
func (img *Image) Frames() []Frame
func (img *Image) Width() int
func (img *Image) Height() int
func (img *Image) Size() Size
func (img *Image) Count() int
func (img *Image) IsAnimated() bool
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) PickColor(x, y int) Color
func (img *Image) PickColorFrame(x, y, frame int) Color
func (img *Image) Exif() map[string]any
func (img *Image) Config() Config
func (img *Image) Origin() Origin
func (img *Image) Save(path string, opts ...EncodeOptions) *Image
```

## Parameters

| Method | Returns / args | Default / notes |
|--------|----------------|-----------------|
| `Err` | first sticky error | `nil` until a step fails |
| `Clone` | deep copy of frames | independent `Pix`; safe to mutate in another goroutine |
| `Native` | primary frame as `image.Image` | owned NRGBA |
| `Width` / `Height` | primary frame size | `0` if the image already failed |
| `Count` | frame count | `1` for stills |
| `Loops` | GIF Netscape loop | `0` means forever |
| `PickColor` | sRGB + alpha at `(x, y)` | out of bounds yields transparent |
| `Save` | encode by path extension | JPEG quality 80 unless `EncodeOptions` passed |

## Example: inspect after a package-level open

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("%dx%d frames=%d animated=%v",
		img.Width(), img.Height(), img.Count(), img.IsAnimated())
	log.Println("origin", img.Origin().MediaType, img.Origin().FilePath)
	log.Println("pixel 0,0", img.PickColor(0, 0))
}
```

## Example: Clone for two pipelines

```go
package main

import (
	"log"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	base := goimage.Open("photo.jpg")
	if err := base.Err(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(400, 300).ToJPEG(80).Save("thumb.jpg")
	}()
	go func() {
		defer wg.Done()
		_ = base.Clone().Greyscale().ToPNG().Save("gray.png")
	}()
	wg.Wait()
}
```

## Notes

- After the first `Err()`, later modifiers return immediately.
- `Native()` is a view of the primary frame. Encoding still goes through `ToJPEG` / `Encode`.
- `Save` writes by file extension (`*.jpg` → JPEG). Prefer `ToJPEG(85).Save(...)` when you want an explicit quality.

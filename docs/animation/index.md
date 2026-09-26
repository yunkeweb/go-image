# Animation

go-image treats a GIF as a slice of owned NRGBA frames plus a loop count. `Animate` builds that slice from stills. `Open` keeps every frame when `WithDecodeAnimation(true)` (the default). Encode with `ToGIF()`.

Lossless WebP encode is still (VP8L). Animated WebP decode is not in this release; multi-frame work is GIF-first.

## Overview

Each frame has pixels, a delay in seconds, and (on encode) a disposal method. Geometry modifiers run on **every** frame and rebase GIF layout to origin `(0, 0)` so disposal stays consistent. Opaque full-canvas frames keep `DisposalNone`; frames with alpha or incomplete coverage use `DisposalBackground` to avoid flashes.

## Signatures

```go
func Animate(init func(*Animation), opts ...Option) *Image

func (a *Animation) Add(src *Image, delaySeconds float64) *Animation
func (a *Animation) AddFile(path string, delaySeconds float64) *Animation
func (a *Animation) SetLoops(n int) *Animation

func (img *Image) IsAnimated() bool
func (img *Image) Count() int
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) RemoveAnimation(position any) *Image
func (img *Image) SliceAnimation(offset int, length int) *Image
```

## Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `init` | `func(*Animation)` | required | Adds frames inside `Animate` |
| `delaySeconds` | `float64` | required | Frame delay in **seconds** (GIF stores centiseconds) |
| `SetLoops(n)` | `int` | `0` | `0` means loop forever |
| `WithDecodeAnimation` | `bool` | `true` | `false` keeps only the first GIF frame |
| `RemoveAnimation` | `int` or `"50%"` | required | Keep one frame by index or percent |
| `SliceAnimation` | `offset, length` | required | Keep a half-open range of frames |

Empty builders set `ErrAnimation`.

## Example: package-level two-frame GIF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	anim := goimage.Animate(func(a *goimage.Animation) {
		a.Add(goimage.New(8, 8).Fill("#ff0000"), 0.2)
		a.Add(goimage.New(8, 8).Fill("#0000ff"), 0.2)
		a.SetLoops(0)
	})
	if err := anim.Err(); err != nil {
		log.Fatal(err)
	}
	if err := anim.ToGIF().Save("blink.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: Open an animated GIF, slice, encode

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("in.gif", goimage.WithDecodeAnimation(true)).
		SliceAnimation(0, 10).
		Cover(240, 240)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("frames", img.Count(), "loops", img.Loops())
	if err := img.ToGIF().Save("clip.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## Next

- [Frames & Disposal](/animation/frames) — how encode picks disposal, `RemoveAnimation`, `Clone`

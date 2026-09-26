# Frames & Disposal

This page covers per-frame control: picking one still, slicing a range, and how GIF disposal is chosen on encode so opaque animations do not flash.

## Overview

A decoded GIF compositor expands each frame onto a full-canvas NRGBA. After geometry, every frame's bounds start at `(0, 0)`. Encode then:

1. Converts delay seconds to centiseconds.
2. Sets `DisposalNone` when the frame is opaque and covers the canvas.
3. Sets `DisposalBackground` when the frame has alpha or only partial coverage.

That pairing prevents a transparent hole from revealing a previous frame, and prevents opaque loops from clearing to the background between frames.

## RemoveAnimation / SliceAnimation

```go
func (img *Image) RemoveAnimation(position any) *Image
func (img *Image) SliceAnimation(offset int, length int) *Image
```

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `position` | `int` or percent `string` | required | Index, or `"0%"`–`"100%"` of `Count()` |
| `offset` | `int` | required | First frame to keep |
| `length` | `int` | required | How many frames to keep |

`RemoveAnimation` **keeps** one frame and drops the rest (it names the animation that is removed). Out-of-range indices clamp.

## Example: freeze the middle frame

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	still := goimage.Open("in.gif").RemoveAnimation("50%")
	if err := still.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("count", still.Count()) // 1
	if err := still.ToPNG().Save("still.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: Clone, then mutate frames on two pipelines

```go
package main

import (
	"log"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	base := goimage.Open("in.gif", goimage.WithDecodeAnimation(true))
	if err := base.Err(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = base.Clone().SliceAnimation(0, 5).ToGIF().Save("head.gif")
	}()
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(120, 120).Greyscale().ToGIF().Save("gray.gif")
	}()
	wg.Wait()
}
```

## Notes

- `WithDecodeAnimation(false)` is cheaper when you only need a poster frame.
- `Trim` cannot run on animations (`ErrNotSupported`); freeze a frame first.
- `sync.Pool` returns discarded frame buffers up to 16 MiB.

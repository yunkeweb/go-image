# Manager

Entry point matching PHP `ImageManager`.

## Signatures

```go
func New(opts ...Option) *Manager
func Create(width, height int) *Image
func Read(input any) *Image
func Animate(init func(*Animation)) *Image

func (m *Manager) Driver() string
func (m *Manager) Config() Config
func (m *Manager) Create(width, height int) *Image
func (m *Manager) Read(input any) *Image
func (m *Manager) Animate(init func(*Animation)) *Image
```

## Options

```go
func WithAutoOrientation(v bool) Option
func WithDecodeAnimation(v bool) Option
func WithBlendingColor(color any) Option
func WithStrip(v bool) Option
```

| Parameter | Type | Notes |
|-----------|------|-------|
| `opts` | `...Option` | Applied to a copy of the default config |
| `width`, `height` | `int` | Must be `>= 1` for `Create` |
| `input` | `any` | Path, bytes, reader, data URI, Base64, `*Image`, `image.Image` |

`Driver()` always returns `"go"`.

## Example

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	mgr := goimage.New(goimage.WithAutoOrientation(true))
	img := mgr.Create(16, 16).Fill("#ff0000")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("red.png"); err != nil {
		log.Fatal(err)
	}
}
```

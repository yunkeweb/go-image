# Package API

Entry points. Most callers use package-level functions. `NewManager` is optional when many images share one `Config`.

## Signatures

```go
func New(width, height int, opts ...Option) *Image
func Create(width, height int, opts ...Option) *Image
func Open(path string, opts ...Option) *Image
func Decode(r io.Reader, opts ...Option) *Image
func DecodeBytes(data []byte, opts ...Option) *Image
func DecodeDataURI(uri string, opts ...Option) *Image
func FromImage(src image.Image, opts ...Option) *Image
func Animate(init func(*Animation), opts ...Option) *Image

func NewManager(opts ...Option) *Manager

func (m *Manager) New(width, height int) *Image
func (m *Manager) Create(width, height int) *Image
func (m *Manager) Open(path string) *Image
func (m *Manager) Decode(r io.Reader) *Image
func (m *Manager) DecodeBytes(data []byte) *Image
func (m *Manager) DecodeDataURI(uri string) *Image
func (m *Manager) FromImage(src image.Image) *Image
func (m *Manager) Animate(init func(*Animation)) *Image
func (m *Manager) Driver() string
func (m *Manager) Config() Config
```

`Create` is an alias of `New`. `Driver()` returns `"go"`.

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
| `width`, `height` | `int` | Must be `>= 1` for `New` |
| `path` | `string` | Filesystem path for `Open` |
| `r` | `io.Reader` | Encoded image bytes for `Decode` |
| `data` | `[]byte` | Encoded image bytes for `DecodeBytes` |
| `src` | `image.Image` | Copied into an owned NRGBA buffer with `draw.Draw` |

| Option | Default | Meaning |
|--------|---------|---------|
| `WithAutoOrientation` | `true` | Apply JPEG EXIF orientation 2–8 after decode |
| `WithDecodeAnimation` | `true` | Keep all GIF frames; `false` keeps the first frame |
| `WithBlendingColor` | `"ffffff"` | Color used when flattening transparency (JPEG) |
| `WithStrip` | `false` | Drop ICC profile bytes on encode |

## Example

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(16, 16, goimage.WithBlendingColor("#ff0000")).
		Fill("#ff0000")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("red.png"); err != nil {
		log.Fatal(err)
	}

	photo := goimage.Open("input.png", goimage.WithAutoOrientation(true))
	if err := photo.Err(); err != nil {
		log.Fatal(err)
	}
}
```

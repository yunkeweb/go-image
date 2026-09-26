# 包级 API

入口函数。大多数场景直接调用包级函数。多图共享同一 `Config` 时再用 `NewManager`。

## 签名

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

`Create` 是 `New` 的别名。`Driver()` 返回 `"go"`。

## 选项

```go
func WithAutoOrientation(v bool) Option
func WithDecodeAnimation(v bool) Option
func WithBlendingColor(color any) Option
func WithStrip(v bool) Option
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `opts` | `...Option` | 应用到默认配置的副本 |
| `width`, `height` | `int` | `New` 时必须 `>= 1` |
| `path` | `string` | `Open` 的文件系统路径 |
| `r` | `io.Reader` | `Decode` 的编码字节流 |
| `data` | `[]byte` | `DecodeBytes` 的编码字节 |

| 选项 | 默认 | 含义 |
|------|------|------|
| `WithAutoOrientation` | `true` | 解码后应用 JPEG EXIF 方向 2–8 |
| `WithDecodeAnimation` | `true` | 保留全部 GIF 帧；`false` 只保留第一帧 |
| `WithBlendingColor` | `"ffffff"` | 压平透明通道时使用的底色（JPEG） |
| `WithStrip` | `false` | 编码时丢弃 ICC profile |

## 示例

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

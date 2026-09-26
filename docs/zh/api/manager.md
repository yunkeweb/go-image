# Manager

对应 PHP `ImageManager` 的入口。

## 签名

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
| `width`, `height` | `int` | `Create` 时必须 `>= 1` |
| `input` | `any` | 路径、字节、Reader、Data URI、Base64、`*Image`、`image.Image` |

`Driver()` 始终返回 `"go"`。

## 示例

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

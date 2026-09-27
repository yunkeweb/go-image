# Place 水印 / 图层

`Place` 把另一张图合成到当前画布的九点 [Anchor](/zh/getting-started/design)。偏移和透明度走 Option，避免连续传多个整数容易传错。

## 功能概述

用 `Open` / `Decode` / `New` 打开叠加层，或传入任意 `image.Image`。叠加层为 nil，或叠加层 `*Image` 已经失败时，目标图像写入 `ErrInput`（或叠加层自身的错误）。叠加层绘制到目标的每一帧。

未知锚点和超出 0–100 的透明度会写入 `Image.Err()`，不会静默夹紧或当成 `top-left`。

这是水印原语。日期戳见 [动态水印](/zh/cookbook/watermark)，流式输出见 [HTTP Handler](/zh/cookbook/http)。

## 签名

```go
func (img *Image) Place(src image.Image, position Anchor, options ...PlaceOption) *Image

func WithOffset(x, y int) GeometryOption
func WithOpacity(opacity int) PlaceOption
```

错误处理：调用链结束后检查 `img.Err()`。

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `src` | `image.Image` | 必填 | 叠加层。`*Image` 必须非 nil 且 `Err() == nil` |
| `position` | `Anchor` | 必填 | 优先用 `AnchorBottomRight`、`AnchorCenter` 等常量。未类型化的 `"center"` 字符串字面量仍可编译 |
| `WithOffset` | `int, int` | `0, 0` | 从锚点向内的平移（像素） |
| `WithOpacity` | `int` | `100` | 0–100。超出范围会设置 `Err()` |

## 示例：包级 Logo 水印

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	logo := goimage.Open("logo.png")
	img := goimage.Open("photo.jpg").
		Cover(1200, 800).
		Place(logo, goimage.AnchorBottomRight, goimage.WithOffset(16, 16), goimage.WithOpacity(70))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("marked.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Open 选项 + 程序生成徽章

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	badge := goimage.New(120, 32, goimage.WithBlendingColor("transparent")).
		Fill("#ef4444").
		Text("SALE", 60, 22, func(f *goimage.Font) {
			f.Size(14).Color("#ffffff").Align("center")
		})
	img := goimage.Open("product.jpg").
		Place(badge, goimage.AnchorTopLeft, goimage.WithOffset(12, 12))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("sale.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 需要更小的水印时，先缩放叠加层：`logo.Resize(120)`。
- `opacity` 为 `0` 时叠加层不可见。省略 `WithOpacity` 即为不透明。

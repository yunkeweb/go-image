# Place 水印 / 图层

`Place` 把另一张 `*Image` 合成到当前画布。位置是九点锚。`opacity` 为 0–100。偏移是从锚点向内的像素。

## 功能概述

用 `Open` / `Decode` / `New` 打开叠加层，再作为 `src` 传入。叠加层为 nil 或已经失败时，目标图像写入 `ErrInput`（或叠加层自身的错误）。叠加层用其主缓冲绘制到目标的每一帧。

这是水印原语。日期戳见 [动态水印](/zh/cookbook/watermark)，流式输出见 [HTTP Handler](/zh/cookbook/http)。

## 签名

```go
func (img *Image) Place(src *Image, position string, offsetX, offsetY, opacity int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `src` | `*Image` | 必填 | 叠加层。必须非 nil 且 `Err() == nil` |
| `position` | `string` | 必填 | `center`、`top`、`top-left`、`top-right`、`left`、`right`、`bottom`、`bottom-left`、`bottom-right` |
| `offsetX`, `offsetY` | `int` | 必填 | 从锚点向内的平移（像素） |
| `opacity` | `int` | 必填 | 0–100。`100` 为不透明 |

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
		Place(logo, "bottom-right", 16, 16, 70)
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
		Place(badge, "top-left", 12, 12, 100)
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
- `opacity` 为 `0` 时叠加层不可见；签名不夹紧该值 — 请保持在 0–100。

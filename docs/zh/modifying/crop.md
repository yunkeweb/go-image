# Crop 裁剪

`Crop` 取出 `width×height` 的矩形。默认锚点是 `top-left`（与 Cover 的 `center` 不同）。越界区域用 `WithBackground` 填充。`WithOffset` 在锚点生效后再平移矩形。

## 功能概述

Crop 不做重采样。结果的像素尺寸就是 `width×height`。矩形超出源图时，新像素使用背景色（默认白色）。

需要“填满某宽高比并重采样”时用 [Cover](/zh/modifying/cover)。需要改画布大小但不裁切画面时用 [Resize Canvas](/zh/modifying/canvas)。

## 签名

```go
func (img *Image) Crop(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor string) GeometryOption
func WithBackground(color any) GeometryOption
func WithOffset(x, y int) GeometryOption
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width`, `height` | `int` | 必填 | 裁剪尺寸，均 `>= 1` |
| `WithAnchor` | `string` | `"top-left"` | 矩形的九点原点 |
| `WithBackground` | `any` | `"ffffff"` | 源图外像素的填充色 |
| `WithOffset` | `int, int` | `0, 0` | 叠加在锚点原点上 |

## 示例：包级左上角裁剪

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Crop(400, 300)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("crop.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：居中裁剪并偏移

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Crop(400, 300,
			goimage.WithAnchor("center"),
			goimage.WithOffset(8, 0), // 相对中心向右 8px
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("crop-center.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- `Crop(0, 100)` 写入 `ErrInvalidDimensions`。
- 动画 GIF 逐帧裁剪并重基到 `(0, 0)`。
- 方形头像流水线见 [头像裁剪](/zh/cookbook/avatar)。

# Cover / Fit 覆盖裁剪

`Cover` 填满 `width×height` 的盒子并裁掉溢出。这是缩略图算子：输出尺寸就是你指定的值。`Fit` 是 `Cover` 的别名。`CoverDown` 使用相同裁剪，然后只缩小、不放大。

## 功能概述

算法选出与目标宽高比相同的最大源矩形，按九点锚对齐（默认 `center`），再重采样到 `width×height`。长轴上多出的像素被丢弃。

`CoverDown` 仍然按目标比例裁剪，但重采样步骤不会放大 — 适合 64×64 的源图不能变成模糊的 400×400 的场景。

## 签名

```go
func (img *Image) Cover(width, height int, opts ...GeometryOption) *Image
func (img *Image) CoverDown(width, height int, opts ...GeometryOption) *Image
func (img *Image) Fit(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor Anchor) GeometryOption
func WithOffset(x, y int) GeometryOption
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width`, `height` | `int` | 必填 | 输出尺寸，均 `>= 1` |
| `WithAnchor` | `string` | `"center"` | `center`、`top`、`top-left`、`top-right`、`left`、`right`、`bottom`、`bottom-left`、`bottom-right` |
| `WithOffset` | `int, int` | `0, 0` | 锚点之后的额外平移（像素） |

非法尺寸写入 `ErrInvalidDimensions`。

## 示例：包级居中覆盖

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(400, 300)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Functional Options — 顶部锚点且不放大

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg", goimage.WithAutoOrientation(true)).
		CoverDown(400, 400, goimage.WithAnchor("top"))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 人像面部靠近顶部时，常用 `WithAnchor("top")`。
- 本库中 `Fit` 与 `Cover` 完全相同。
- 需要留边而不裁切时用 [Contain](/zh/modifying/contain)。

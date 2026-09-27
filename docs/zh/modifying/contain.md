# Contain / Pad 包含与留边

`Contain` 把图像等比装进 `width×height`，剩余区域用 `WithBackground` 铺色。`Pad` 是 contain-down：永不放大，再留边到盒子尺寸。

## 功能概述

画面保持宽高比。空白带用背景色填充（默认白色）。`WithAnchor` 决定画面贴在盒子的哪一边（默认 `center`）。

Contain 与 [Cover](/zh/modifying/cover) 相反：Cover 丢掉溢出，Contain 增加画布。

## 签名

```go
func (img *Image) Contain(width, height int, opts ...GeometryOption) *Image
func (img *Image) Pad(width, height int, opts ...GeometryOption) *Image

func WithAnchor(anchor Anchor) GeometryOption
func WithBackground(color any) GeometryOption
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width`, `height` | `int` | 必填 | 画布尺寸，均 `>= 1` |
| `WithAnchor` | `string` | `"center"` | 盒子内的九点锚 |
| `WithBackground` | `any` | `"ffffff"` | 留边填充；十六进制、名称或 `Color` |

## 示例：包级白底 Contain

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Contain(800, 800)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("square.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Functional Options 的 Pad

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png").
		Pad(512, 256,
			goimage.WithAnchor(goimage.AnchorLeft),
			goimage.WithBackground("#0f172a"),
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("banner.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- JPEG 会按 `BlendingColor` 压扁透明。留边色为 `transparent` 时请用 PNG 或 WebP。
- 源图已经大于盒子时，`Pad` 先缩小再留边。

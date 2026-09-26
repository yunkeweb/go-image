# Resize Canvas 画布

`ResizeCanvas` 改变画布尺寸，不对画面重采样。多出的空间用 `WithBackground` 填充。`ResizeCanvasRelative` 在当前尺寸上增加（或减少）像素。

## 功能概述

已有像素保持 1:1。更大的画布按 `WithAnchor`（默认 `center`）在四周露出背景。更小的画布按与 Crop 相同的方式裁掉画面。

相对值可以为负：`ResizeCanvasRelative(-20, -20)` 两边各缩小 20 像素。

## 签名

```go
func (img *Image) ResizeCanvas(width, height int, opts ...GeometryOption) *Image
func (img *Image) ResizeCanvasRelative(width, height int, opts ...GeometryOption) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width`, `height` | `int` | 必填 | 绝对画布尺寸（`ResizeCanvas`）或增量（`Relative`） |
| `WithAnchor` | `string` | `"center"` | 旧画面在新画布上的位置 |
| `WithBackground` | `any` | `"ffffff"` | 新像素颜色 |

相对增量应用后，绝对宽高必须 `>= 1`。

## 示例：包级加衬边

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		ResizeCanvas(840, 640, goimage.WithBackground("#111827"))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("matted.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：相对留白 + Functional Options

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("icon.png").
		ResizeCanvasRelative(32, 32,
			goimage.WithAnchor("center"),
			goimage.WithBackground("transparent"),
		)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("padded-icon.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 相对缩小到小于 1×1 时写入 `ErrInvalidDimensions`。
- 还需要把画面缩进盒子时，优先用 [Pad](/zh/modifying/contain)。

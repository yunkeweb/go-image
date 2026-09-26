# Resize 缩放

`Resize` 把图像拉伸到精确尺寸。Catmull-Rom 重采样（`golang.org/x/image/draw`）写出新的自有 NRGBA。省略高度则保持原始宽高比。`ResizeDown` 是同一变换，但永远不会放大。

## 功能概述

设源尺寸 `Sw×Sh`、目标 `Tw×Th`：

- `Resize(Tw, Th)` 把整张源图映射到 `Tw×Th`。宽高比可能改变。
- `Resize(Tw)` 计算 `Th = round(Tw * Sh / Sw)`，保持比例。
- `ResizeDown` 在图像已经落在目标盒子内时保持原样。

宽或高为 0 / 负数会写入延迟错误 `ErrInvalidDimensions`。`0` 从不表示“自动”。

## 签名

```go
func (img *Image) Resize(width int, height ...int) *Image
func (img *Image) ResizeDown(width int, height ...int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width` | `int` | 必填 | 目标宽度，必须 `>= 1` |
| `height` | `...int` | 按宽高比计算 | 传入时必须 `>= 1` |

## 示例：包级只指定宽度

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	// 400×200 的源图 → Resize(100) 得到 100×50
	img := goimage.Open("photo.jpg").Resize(100)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("wide.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：明确盒子 + 只缩小

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		ResizeDown(1920, 1080) // 已经更小时为空操作
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("hd.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- `Resize(400, 300)` 会拉伸。要装进盒子用 [Scale](/zh/modifying/scale)，要填满并裁切用 [Cover](/zh/modifying/cover)。
- 动画 GIF 逐帧缩放并重基到 `(0, 0)`。
- `errors.Is(err, goimage.ErrInvalidDimensions)` 见 [错误处理](/zh/concepts/errors)。

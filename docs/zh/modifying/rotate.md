# Rotate 旋转

`Rotate` 把图像**逆时针**旋转 `angle` 度。新露出的角落用 `background` 填充。画布扩大到旋转后画面的轴对齐包围盒。

## 功能概述

正角度为逆时针（`90` 让图像立在左侧）。插值在 NRGBA 中完成。编码 PNG / WebP 时背景可传 `"transparent"`；JPEG 会按 `BlendingColor` 压扁。

需要按 EXIF 扶正时，用 [Orient](/zh/modifying/flip)，不要手写 `Rotate(90, ...)`。

## 签名

```go
func (img *Image) Rotate(angle float64, background any) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `angle` | `float64` | 必填 | 逆时针角度。`0` 为空操作 |
| `background` | `any` | 必填 | 新角落填充色；十六进制、名称或 `Color` |

## 示例：包级旋转 90°

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Rotate(90, "#000000")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("rotated.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：任意角度 + 透明角落

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png", goimage.WithBlendingColor("ffffff")).
		Rotate(15, "transparent")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("tilt.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 对 GIF 每一帧旋转后会把布局重基到 `(0, 0)`。
- 无法解析的 `background` 会走 `ParseColor` 回退；请传入合法十六进制。

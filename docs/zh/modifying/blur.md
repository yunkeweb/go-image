# Blur 模糊

`Blur` 以半径 `amount` 做方框模糊。`amount <= 0` 为空操作。半径越大越耗 CPU；只需缩略图时先 [Cover](/zh/modifying/cover) 再模糊。

## 功能概述

滤镜在 NRGBA 邻域正方形内求平均。这是方框模糊（不是高斯）。多次调用近似更宽的核：`Blur(2).Blur(2)` 比单次 `Blur(2)` 更重。

## 签名

```go
func (img *Image) Blur(amount int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `amount` | `int` | 必填 | 方框模糊半径（像素）。`<= 0` 为空操作 |

## 示例：包级轻度模糊

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Blur(4)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("soft.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：缩小后再做背景模糊

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(800, 450, goimage.WithAnchor("center")).
		Blur(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("backdrop.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- GIF 每一帧独立模糊；编码时重新计算 Disposal。
- 隐私遮挡请优先 [Pixelate](/zh/modifying/pixelate) 或实心 [Fill](/zh/modifying/fill) 矩形。

# Greyscale 灰度

`Greyscale` 把每个像素转为亮度，并写入 R、G、B。Alpha 不变。动画的每一帧都会处理。

## 功能概述

亮度使用接近 Rec. 709 的 sRGB 加权和。结果仍是 NRGBA，后续着色（Colorize）或 JPEG 编码可以继续进行。

## 签名

```go
func (img *Image) Greyscale() *Image
```

## 参数说明

该方法没有参数。

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| — | — | — | 在自有 NRGBA 帧上就地运算 |

## 示例：包级灰度 JPEG

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Greyscale()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("gray.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Cover 之后再转灰度

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Cover(400, 300, goimage.WithAnchor("center")).
		Greyscale().
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("gray-thumb.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 已经是灰色的图像保持灰色；仍会跑完整像素遍历。
- 与 [Contrast](/zh/modifying/contrast) 组合可得到更有力度的黑白缩略图。

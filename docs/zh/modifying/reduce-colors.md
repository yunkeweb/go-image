# Reduce Colors 减色

`ReduceColors` 用中位切割调色板把图像量化到最多 `limit` 种颜色。`limit < 1` 视为 `1`。`background` 为 API 对称而保留；量化作用在当前 NRGBA 像素上。

## 功能概述

中位切割不断拆分颜色空间直到剩下 `limit` 个盒子，然后每个像素映射到盒中心。在 GIF 编码前若想要小于编码器默认的调色板，或给 PNG 做海报化，使用本方法。

## 签名

```go
func (img *Image) ReduceColors(limit int, background any) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `limit` | `int` | 必填 | 调色板大小。`< 1` 变为 `1` |
| `background` | `any` | 切割过程不使用 | 保留参数；传 `nil` 或颜色即可 |

## 示例：包级 16 色海报化

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").ReduceColors(16, nil)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("poster.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：256 色后编码 GIF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(320, 240, goimage.WithAnchor("center")).
		ReduceColors(256, "#ffffff")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToGIF().Save("mini.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- GIF 编码本身会量化；额外的 `ReduceColors(256, …)` 是可选的预整形。
- 动画帧独立量化，调色板不一致时可能闪烁。

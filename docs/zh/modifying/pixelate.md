# Pixelate 像素化

`Pixelate` 把每个 `size×size` 块替换为平均色。`size <= 1` 为空操作。

## 功能概述

网格对齐左上角。右侧与底部的剩余条带使用相同块大小并裁到图像边界。Alpha 与颜色一起平均。

## 签名

```go
func (img *Image) Pixelate(size int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `size` | `int` | 必填 | 块大小（像素）。`<= 1` 保持原图 |

## 示例：包级马赛克

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Pixelate(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("mosaic.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Cover 缩略图后再像素化

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, goimage.WithAnchor("center")).
		Pixelate(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("preview.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 小图配大 `size` 只会剩下少数色块。
- 这是视觉效果，不是加密。

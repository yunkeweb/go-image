# Invert 反色

`Invert` 用 255 减去每个 RGB 通道。Alpha 不变。适合暗色预览和负片效果。

## 功能概述

映射为 `R' = 255 - R`（G、B 相同）。运算在存储的 sRGB 字节上线性进行，不是线性光。

## 签名

```go
func (img *Image) Invert() *Image
```

## 参数说明

该方法没有参数。

## 示例：包级反色

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Invert()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("negative.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：反色程序生成的画布

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(128, 128, goimage.WithBlendingColor("#000000")).
		Fill("#f8fafc").
		Invert()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("dark.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- Invert 两次会恢复原始 RGB。
- 透明像素保持透明（`A` 不反转）。

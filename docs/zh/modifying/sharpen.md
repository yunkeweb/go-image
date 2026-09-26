# Sharpen 锐化

`Sharpen` 做反锐化风格的增强。`amount <= 0` 为空操作。网页缩略图在降采样后常用 `8`–`16`。

## 功能概述

滤镜强化边缘附近的局部对比。过大的 amount 会放大 JPEG 伪影。请在几何变换（`Cover`、`Resize`）**之后**锐化，让核看到最终像素网格。

## 签名

```go
func (img *Image) Sharpen(amount int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `amount` | `int` | 必填 | 反锐化强度。`<= 0` 为空操作 |

## 示例：包级缩略图锐化

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300).
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：灰度图上更强锐化

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Resize(1200).
		Greyscale().
		Sharpen(14)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("crisp.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 在 JPEG 编码之前锐化；随后质量参数会压缩多出的边缘能量。
- 动画帧独立锐化。

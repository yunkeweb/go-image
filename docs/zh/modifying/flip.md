# Flip / Flop / Orient

`Flip` 上下镜像。`Flop` 左右镜像。`Orient`（别名 `Orientate`）应用 JPEG EXIF 方向 2–8，然后把方向标记写成 `1`，避免后续 JPEG 编码再次旋转。

## 功能概述

`Open` 在 `WithAutoOrientation(true)`（默认）时已经处理 EXIF 方向。若解码时关闭了自动方向，或图像来自 `FromImage`，再自行调用 `Orient`。

JPEG 编码不写 EXIF APP1。`Orient` 之后，查看器看到的就是像素本身。

## 签名

```go
func (img *Image) Flip() *Image
func (img *Image) Flop() *Image
func (img *Image) Orient() *Image
func (img *Image) Orientate() *Image // Orient 的别名
```

## 参数说明

这些方法没有参数。

| 方法 | 效果 |
|------|------|
| `Flip` | 垂直镜像 |
| `Flop` | 水平镜像 |
| `Orient` | EXIF 2–8 → 像素，随后标记为 Normal |
| `Orientate` | 与 `Orient` 相同 |

## 示例：包级 Flop 自拍

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("selfie.jpg").Flop()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("mirror.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：关闭自动方向后再 Orient

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(false)).
		Orient()
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(80).Save("upright.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- `Flip` 再 `Flop` 等价于 180° 旋转，且不扩大画布。
- 没有 EXIF 时 `Orient` 为空操作。

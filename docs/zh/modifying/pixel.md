# Pixel 像素

`DrawPixel` 在每一帧的 `(x, y)` 写入一个 NRGBA 像素。越界坐标被绘制器忽略。非法颜色写入延迟 `ErrColor`。

## 功能概述

这是最低层的绘制调用。填充区域请用 [Fill](/zh/modifying/fill) 或 [Shapes](/zh/modifying/shapes)。坐标属于当前画布空间，已经应用过的几何变换会反映在坐标系里。

## 签名

```go
func (img *Image) DrawPixel(x, y int, col any) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `x`, `y` | `int` | 必填 | 像素坐标，原点在左上 |
| `col` | `any` | 必填 | 十六进制、HTML 名称或 `Color` |

## 示例：包级打点

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(32, 32).Fill("#0f172a").DrawPixel(16, 16, "#38bdf8")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("dot.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Open 后盖一层网格

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(64, 64)
	for x := 0; x < 64; x += 8 {
		img.DrawPixel(x, 0, "#ef4444")
	}
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("grid.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- `PickColor(x, y)` 读取的就是 `DrawPixel` 写入的缓冲。
- 已失败的 `Image` 上绘制为空操作；链结束后检查 `Err()`。

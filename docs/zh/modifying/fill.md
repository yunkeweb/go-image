# Fill 填充

`Fill` 铺满每一帧的全部像素。`FloodFill` 从种子点出发，替换四连通的同色区域。

## 功能概述

整布填充是给 `New` 画布上色的常规方式。洪水填充沿四连通、颜色与种子相同的邻居行走。两条路径都会写入每一帧动画。

## 签名

```go
func (img *Image) Fill(col any) *Image
func (img *Image) FloodFill(x, y int, col any) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `col` | `any` | 必填 | 填充色 |
| `x`, `y` | `int` | FloodFill 必填 | 种子像素 |

## 示例：包级画布填充

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 180).Fill("#1e293b")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("bg.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：区域洪水填充

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("mask.png").FloodFill(10, 10, "#22c55e")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("flood.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 洪水填充在任何颜色差异处停止，包括抗锯齿边缘。
- 非法 `col` 写入 `ErrColor` 并跳过绘制。
- `Fill(c, x, y)` 不再表示洪水填充，请改用 `FloodFill`。

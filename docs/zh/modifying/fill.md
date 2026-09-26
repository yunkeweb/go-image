# Fill 填充

不带额外坐标的 `Fill` 铺满整张画布。再传两个 int 时，从该种子点做洪水填充，替换连通的同色区域。

## 功能概述

整布填充是给 `New` 画布上色的常规方式。洪水填充沿四连通、颜色与种子相同的邻居行走。两条路径都会写入每一帧动画。

## 签名

```go
func (img *Image) Fill(col any, xy ...int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `col` | `any` | 必填 | 填充色 |
| `xy` | `...int` | 省略 = 整张画布 | `len >= 2` 时从 `(xy[0], xy[1])` 洪水填充 |

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
	img := goimage.Open("mask.png").Fill("#22c55e", 10, 10)
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

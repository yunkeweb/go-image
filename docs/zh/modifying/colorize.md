# Colorize 着色

`Colorize` 按有符号百分比给 RGB 通道染色。`red`、`green`、`blue` 均限制在 `-100…100`。

## 功能概述

正红把图像推向红色；负红抽掉红色。三轴独立，因此 `Colorize(20, -10, 0)` 会轻微偏暖。

## 签名

```go
func (img *Image) Colorize(red, green, blue int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `red` | `int` | 必填 | `-100…100` 红色偏移 |
| `green` | `int` | 必填 | `-100…100` 绿色偏移 |
| `blue` | `int` | 必填 | `-100…100` 蓝色偏移 |

## 示例：包级暖色

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Colorize(18, 4, -8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("warm.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：灰度后再着色

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Greyscale().
		Colorize(30, 10, 0) // 双色调风格的暖灰
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("duo.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 数值会被夹紧；`Colorize(500, 0, 0)` 等同于 `100, 0, 0`。
- Alpha 不变。

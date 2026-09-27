# Brightness 亮度

`Brightness` 给每个 RGB 通道加上有符号偏移。`level` 限制在 `-100…100`。Alpha 不变。

## 功能概述

`level` 是相对 255 的百分比：`+100` 把通道推向白，`-100` 推向黑。超出范围的值在像素遍历前被夹紧。

## 签名

```go
func (img *Image) Brightness(level int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | `int` | 必填 | `-100…100`。夹紧后 `0` 为空操作 |

## 示例：包级提亮

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Brightness(12)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("bright.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Cover 后压暗缩略图

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300, goimage.WithAnchor(goimage.AnchorCenter)).
		Brightness(-15)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("dim.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 与 [Contrast](/zh/modifying/contrast) 搭配，避免堆叠过大的亮度值。
- 动画各帧使用同一 level。

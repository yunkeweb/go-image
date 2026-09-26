# Contrast 对比度

`Contrast` 把每个 RGB 通道拉离或推向中灰。`level` 限制在 `-100…100`。

## 功能概述

正值增大相对 128 的分离。负值把颜色压向灰色。运算是 sRGB 字节上的逐通道仿射映射。

## 签名

```go
func (img *Image) Contrast(level int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | `int` | 必填 | `-100…100`。夹紧后 `0` 为空操作 |

## 示例：包级增强对比

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Contrast(18)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("punch.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：先灰度再对比度

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true)).
		Greyscale().
		Contrast(25)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("bw.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 过大的正对比度会裁切高光和阴影。
- 与 [Brightness](/zh/modifying/brightness) 组合可做曝光风格调整。

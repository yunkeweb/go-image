# Scale 等比缩放

`Scale` 在保持宽高比的前提下把图像装进盒子。结果在某一轴上可能小于盒子。`ScaleDown` 永远不会放大。二者都用 Catmull-Rom 重采样。

## 功能概述

`Resize(w, h)` 会拉伸到两边；`Scale(w, h)` 选择仍能装进盒子的较大缩放比。`Scale(w)` 只指定宽度，对保持比例的静图与 `Resize(w)` 相同。

需要“最大 800×600”且不留边时用 Scale。留边见 [Contain](/zh/modifying/contain)。填满并裁切见 [Cover](/zh/modifying/cover)。

## 签名

```go
func (img *Image) Scale(width int, height ...int) *Image
func (img *Image) ScaleDown(width int, height ...int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width` | `int` | 必填 | 盒子宽度，`>= 1` |
| `height` | `...int` | 按宽高比计算 | 传入时为盒子高度，`>= 1` |

0 或负数写入 `ErrInvalidDimensions`。

## 示例：包级按最大宽度缩放

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Scale(800) // 高度跟随比例
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG().Save("scaled.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：装进盒子且不放大

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("icon.png", goimage.WithAutoOrientation(true)).
		ScaleDown(256, 256) // 小图标保持原尺寸
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("icon-256.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 仅当源图比例等于 `width:height` 时，输出尺寸才等于盒子。
- 动画帧独立缩放，然后重基到 `(0, 0)`。

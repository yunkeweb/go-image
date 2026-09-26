# 动图概述

go-image 把 GIF 看成自有 NRGBA 帧切片加上循环次数。`Animate` 用静图组装该切片。`Open` 在 `WithDecodeAnimation(true)`（默认）时保留全部帧。用 `ToGIF()` 编码。

无损 WebP 编码仍是静图（VP8L）。本版本不含动画 WebP 解码；多帧工作以 GIF 为主。

## 功能概述

每帧包含像素、以秒计的延迟，以及（编码时）Disposal 方法。几何修饰器作用于**每一帧**，并把 GIF 布局重基到原点 `(0, 0)`，保证 Disposal 一致。不透明全画布帧使用 `DisposalNone`；有 alpha 或未铺满画布的帧使用 `DisposalBackground`，避免闪烁。

## 签名

```go
func Animate(init func(*Animation), opts ...Option) *Image

func (a *Animation) Add(src *Image, delaySeconds float64) *Animation
func (a *Animation) AddFile(path string, delaySeconds float64) *Animation
func (a *Animation) SetLoops(n int) *Animation

func (img *Image) IsAnimated() bool
func (img *Image) Count() int
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) RemoveAnimation(position any) *Image
func (img *Image) SliceAnimation(offset int, length int) *Image
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `init` | `func(*Animation)` | 必填 | 在 `Animate` 内添加帧 |
| `delaySeconds` | `float64` | 必填 | 帧延迟，单位**秒**（GIF 存储为百分之一秒） |
| `SetLoops(n)` | `int` | `0` | `0` 表示无限循环 |
| `WithDecodeAnimation` | `bool` | `true` | `false` 只保留 GIF 首帧 |
| `RemoveAnimation` | `int` 或 `"50%"` | 必填 | 按索引或百分比保留一帧 |
| `SliceAnimation` | `offset, length` | 必填 | 保留半开区间内的帧 |

空构建器写入 `ErrAnimation`。

## 示例：包级两帧 GIF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	anim := goimage.Animate(func(a *goimage.Animation) {
		a.Add(goimage.New(8, 8).Fill("#ff0000"), 0.2)
		a.Add(goimage.New(8, 8).Fill("#0000ff"), 0.2)
		a.SetLoops(0)
	})
	if err := anim.Err(); err != nil {
		log.Fatal(err)
	}
	if err := anim.ToGIF().Save("blink.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：打开动图、切片、编码

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("in.gif", goimage.WithDecodeAnimation(true)).
		SliceAnimation(0, 10).
		Cover(240, 240)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("frames", img.Count(), "loops", img.Loops())
	if err := img.ToGIF().Save("clip.gif"); err != nil {
		log.Fatal(err)
	}
}
```

## 下一步

- [帧控制与 Disposal](/zh/animation/frames) — 编码如何选择 Disposal、`RemoveAnimation`、`Clone`

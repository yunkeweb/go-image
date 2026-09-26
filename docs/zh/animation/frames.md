# 帧控制与 Disposal

本页说明逐帧控制：抽出静帧、切片区间，以及 GIF 编码如何选择 Disposal，避免不透明动画闪烁。

## 功能概述

解码时的合成器把每一帧展开到全画布 NRGBA。几何变换之后，每帧边界从 `(0, 0)` 开始。编码时：

1. 把延迟秒数换成百分之一秒。
2. 帧不透明且铺满画布时使用 `DisposalNone`。
3. 帧有 alpha 或只覆盖部分画布时使用 `DisposalBackground`。

这样既避免透明空洞露出上一帧，也避免不透明循环在两帧之间清成背景。

## RemoveAnimation / SliceAnimation

```go
func (img *Image) RemoveAnimation(position any) *Image
func (img *Image) SliceAnimation(offset int, length int) *Image
```

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `position` | `int` 或百分比 `string` | 必填 | 下标，或 `Count()` 的 `"0%"`–`"100%"` |
| `offset` | `int` | 必填 | 保留的第一帧 |
| `length` | `int` | 必填 | 保留多少帧 |

`RemoveAnimation` **保留**一帧并丢掉其余（名字表示去掉动画）。越界下标会夹紧。

## 示例：冻结中间帧

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	still := goimage.Open("in.gif").RemoveAnimation("50%")
	if err := still.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("count", still.Count()) // 1
	if err := still.ToPNG().Save("still.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Clone 后两条流水线改帧

```go
package main

import (
	"log"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	base := goimage.Open("in.gif", goimage.WithDecodeAnimation(true))
	if err := base.Err(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = base.Clone().SliceAnimation(0, 5).ToGIF().Save("head.gif")
	}()
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(120, 120).Greyscale().ToGIF().Save("gray.gif")
	}()
	wg.Wait()
}
```

## 注意细节

- 只要海报帧时，`WithDecodeAnimation(false)` 更省。
- 动图不能 `Trim`（`ErrNotSupported`）；先冻结一帧。
- 丢弃的帧缓冲在 16 MiB 上限内归还 `sync.Pool`。

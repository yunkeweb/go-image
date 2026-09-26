# Image 结构体

`Image` 是一张静图或一段动画的流式句柄。它持有 NRGBA 帧切片、延迟错误、EXIF（若有）、循环次数，以及解码或 `New` 时捕获的 `Config`。修改方法返回同一指针，因此可以链式组合。

## 功能概述

解码始终把像素拷贝进库自有的 `*image.NRGBA` 缓冲。原始 `image.Image`（YCbCr JPEG、调色板 GIF、调用方 RGBA）不会被保留。改变尺寸的几何操作会把每一帧 GIF 重基到原点 `(0, 0)`，保证 Disposal 一致。

`Image` 不是并发安全的可变对象。独立文件放进独立 goroutine。同一来源要跨 goroutine 使用时，必须先 `Clone()` — `Clone` 会深拷贝每一帧的 `Pix`。

## 查询与生命周期

```go
func (img *Image) Err() error
func (img *Image) Clone() *Image
func (img *Image) Native() image.Image
func (img *Image) Frames() []Frame
func (img *Image) Width() int
func (img *Image) Height() int
func (img *Image) Size() Size
func (img *Image) Count() int
func (img *Image) IsAnimated() bool
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) PickColor(x, y int) Color
func (img *Image) PickColorFrame(x, y, frame int) Color
func (img *Image) Exif() map[string]any
func (img *Image) Config() Config
func (img *Image) Origin() Origin
func (img *Image) Save(path string, opts ...EncodeOptions) *Image
```

## 参数说明

| 方法 | 返回 / 参数 | 默认值 / 说明 |
|------|-------------|---------------|
| `Err` | 第一次粘住的错误 | 成功前为 `nil` |
| `Clone` | 帧的深拷贝 | 独立 `Pix`；可在另一 goroutine 中修改 |
| `Native` | 主帧的 `image.Image` | 自有 NRGBA |
| `Width` / `Height` | 主帧尺寸 | 已失败时为 `0` |
| `Count` | 帧数 | 静图为 `1` |
| `Loops` | GIF Netscape 循环 | `0` 表示无限 |
| `PickColor` | `(x, y)` 的 sRGB + alpha | 越界为透明 |
| `Save` | 按路径扩展名编码 | 未传 `EncodeOptions` 时 JPEG 质量 80 |

## 示例：包级打开后查询

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("%dx%d frames=%d animated=%v",
		img.Width(), img.Height(), img.Count(), img.IsAnimated())
	log.Println("origin", img.Origin().MediaType, img.Origin().FilePath)
	log.Println("pixel 0,0", img.PickColor(0, 0))
}
```

## 示例：Clone 后走两条流水线

```go
package main

import (
	"log"
	"sync"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	base := goimage.Open("photo.jpg")
	if err := base.Err(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = base.Clone().Cover(400, 300).ToJPEG(80).Save("thumb.jpg")
	}()
	go func() {
		defer wg.Done()
		_ = base.Clone().Greyscale().ToPNG().Save("gray.png")
	}()
	wg.Wait()
}
```

## 注意细节

- 第一次 `Err()` 之后，后续修饰立即返回。
- `Native()` 是主帧视图。编码仍走 `ToJPEG` / `Encode`。
- `Save` 按扩展名选择编码器（`*.jpg` → JPEG）。需要明确质量时用 `ToJPEG(85).Save(...)`。

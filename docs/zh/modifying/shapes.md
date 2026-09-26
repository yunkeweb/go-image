# Shapes 图形

矩形、椭圆、圆、多边形、直线和贝塞尔共用 `func(*Drawable)` 回调。在 `Drawable` 上设置尺寸、填充、描边和顶点，库再栅格化到每一帧。

## 功能概述

矩形的 `(x, y)` 是**左上角**。椭圆和圆的 `(x, y)` 是**圆心**。多边形与贝塞尔用 `AddPoint` 收集顶点。直线用 `Line(x1,y1,x2,y2)`。填充来自 `SetBackground`，描边来自 `SetBorder`。

## 签名

```go
func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawPolygon(init func(*Drawable)) *Image
func (img *Image) DrawLine(init func(*Drawable)) *Image
func (img *Image) DrawBezier(init func(*Drawable)) *Image
```

## Drawable 辅助方法

| 方法 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Size(w, h)` | `int, int` | `0, 0` | 矩形 / 椭圆盒子 |
| `SetRadius(r)` | `int` | `0` | 圆半径；椭圆两轴共用 |
| `SetBackground(c)` | `any` | `nil`（不填充） | 填充色 |
| `SetBorder(size, col)` | `int, any` | `0`、`nil` | 描边宽度与颜色 |
| `Line(x1,y1,x2,y2)` | `int×4` | `0` | 直线端点 |
| `AddPoint(x, y)` | `int, int` | 空 | 多边形 / 贝塞尔顶点 |

省略 `SetBorder` 时，直线与贝塞尔默认为 1 像素黑色描边。

## 示例：包级卡片（矩形 + 圆）

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 180).Fill("#0f172a")
	img.DrawRectangle(20, 20, func(d *goimage.Drawable) {
		d.Size(80, 40).SetBackground("#22c55e").SetBorder(2, "#ffffff")
	})
	img.DrawCircle(240, 90, func(d *goimage.Drawable) {
		d.SetRadius(40).SetBackground("#38bdf8")
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("card.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：多边形、直线与贝塞尔

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(200, 200, goimage.WithBlendingColor("#ffffff")).
		Fill("#111827")
	img.DrawPolygon(func(d *goimage.Drawable) {
		d.AddPoint(100, 20).AddPoint(180, 160).AddPoint(20, 160).
			SetBackground("#f97316")
	})
	img.DrawLine(func(d *goimage.Drawable) {
		d.Line(20, 20, 180, 20).SetBorder(3, "#e2e8f0")
	})
	img.DrawBezier(func(d *goimage.Drawable) {
		d.AddPoint(20, 180).AddPoint(100, 80).AddPoint(180, 180).
			SetBorder(2, "#38bdf8")
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("shapes.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- `init` 可以为 `nil`；此时形状尺寸为 0，会被跳过。
- 椭圆半径来自 `Size`（`width/2`、`height/2`），除非设置了 `SetRadius`。
- 叠加已有图像见 [Place](/zh/modifying/place)。

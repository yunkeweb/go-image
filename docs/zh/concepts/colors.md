# 颜色系统

go-image 中的颜色是每通道 8 位的 sRGB，直通 alpha（`255` 为不透明）。`ParseColor` 接受十六进制、HTML 名称、`rgb()` / `rgba()` 字符串、`image/color.Color` 以及 `goimage.Color`。无法解析时返回 `ErrColor`。

## 功能概述

库以 NRGBA 存储像素。修饰器需要填充色（画布留白、旋转角落、洪水填充、JPEG 压扁透明）时，会把参数交给 `ParseColor`。`WithBackground` 与 `WithBlendingColor` 接受同一组 `any`。

HTML 名称（`red`、`steelblue`、`transparent` 等）走内部表。十六进制可以是 `#rgb`、`#rgba`、`#rrggbb`、`#rrggbbaa`，`#` 可省略。

## ParseColor

```go
func ParseColor(v any) (Color, error)

type Color struct {
	R, G, B, A uint8
}

var (
	ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
	ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
	ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)
```

## 参数说明

| 输入类型 | 示例 | 说明 |
|----------|------|------|
| `string` 十六进制 | `"#38bdf8"`、`"fff"`、`"ff000080"` | 3/4/6/8 位；`#` 可省略 |
| `string` 名称 | `"red"`、`"transparent"` | 不区分大小写的 HTML 名 |
| `string` css | `"rgb(16, 185, 129)"`、`"rgba(0,0,0,0.5)"` | 通道 0–255 或 `%` |
| `goimage.Color` | `goimage.Color{R: 56, G: 189, B: 248, A: 255}` | 原样拷贝 |
| `color.Color` | `color.NRGBA{…}` | 16 位右移到 8 位 |
| `nil` / 未知 | | `ErrColor` |

`WithBlendingColor` 默认 `"ffffff"`。`ColorTransparent` 是 alpha 为 0 的白（GIF/PNG 保留透明；JPEG 按混合色压扁）。

## 示例：包级解析并填充

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	c, err := goimage.ParseColor("#0ea5e9")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("R=%d G=%d B=%d A=%d", c.R, c.G, c.B, c.A)

	img := goimage.New(64, 64).Fill("steelblue")
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("swatch.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：用 Functional Options 设置混合色与画布填充

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("logo.png",
		goimage.WithBlendingColor("#111827"), // JPEG 压扁透明时使用
	).Pad(256, 256,
		goimage.WithBackground("transparent"),
		goimage.WithAnchor(goimage.AnchorCenter),
	)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("padded.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 内部 `mustColor` 解析失败时回退为白。公开的 `ParseColor` 始终返回错误。
- JPEG 没有 alpha。编码时按 `BlendingColor`（默认白）压扁，除非你先调用 `BlendTransparency`。
- `PickColor` 读取自有 NRGBA，与修饰器写入的值一致。

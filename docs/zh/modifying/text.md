# Text 文字

`Text` 在 `(x, y)` 栅格化字符串。传入 `nil` 使用内置 `basicfont`；传入 `func(*Font)` 配置字体；或传入 `NewFont` 得到的 `*Font`。TTF/OTF 走 `opentype`。

## 功能概述

`(x, y)` 结合 `Align`（`left` / `center` / `right`）与 `Valign`（`top` / `middle` / `bottom`）解释。`Wrap` 按像素宽度换行。`Stroke` 描边字形（宽度限制 0–10）。`Angle` 逆时针旋转整段文字。

找不到字体文件时写入延迟 `ErrFont`。

## 签名

```go
func (img *Image) Text(text string, x, y int, fontInit any) *Image
func NewFont(filename ...string) *Font
```

## 参数说明

| 名称 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `text` | `string` | 必填 | UTF-8 字符串 |
| `x`, `y` | `int` | 必填 | 锚点 |
| `fontInit` | `any` | `nil` → 内置点阵 | `nil`、`*Font`、`Font` 或 `func(*Font)` |
| `Font.Size` | `float64` | `12` | em 大小 |
| `Font.Color` | `any` | `"000000"` | 填充色 |
| `Font.Align` | `string` | `"left"` | `left` / `center` / `right` |
| `Font.Valign` | `string` | `"bottom"` | `top` / `middle` / `bottom` |
| `Font.LineHeight` | `float64` | `1.25` | 换行行高倍数 |
| `Font.Wrap` | `int` | `0`（关闭） | 换行宽度（像素） |
| `Font.Stroke` | `any, int` | 白、`0` | 描边颜色与宽度 0–10 |
| `Font.Angle` | `float64` | `0` | 逆时针角度 |
| `Font.Filename` | `string` | 空 | TTF/OTF 路径 |

## 示例：包级内置字体

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New(320, 80).Fill("#0f172a")
	img.Text("go-image", 16, 50, nil)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToPNG().Save("label.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Functional Font 回调 + TTF

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(800, 450)
	img.Text("Summer sale", 400, 400, func(f *goimage.Font) {
		f.Filename("fonts/Inter-Bold.ttf").
			Size(36).
			Color("#ffffff").
			Align("center").
			Valign("bottom").
			Stroke("#000000", 2)
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("sale.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 没有字体文件时使用 `golang.org/x/image/font/basicfont`（固定 7×13）。
- `NewFont("face.ttf")` 等价于 `NewFont().Filename("face.ttf")`。
- 时间戳叠加见 [动态水印](/zh/cookbook/watermark)。

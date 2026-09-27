# 动态水印

电商图通常要在角落放 Logo，再用文字写时间戳或 SKU。Logo 用 `Place` 合成，动态字符串用 `Text`。

## 流水线

1. `Open` 商品图；`Cover` 到目录尺寸。
2. `Open` Logo，`Resize` 到大约宽度的 10%。
3. `Place(logo, AnchorBottomRight, WithOffset(16, 16), WithOpacity(70))`。
4. 用 `func(*Font)` 回调写 `Text`，配置颜色、描边和对齐。

`Place` 接受 `image.Image`（包括 `*Image`）。Logo 失败（文件缺失）会通过 `ErrInput` / 解码错误粘到目标上 — 检查一次 `Err()` 即可。

## 示例：包级 Logo + 标题

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	logo := goimage.Open("logo.png").Resize(120)
	img := goimage.Open("product.jpg").
		Cover(1200, 800).
		Place(logo, goimage.AnchorBottomRight, goimage.WithOffset(16, 16), goimage.WithOpacity(70))
	img.Text("ACME", 24, 40, func(f *goimage.Font) {
		f.Size(22).Color("#ffffff").Stroke("#000000", 2)
	})
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("marked.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：用 New 做时间戳徽章

```go
package main

import (
	"log"
	"time"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	stamp := time.Now().Format("2006-01-02 15:04")
	badge := goimage.New(220, 36, goimage.WithBlendingColor("transparent")).
		Fill("#111827cc")
	badge.Text(stamp, 110, 24, func(f *goimage.Font) {
		f.Size(13).Color("#e2e8f0").Align("center")
	})

	img := goimage.Open("product.jpg", goimage.WithAutoOrientation(true)).
		Place(badge, goimage.AnchorTopLeft, goimage.WithOffset(12, 12), goimage.WithOpacity(90))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("dated.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 八位十六进制 `#111827cc` 带 alpha `cc`。JPEG 会压扁该徽章；PNG/WebP 保留透明。
- 跨 goroutine 复用已解码的 `logo` 必须先 `Clone()`，或像 [WebP 缩略图](/zh/cookbook/webp-thumbnails) 那样每任务解码一次。

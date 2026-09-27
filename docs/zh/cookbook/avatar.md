# 头像裁剪

用户头像是正方形、填满盒子、并偏向面部裁切。`Cover` 配合 `WithAnchor(AnchorTop)` 保住额头和眼睛；横向产品图更适合默认的 `center`。

## 流水线

1. `Open` 打开自动方向，让手机 JPEG 扶正。
2. `Cover(256, 256, WithAnchor(AnchorTop))` — 精确 256×256，裁掉溢出。
3. 降采样后 `Sharpen(8)`。
4. 支持的客户端用无损 WebP；回退 JPEG 质量 85。

尺寸为 0 从不表示“自动”：`Cover(0, 256)` 写入 `ErrInvalidDimensions`。

## 示例：包级方形头像

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg").
		Cover(256, 256).
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Functional Options — 顶部锚点 + JPEG 回退

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("portrait.jpg",
		goimage.WithAutoOrientation(true),
		goimage.WithStrip(true),
	).Cover(256, 256, goimage.WithAnchor(goimage.AnchorTop))
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	if err := img.ToWebP().Save("avatar.webp"); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("avatar.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- 64×64 的源图不能放大到 256×256 时，用 `CoverDown`。
- 若另有 goroutine 仍持有 `img`，写两种格式前先 `Clone()`。

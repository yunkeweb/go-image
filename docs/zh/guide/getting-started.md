# 快速开始

## 安装

```bash
go get github.com/yunkeweb/go-image
```

需要 **Go 1.22+**。额外依赖只有官方 `golang.org/x/image`。`EncodedImage.WriteTo` 可写入任意 `io.Writer`（含 HTTP Handler）。单个 `Image` 实例非并发安全，请先 `Clone()`。

## 第一次编码

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300).
		Sharpen(8)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	enc := img.ToJPEG(85)
	if err := enc.Save("thumb.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 创建画布

```go
img := goimage.New(640, 480).
	Fill("#1e293b").
	DrawCircle(320, 240, func(d *goimage.Drawable) {
		d.SetRadius(80).SetBackground("#38bdf8")
	})
if err := img.Err(); err != nil {
	log.Fatal(err)
}
if err := img.ToPNG().Save("circle.png"); err != nil {
	log.Fatal(err)
}
```

## Functional Options

选项作用于单次包级调用：

```go
img := goimage.Open("input.png",
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
```

多图共享配置时用 `NewManager`：

```go
mgr := goimage.NewManager(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
)
img := mgr.Open("input.png")
```

| 选项 | 默认 | 含义 |
|------|------|------|
| `WithAutoOrientation` | `true` | 解码后应用 JPEG EXIF 方向 2–8 |
| `WithDecodeAnimation` | `true` | 保留全部 GIF 帧；`false` 只保留第一帧 |
| `WithBlendingColor` | `"ffffff"` | 压平透明通道时使用的底色（JPEG） |
| `WithStrip` | `false` | 编码时丢弃 ICC profile |

## 强类型输入

| 函数 | 输入 |
|------|------|
| `Open` | 文件系统路径 |
| `Decode` | `io.Reader` |
| `DecodeBytes` | 已编码的 `[]byte` |
| `DecodeDataURI` | `data:image/...;base64,...` |
| `FromImage` | `image.Image` |
| `New` | 空白画布 |

## 下一步

- [错误模型](/zh/guide/errors)
- [支持的格式](/zh/guide/formats)
- [设计理念](/zh/guide/design)

# 快速开始

## 安装

```bash
go get github.com/yunkeweb/go-image
```

需要 **Go 1.22+**。额外依赖仅为官方 `golang.org/x/image`。

## 第一次编码

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Read("photo.jpg").
		Cover(400, 300, "center").
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
img := goimage.Create(640, 480).
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

## Manager 选项

```go
mgr := goimage.New(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)
img := mgr.Read("input.png")
```

| 选项 | 默认 | 含义 |
|------|------|------|
| `WithAutoOrientation` | `true` | 解码后应用 JPEG EXIF 方向 2–8 |
| `WithDecodeAnimation` | `true` | 保留全部 GIF 帧；`false` 只保留第一帧 |
| `WithBlendingColor` | `"ffffff"` | 压平透明通道时使用的底色（JPEG） |
| `WithStrip` | `false` | 编码时丢弃 ICC profile |

包级 `Create`、`Read`、`Animate` 使用上述默认 Manager。

## 可读输入

`Read` 接受：

- 文件路径（存在的 `string` 路径）
- 原始字节（`[]byte`）
- `io.Reader`
- Data URI（`data:image/png;base64,...`）
- Base64 字符串
- `*Image`（克隆）
- `image.Image`（标准库图像）

## 下一步

- [错误模型](/zh/guide/errors)
- [支持的格式](/zh/guide/formats)
- [PHP → Go 对照](/zh/guide/migration)

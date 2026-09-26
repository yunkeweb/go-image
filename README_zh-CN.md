# go-image

[English](README.md) | [简体中文](README_zh-CN.md)

面向 Go 的流式图像处理库。包级函数、Functional Options 与 `Image` 上的延迟错误，让链式调用短、并发友好。

仅依赖 Go 标准库与官方 `golang.org/x/image`。无 CGO。

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").
		Cover(400, 300).
		Greyscale().
		Sharpen(10)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(85).Save("out.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 安装

```bash
go get github.com/yunkeweb/go-image
```

需要 Go 1.22+。

## 功能

- **Open / Decode / New**：路径、`io.Reader`、`[]byte`、`image.Image` 或空白画布
- **Animate** 多帧 GIF
- **几何**：`Resize(400)` 或 `Resize(400, 300)`；Cover / Crop / Pad 使用 `WithAnchor`
- **效果**：Greyscale、Invert、Brightness、Contrast、Gamma、Colorize、Blur、Sharpen、Pixelate、Rotate、Flip、Flop、Orient
- **绘制**：像素、矩形、椭圆、圆、多边形、直线、贝塞尔、洪水填充
- **Place** 水印：九点对齐与透明度
- **文字**：TTF/OTF 文件或内置点阵字体
- **编码**：JPEG、PNG、GIF（含动画）、WebP（无损 VP8L）、BMP、TIFF；`WriteTo` 可直写 HTTP
- **Clone** 深拷贝像素，供并发流水线使用（`Image` 非并发安全）
- **sync.Pool** 回收 NRGBA 缓冲区，超过 16 MiB 的大图不回池

## 包级 API

```go
canvas := goimage.New(800, 600)
photo := goimage.Open("input.png")
fromReader := goimage.Decode(r)
fromBytes := goimage.DecodeBytes(raw)
fromStd := goimage.FromImage(stdImg)

img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
```

| 函数 | 输入 |
|------|------|
| `Open` | 文件系统路径（`string`） |
| `Decode` | `io.Reader` |
| `DecodeBytes` | 已编码的 `[]byte` |
| `DecodeDataURI` | `data:image/...;base64,...` |
| `FromImage` | `image.Image` |
| `New` | 画布宽高 |
| `Animate` | 帧构建回调 |

Functional Options（`WithAutoOrientation`、`WithDecodeAnimation`、`WithBlendingColor`、`WithStrip`）作用于单次调用。多图共享配置时用 `NewManager`。

```go
mgr := goimage.NewManager(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
)
canvas := mgr.New(800, 600)
photo := mgr.Open("input.png")
anim := mgr.Animate(func(a *goimage.Animation) {
	a.AddFile("frame1.png", 0.1).AddFile("frame2.png", 0.1).SetLoops(0)
})
```

## 延迟错误

修改器返回 `*Image`，便于链式调用。第一次失败会被保存，后续调用成为空操作，直到你检查错误：

```go
img := goimage.Open("missing.jpg").Cover(200, 200)
if err := img.Err(); err != nil {
	log.Fatal(err)
}
```

编码方法返回 `EncodedImage`。请检查 `enc.Err()` 或 `Save` / `WriteTo` 返回的 `error`。单个 `Image` 实例非并发安全；跨 goroutine 共享同一来源前先 `Clone()`。

```go
enc := img.ToJPEG(85)
w.Header().Set("Content-Type", enc.MimeType())
_, _ = enc.WriteTo(w) // http.ResponseWriter、Gin 或任意 io.Writer
```

## 格式

| 格式 | 解码 | 编码 |
|------|------|------|
| JPEG | 支持（开启自动方向时应用 EXIF Orient） | 支持（质量 1–100，默认 80；`ToJPEG()` / `ToJPEG(95)`） |
| PNG | 支持 | 支持 |
| GIF | 静态 + 动画 | 静态 + 动画 |
| WebP | 支持（`x/image/webp`） | 无损 VP8L |
| BMP | 支持 | 支持 |
| TIFF | 支持 | 支持 |
| AVIF / HEIC / JPEG 2000 | 延迟 `ErrNotSupported` | 延迟 `ErrNotSupported` |

`Resize(400)` 按原图比例计算高度。`Resize(400, 300)` 同时指定宽高。零或负数返回延迟错误 `ErrInvalidDimensions`。Cover、Crop、Pad 使用 `WithAnchor` / `WithBackground` / `WithOffset`。

## 文档

- [快速开始](https://yunkeweb.github.io/go-image/zh/guide/getting-started.html)
- [API 参考](https://yunkeweb.github.io/go-image/zh/api/manager.html)
- [示例](https://yunkeweb.github.io/go-image/zh/recipes/)
- 本地文档：`npm install && npm run docs:dev`

## 测试

```bash
go test -v ./...
```

## 许可证

MIT。详见 [LICENSE](LICENSE)。

## 致谢

感谢 [Oliver Vogel](https://intervention.io) 与 [Intervention Image](https://github.com/Intervention/image) 的贡献者。他们的 MIT 作品为本库的功能范围提供了参考。

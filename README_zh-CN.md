# go-image

[English](README.md) | [简体中文](README_zh-CN.md)

面向 Go 的流式图像处理库。本模块将 PHP 的 [Intervention Image](https://github.com/Intervention/image) 以惯用 Go 风格移植为 `github.com/yunkeweb/go-image`。

仅依赖 Go 标准库与官方 `golang.org/x/image`。PHP 异常对应 `Image` 上的延迟错误，通过 `Err()` 读取。

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.New().Read("photo.jpg").
		Cover(400, 300, "center").
		Greyscale().
		Sharpen(10)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}
	if err := img.ToJPEG(goimage.EncodeOptions{Quality: 85}).Save("out.jpg"); err != nil {
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

- **读取** 文件路径、`[]byte`、`io.Reader`、Data URI、Base64 以及 `image.Image`
- **创建** 画布，**制作** 多帧 GIF
- **几何**：Resize、Scale、Cover、Contain、Pad、Crop、Trim、ResizeCanvas
- **效果**：Greyscale、Invert、Brightness、Contrast、Gamma、Colorize、Blur、Sharpen、Pixelate、Rotate、Flip、Flop、Orient
- **绘制**：像素、矩形、椭圆、圆、多边形、直线、贝塞尔、洪水填充
- **Place** 水印：九点对齐与透明度
- **文字**：TTF/OTF 文件或内置点阵字体
- **编码**：JPEG、PNG、GIF（含动画）、WebP（无损 VP8L）、BMP、TIFF

## 延迟错误

修改器返回 `*Image`，便于链式调用。第一次失败会被保存，后续调用成为空操作，直到你检查错误：

```go
img := goimage.Read("missing.jpg").Cover(200, 200, "center")
if err := img.Err(); err != nil {
	log.Fatal(err)
}
```

编码方法返回 `EncodedImage`。请检查 `enc.Err()` 或 `Save` 返回的 `error`。

## Manager

```go
mgr := goimage.New(
	goimage.WithAutoOrientation(true),
	goimage.WithDecodeAnimation(true),
	goimage.WithBlendingColor("ffffff"),
	goimage.WithStrip(false),
)

canvas := mgr.Create(800, 600)
photo := mgr.Read("input.png")
anim := mgr.Animate(func(a *goimage.Animation) {
	a.Add("frame1.png", 0.1).Add("frame2.png", 0.1).SetLoops(0)
})
```

包级 `Create`、`Read`、`Animate` 使用默认 Manager。

## 格式

| 格式 | 解码 | 编码 |
|------|------|------|
| JPEG | 支持（开启自动方向时应用 EXIF Orient） | 支持（质量 1–100，默认 75） |
| PNG | 支持 | 支持 |
| GIF | 静态 + 动画 | 静态 + 动画 |
| WebP | 支持（`x/image/webp`） | 无损 VP8L |
| BMP | 支持 | 支持 |
| TIFF | 支持 | 支持 |
| AVIF / HEIC / JPEG 2000 | 延迟 `ErrNotSupported` | 延迟 `ErrNotSupported` |

宽度或高度为 `0` 表示“未指定”，对应 PHP 缩放参数里的 `null`。

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

本项目是 [Oliver Vogel](https://intervention.io) 所著 [Intervention Image](https://github.com/Intervention/image) 的 Go 语言移植，原项目采用 MIT 许可证。感谢 Oliver Vogel 与 Intervention Image 的贡献者。

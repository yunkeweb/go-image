# 编解码器

go-image 用 Go 标准库与 `golang.org/x/image` 解码 JPEG、PNG、GIF、无损 WebP（VP8L）、BMP、TIFF，编码写入同一集合。AVIF、HEIC、JPEG 2000 返回 `ErrNotSupported`。全程无 CGO。

## 功能概述

`Open` / `Decode` / `DecodeBytes` 先嗅探魔数，再把像素拷贝进自有 NRGBA。`WithDecodeAnimation(true)`（默认）时 GIF 保留全部帧。`WithAutoOrientation(true)`（同样默认）时会应用 JPEG EXIF 方向 2–8。

编码辅助返回 `EncodedImage`：字节、媒体类型、`WriteTo`、`Save`、`ToDataURI`。`ToJPEG` 接受变长质量（`ToJPEG()` → 80，`ToJPEG(95)` → 95）。其他格式走 `EncodeOptions`。

## 解码选项

```go
func Open(path string, opts ...Option) *Image
func Decode(r io.Reader, opts ...Option) *Image
func DecodeBytes(data []byte, opts ...Option) *Image
func FromImage(src image.Image, opts ...Option) *Image
```

| 选项 | 类型 | 默认值 | 含义 |
|------|------|--------|------|
| `WithAutoOrientation` | `bool` | `true` | 解码后应用 JPEG EXIF 方向 2–8 |
| `WithDecodeAnimation` | `bool` | `true` | 保留全部 GIF 帧；`false` 只留首帧 |
| `WithBlendingColor` | `any` | `"ffffff"` | JPEG 编码时的压扁色 |
| `WithStrip` | `bool` | `false` | 编码时丢弃 ICC 配置 |

## 编码 API

```go
func (img *Image) ToJPEG(quality ...int) EncodedImage
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage
func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage

func (e EncodedImage) Err() error
func (e EncodedImage) Result() ([]byte, error)
func (e EncodedImage) Bytes() []byte
func (e EncodedImage) MimeType() string
func (e EncodedImage) WriteTo(w io.Writer) (int64, error)
func (e EncodedImage) Save(path string) error
func (e EncodedImage) ToDataURI() string
```

`Err`、`Result`、`WriteTo` 和 `Save` 报告同一次延迟编码失败。使用 `Bytes` 或 `ToDataURI` 之前先检查其中之一。

| 字段 / 参数 | 类型 | 默认值 | 说明 |
|-------------|------|--------|------|
| `ToJPEG` quality | `...int` | `80` | 限制在 1–100；`<= 0` 使用 80 |
| `EncodeOptions.Quality` | `int` | `80` | JPEG / WebP |
| `EncodeOptions.Progressive` | `bool` | `false` | 接受该字段；标准库 JPEG 写 baseline |
| `EncodeOptions.Indexed` | `bool` | `false` | PNG 调色板 |
| WebP | | | 仅无损 VP8L |

`ToAVIF`、`ToHEIC`、`ToJPEG2000` 返回的 `EncodedImage.Err()` 为 `ErrNotSupported`。

## 示例：包级 JPEG，质量 85

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(800, 600)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	enc := img.ToJPEG(85) // 显式质量
	if err := enc.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println(enc.MimeType(), enc.Size())
	if err := enc.Save("out.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## 示例：Functional Options + WebP / PNG

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithStrip(true)).
		Scale(1200)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	webp := img.ToWebP() // 无损 VP8L
	if err := webp.Save("out.webp"); err != nil {
		log.Fatal(err)
	}

	png := img.ToPNG(goimage.EncodeOptions{Indexed: false})
	if err := png.Save("out.png"); err != nil {
		log.Fatal(err)
	}
}
```

## 注意细节

- JPEG 编码不写 EXIF APP1，因此 `Orient()` 之后方向标签不会二次生效。
- GIF 编码对不透明全画布帧使用 `DisposalNone`；有 alpha 或未铺满画布时使用 `DisposalBackground`。
- `WriteTo` 实现 `io.WriterTo`，是 HTTP 输出路径；见 [HTTP Handler](/zh/cookbook/http)。

# 格式

Go 使用单一 NRGBA 驱动，没有 GD 或 Imagick 后端。

## 编解码对照

| 格式 | 解码 | 编码 | 说明 |
|------|------|------|------|
| JPEG | 支持 | 支持 | 质量 1–100，默认 80（`ToJPEG()` / `ToJPEG(95)`）。Progressive 会被接受并忽略（标准库写 baseline）。自动方向应用 EXIF 2–8。 |
| PNG | 支持 | 支持 | |
| GIF | 静态 + 动画 | 静态 + 动画 | 帧延迟单位为秒。无透明全屏帧保留 `DisposalNone`；含透明或区域裁切帧使用 `DisposalBackground`。 |
| WebP | 支持（`x/image/webp`） | 无损 VP8L | 纯 Go 编码器，无 CGO。 |
| BMP | 支持 | 支持 | `x/image/bmp` |
| TIFF | 支持 | 支持 | `x/image/tiff` |
| AVIF | 否 | 否 | 延迟 `ErrNotSupported` |
| HEIC | 否 | 否 | 延迟 `ErrNotSupported` |
| JPEG 2000 | 否 | 否 | 延迟 `ErrNotSupported` |

## 标识符

`Encode`、`EncodeByMediaType`、`EncodeByExtension`、`EncodeByPath` 接受 `jpg`、`image/jpeg`、`.png`、`tif` 等别名。

```go
enc := img.EncodeByExtension("webp")
if err := enc.Save("out.webp"); err != nil {
    log.Fatal(err)
}
```

## 透明度

JPEG 没有 alpha。编码 JPEG 时会把透明像素合成到 `BlendingColor`（默认白色）上。PNG、GIF、WebP、BMP、TIFF 在存在 alpha 时保留它。

## WebP

解码使用 `golang.org/x/image/webp`。编码写入无损 VP8L，再包装为 RIFF/`WEBP`/`VP8L`。不产出有损 VP8。

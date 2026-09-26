# 编码

对应 PHP `EncodedImage` 的二进制输出。

## 签名

```go
func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByMediaType(mediaType string, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByExtension(ext string, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByPath(path string, opts ...EncodeOptions) EncodedImage
func (img *Image) ToJPEG(args ...any) EncodedImage
func (img *Image) ToJPG(args ...any) EncodedImage
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToBitmap(opts ...EncodeOptions) EncodedImage
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToTIF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToJPEG2000(opts ...EncodeOptions) EncodedImage
func (img *Image) ToJP2(opts ...EncodeOptions) EncodedImage
func (img *Image) ToAVIF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToHEIC(opts ...EncodeOptions) EncodedImage
```

## EncodedImage

```go
func (e EncodedImage) Err() error
func (e EncodedImage) Bytes() []byte
func (e EncodedImage) MimeType() string
func (e EncodedImage) Size() int
func (e EncodedImage) ToDataURI() string
func (e EncodedImage) Save(path string) error
```

## EncodeOptions

```go
type EncodeOptions struct {
    Quality     int  // JPEG 0–100；默认 75
    Progressive bool // 接受该字段；标准库 JPEG 为 baseline
    Indexed     bool
    Interlaced  bool
    Bitdepth    int
}
```

`ToJPEG` / `ToJPG` 接受整数质量（`ToJPEG(85)`）或 `EncodeOptions`。JPEG 编码不写入 EXIF APP1，因此 `Orient()` / `Orientate()` 之后不会二次旋转。

`ToAVIF`、`ToHEIC`、`ToJPEG2000` 返回的 `EncodedImage` 其 `Err()` 为 `ErrNotSupported`。

## 示例

```go
img := goimage.Create(32, 32).Fill("blue")
enc := img.ToJPEG(85)
if err := enc.Err(); err != nil {
    log.Fatal(err)
}
if err := enc.Save("blue.jpg"); err != nil {
    log.Fatal(err)
}

enc = img.ToPNG()
if err := enc.Err(); err != nil {
    log.Fatal(err)
}
log.Println(enc.MimeType(), enc.Size(), enc.ToDataURI()[:32])
if err := enc.Save("blue.png"); err != nil {
    log.Fatal(err)
}
```

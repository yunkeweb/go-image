# Encoding

Encoded binary output plus a media type.

## Signatures

```go
func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByMediaType(mediaType string, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByExtension(ext string, opts ...EncodeOptions) EncodedImage
func (img *Image) EncodeByPath(path string, opts ...EncodeOptions) EncodedImage
func (img *Image) ToJPEG(quality ...int) EncodedImage
func (img *Image) ToJPG(quality ...int) EncodedImage
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
    Quality     int  // JPEG 0–100; default 80
    Progressive bool // accepted; stdlib JPEG is baseline
    Indexed     bool
    Interlaced  bool
    Bitdepth    int
}
```

`ToJPEG` / `ToJPG` take a variadic quality: `ToJPEG()` uses 80, `ToJPEG(95)` sets quality to 95. Other encode options go through `Encode(FormatJPEG, EncodeOptions{...})`. JPEG encoding writes no EXIF APP1 segment, so orientation tags cannot double-apply after `Orient()` / `Orientate()`.

GIF encoding keeps `DisposalNone` for opaque full-canvas frames and uses `DisposalBackground` only when a frame has alpha or does not cover the canvas, so opaque animations do not flash between frames.

`ToAVIF`, `ToHEIC`, and `ToJPEG2000` return an `EncodedImage` whose `Err()` is `ErrNotSupported`.

## Example

```go
img := goimage.New(32, 32).Fill("blue")
enc := img.ToJPEG()       // quality 80
enc = img.ToJPEG(95)      // quality 95
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

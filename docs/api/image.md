# Image

Fluent image object. Methods mutate the receiver and return it.

## Inspectors

```go
func (img *Image) Err() error
func (img *Image) Clone() *Image
func (img *Image) Native() image.Image
func (img *Image) Frames() []Frame
func (img *Image) Origin() Origin
func (img *Image) SetOrigin(o Origin) *Image
func (img *Image) Count() int
func (img *Image) IsAnimated() bool
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) Width() int
func (img *Image) Height() int
func (img *Image) Size() Size
func (img *Image) Colorspace() ColorspaceName
func (img *Image) SetColorspace(name string) *Image
func (img *Image) Resolution() (x, y float64)
func (img *Image) SetResolution(x, y float64) *Image
func (img *Image) PickColor(x, y int) Color
func (img *Image) PickColorFrame(x, y, frame int) Color
func (img *Image) PickColors(x, y int) []Color
func (img *Image) Exif() map[string]any
func (img *Image) ExifQuery(key string) any
func (img *Image) SetExif(data map[string]any) *Image
func (img *Image) BlendingColor() Color
func (img *Image) SetBlendingColor(v any) *Image
func (img *Image) Profile() []byte
func (img *Image) SetProfile(data []byte) *Image
func (img *Image) RemoveProfile() *Image
func (img *Image) Config() Config
func (img *Image) Save(path string, opts ...EncodeOptions) *Image
```

## Parameters

| Method | Notes |
|--------|-------|
| `Clone` | Deep-copies pixel buffers, frames, EXIF, and ICC profile. An `Image` is not safe for concurrent mutation; clone before sharing across goroutines |
| `PickColor` | Out-of-bounds returns a zero `Color` |
| `SetColorspace` | `"cmyk"` or anything else → RGB tag. Pixel data stays NRGBA |
| `Resolution` | Default 72×72 |
| `Save` | Format taken from the path extension; empty path uses `Origin.FilePath` |

## Frame

```go
type Frame struct {
    Img        *image.NRGBA
    Delay      float64 // seconds
    Dispose    int
    OffsetLeft int
    OffsetTop  int
}
```

## Origin

```go
type Origin struct {
    MediaType string
    FilePath  string
}
func (o Origin) MimeType() string
func (o Origin) FileExtension() string
```

## Example

```go
img := goimage.New(8, 8).Fill("#00ff00")
c := img.PickColor(0, 0)
log.Printf("%dx%d %#02x%02x%02x", img.Width(), img.Height(), c.R, c.G, c.B)
if err := img.Clone().ToPNG().Save("copy.png"); err != nil {
    log.Fatal(err)
}
```

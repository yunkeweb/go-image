# Image

流式图像对象。方法修改接收者并返回自身。

## 查询与元数据

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

## 参数说明

| 方法 | 说明 |
|------|------|
| `Clone` | 深拷贝像素缓冲、动画帧、EXIF 与 ICC profile。`Image` 实例非并发安全；跨 goroutine 处理同一来源时先 `Clone()` |
| `PickColor` | 越界返回零值 `Color` |
| `SetColorspace` | `"cmyk"` 或其他值标记为 RGB。像素数据仍是 NRGBA |
| `Resolution` | 默认 72×72 |
| `Save` | 格式由路径扩展名决定；空路径使用 `Origin.FilePath` |

## Frame

```go
type Frame struct {
    Img        *image.NRGBA
    Delay      float64 // 秒
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

## 示例

```go
img := goimage.New(8, 8).Fill("#00ff00")
c := img.PickColor(0, 0)
log.Printf("%dx%d %#02x%02x%02x", img.Width(), img.Height(), c.R, c.G, c.B)
if err := img.Clone().ToPNG().Save("copy.png"); err != nil {
    log.Fatal(err)
}
```

# 几何变换

尺寸变换：缩放、铺满、包含、裁剪、留边、去边。

## 签名

```go
func (img *Image) Resize(width int, height ...int) *Image
func (img *Image) ResizeDown(width int, height ...int) *Image
func (img *Image) Scale(width int, height ...int) *Image
func (img *Image) ScaleDown(width int, height ...int) *Image
func (img *Image) Cover(width, height int, opts ...GeometryOption) *Image
func (img *Image) CoverDown(width, height int, opts ...GeometryOption) *Image
func (img *Image) Fit(width, height int, opts ...GeometryOption) *Image
func (img *Image) Contain(width, height int, opts ...GeometryOption) *Image
func (img *Image) Pad(width, height int, opts ...GeometryOption) *Image
func (img *Image) Crop(width, height int, opts ...GeometryOption) *Image
func (img *Image) ResizeCanvas(width, height int, opts ...GeometryOption) *Image
func (img *Image) ResizeCanvasRelative(width, height int, opts ...GeometryOption) *Image
func (img *Image) Trim(tolerance int) *Image

func WithAnchor(anchor string) GeometryOption
func WithBackground(color any) GeometryOption
func WithOffset(x, y int) GeometryOption
```

## 参数

| 名称 | 说明 |
|------|------|
| `width` | 必填，必须 `>= 1` |
| `height` | Resize/Scale 可省略以保持宽高比。传入时必须 `>= 1` |
| `opts` | `WithAnchor`、`WithBackground`、`WithOffset` |
| `tolerance` | `Trim` 相对角点颜色的距离 0–100 |

零或负数返回延迟错误 `ErrInvalidDimensions`。`0` 不再表示“自动”。

锚点默认：Cover / Contain / Pad / ResizeCanvas 为 `center`；Crop 为 `top-left`。背景默认白色。重采样使用 Catmull-Rom（`golang.org/x/image/draw`）。

## 行为

- **Resize** 拉伸到目标。`Resize(400)` 按原图比例计算高度。
- **ResizeDown** 从不放大。
- **Scale** 在框内保持宽高比。`Scale(400)` 只指定宽度。
- **ScaleDown** 是从不放大的 Scale。
- **Cover** / **Fit** 裁出最大的目标宽高比区域，再采样到 `width×height`。
- **CoverDown** 对该裁剪使用向下缩放。
- **Contain** 在框内缩放，用 `WithBackground` 补边。
- **Pad** 是从不放大的 Contain。
- **Crop** 提取矩形；越界用 `WithBackground` 填充。
- **ResizeCanvas** 改画布大小，不重采样画面。
- **Trim** 收缩到与角点颜色不同的像素包围盒。

## 示例

```go
img := goimage.Open("photo.jpg").
    Resize(400).
    Cover(400, 300, WithAnchor("center")).
    Pad(420, 320, WithBackground("#000000"))
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG().Save("framed.jpg"); err != nil {
    log.Fatal(err)
}

cropped := goimage.Open("photo.jpg").
    Crop(400, 300, WithAnchor("center"), WithOffset(8, 0))
```

# 几何变换

尺寸变换：缩放、铺满、包含、裁剪、留边、去边。

## 签名

```go
func (img *Image) Resize(width, height int) *Image
func (img *Image) ResizeDown(width, height int) *Image
func (img *Image) Scale(width, height int) *Image
func (img *Image) ScaleDown(width, height int) *Image
func (img *Image) Cover(width, height int, position ...string) *Image
func (img *Image) CoverDown(width, height int, position ...string) *Image
func (img *Image) Contain(width, height int, background any, position ...string) *Image
func (img *Image) Pad(width, height int, background any, position ...string) *Image
func (img *Image) Crop(width, height, offsetX, offsetY int, background any, position ...string) *Image
func (img *Image) ResizeCanvas(width, height int, background any, position ...string) *Image
func (img *Image) ResizeCanvasRelative(width, height int, background any, position ...string) *Image
func (img *Image) Trim(tolerance int) *Image
```

## 参数

| 名称 | 说明 |
|------|------|
| `width`, `height` | `0` 表示未指定（保持宽高比）。两边都为 `0` 时返回延迟错误 `ErrInvalidDimensions` |
| `position` | 九点枢轴；Cover 默认 `center` |
| `background` | 新画布像素颜色（`Contain`、`Pad`、`Crop`、`ResizeCanvas`） |
| `offsetX`, `offsetY` | 枢轴之后的裁剪偏移 |
| `tolerance` | `Trim` 相对角点颜色的距离 0–100 |

重采样使用 Catmull-Rom（`golang.org/x/image/draw`）。

## 行为

- **Resize** 拉伸到目标框（缺省边按比例）。
- **ResizeDown** 从不放大。
- **Scale** 在框内保持宽高比。
- **ScaleDown** 是从不放大的 Scale。
- **Cover** 裁出最大的目标宽高比区域，再采样到 `width×height`。
- **CoverDown** 对该裁剪使用向下缩放。
- **Contain** 在框内缩放，用 `background` 补边。
- **Pad** 是从不放大的 Contain。
- **Crop** 提取矩形；越界用 `background` 填充。
- **ResizeCanvas** 改画布大小，不重采样画面。
- **Trim** 收缩到与角点颜色不同的像素包围盒。

## 示例

```go
img := goimage.Open("photo.jpg").
    Cover(400, 300, "center").
    Pad(420, 320, "#000000", "center")
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG().Save("framed.jpg"); err != nil {
    log.Fatal(err)
}
```

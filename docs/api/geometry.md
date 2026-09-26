# Geometry

Size transforms: resize, cover, contain, crop, pad, trim.

## Signatures

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

## Parameters

| Name | Notes |
|------|-------|
| `width` | Required, must be `>= 1` |
| `height` | Omit on Resize/Scale to keep aspect ratio. When passed, must be `>= 1` |
| `opts` | `WithAnchor`, `WithBackground`, `WithOffset` |
| `tolerance` | `Trim` color distance 0–100 against the corner pixel |

Zero or negative sizes return delayed `ErrInvalidDimensions`. `0` is never “auto”.

Anchor defaults: `center` for Cover / Contain / Pad / ResizeCanvas; `top-left` for Crop. Background defaults to white. Resampling uses Catmull-Rom (`golang.org/x/image/draw`).

## Behavior

- **Resize** stretches to the target. `Resize(400)` computes height from the original ratio.
- **ResizeDown** never enlarges.
- **Scale** keeps aspect ratio inside the box. `Scale(400)` is width-only.
- **ScaleDown** is Scale that never enlarges.
- **Cover** / **Fit** crop the largest region of the target aspect, then resample to `width×height`.
- **CoverDown** uses resize-down on that crop.
- **Contain** scales inside the box and letterboxes with `WithBackground`.
- **Pad** is Contain that never enlarges (contain-down).
- **Crop** extracts a rectangle; out-of-bounds is filled with `WithBackground`.
- **ResizeCanvas** changes canvas size without resampling the picture.
- **Trim** shrinks to the bounding box of pixels that differ from the corner color.

## Example

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

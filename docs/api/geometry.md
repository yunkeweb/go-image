# Geometry

Size transforms matching PHP `RectangleResizer` and related modifiers.

## Signatures

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

## Parameters

| Name | Notes |
|------|-------|
| `width`, `height` | `0` means unspecified (PHP `null`). Both `0` returns delayed `ErrInvalidDimensions` (no divide-by-zero) |
| `position` | 9-point pivot; default `center` for Cover, `top-left` where PHP uses that default |
| `background` | Color for new canvas pixels (`Contain`, `Pad`, `Crop`, `ResizeCanvas`) |
| `offsetX`, `offsetY` | Crop origin shift after pivot |
| `tolerance` | `Trim` color distance 0–100 against the corner pixel |

Resampling uses Catmull-Rom (`golang.org/x/image/draw`).

## Behavior

- **Resize** stretches to the target box (missing side is proportional).
- **ResizeDown** never enlarges.
- **Scale** keeps aspect ratio inside the box.
- **ScaleDown** is Scale that never enlarges.
- **Cover** crops the largest region of the target aspect, then resamples to `width×height`.
- **CoverDown** uses resize-down on that crop.
- **Contain** scales inside the box and letterboxes with `background`.
- **Pad** is Contain that never enlarges (contain-down).
- **Crop** extracts a rectangle; out-of-bounds is filled with `background`.
- **ResizeCanvas** changes canvas size without resampling the picture.
- **Trim** shrinks to the bounding box of pixels that differ from the corner color.

## Example

```go
img := goimage.Read("photo.jpg").
    Cover(400, 300, "center").
    Pad(420, 320, "#000000", "center")
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG().Save("framed.jpg"); err != nil {
    log.Fatal(err)
}
```

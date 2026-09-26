# Effects

Color, filter, and orientation modifiers.

## Signatures

```go
func (img *Image) Greyscale() *Image
func (img *Image) Invert() *Image
func (img *Image) Brightness(level int) *Image
func (img *Image) Contrast(level int) *Image
func (img *Image) Gamma(gamma float64) *Image
func (img *Image) Colorize(red, green, blue int) *Image
func (img *Image) Pixelate(size int) *Image
func (img *Image) Blur(amount int) *Image
func (img *Image) Sharpen(amount int) *Image
func (img *Image) BlendTransparency(col any) *Image
func (img *Image) ReduceColors(limit int, background any) *Image
func (img *Image) Flip() *Image
func (img *Image) Flop() *Image
func (img *Image) Rotate(angle float64, background any) *Image
func (img *Image) Orient() *Image
```

## Parameters

| Method | Parameter | Range / notes |
|--------|-----------|----------------|
| `Brightness` | `level` | Clamped to -100…100 |
| `Contrast` | `level` | Clamped to -100…100 |
| `Gamma` | `gamma` | Must be `> 0` |
| `Colorize` | `red`, `green`, `blue` | Clamped to -100…100 |
| `Pixelate` | `size` | Block size in pixels; `<= 1` is a no-op |
| `Blur` | `amount` | Box-blur radius; `<= 0` is a no-op |
| `Sharpen` | `amount` | Unsharp amount; `<= 0` is a no-op |
| `ReduceColors` | `limit` | Palette size via median-cut |
| `Rotate` | `angle` | Degrees **counter-clockwise** (PHP GD) |
| `Rotate` | `background` | Fill color for new corners |
| `Flip` | | Vertical (top ↔ bottom) |
| `Flop` | | Horizontal (left ↔ right) |
| `Orient` | | Applies EXIF orientation 2–8, then stamps orientation 1 |

## Example

```go
img := goimage.Read("photo.jpg").
    Orient().
    Greyscale().
    Brightness(10).
    Contrast(5).
    Sharpen(12)
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToJPEG().Save("edit.jpg"); err != nil {
    log.Fatal(err)
}
```

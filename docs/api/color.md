# Color

sRGB color with 8-bit channels. Alpha `255` is opaque, matching PHP RGB colors.

## Signatures

```go
func ParseColor(v any) (Color, error)

func (c Color) NRGBA() color.NRGBA
func (c Color) RGBA() (r, g, b, a uint32)
func (c Color) IsTransparent() bool
func (c Color) IsClear() bool
func (c Color) IsGreyscale() bool
func (c Color) ToHex(prefix string) string
func (c Color) String() string
func (c Color) HSL() (h, s, l float64)
func (c Color) HSV() (h, s, v float64)
func (c Color) CMYK() (cyan, magenta, yellow, key float64)
```

## Accepted input

`ParseColor` (and any `col any` argument) accepts:

- `Color` / `*Color` / `color.Color`
- Hex: `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa` (optional `#`)
- `rgb()`, `rgba()`, `srgb()` function strings
- HTML/CSS names (`red`, `mediumseagreen`, …)
- `"transparent"`

## Constants

```go
var (
    ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
    ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
    ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)
```

## Example

```go
c, err := goimage.ParseColor("rgba(16, 185, 129, 0.5)")
if err != nil {
    log.Fatal(err)
}
h, s, l := c.HSL()
log.Println(c.ToHex("#"), h, s, l)

img := goimage.Create(4, 4).Fill("tomato")
```

# 颜色

每通道 8 位的 sRGB 颜色。Alpha `255` 为不透明。

## 签名

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

## 可接受输入

`ParseColor`（以及任何 `col any` 参数）接受：

- `Color` / `*Color` / `color.Color`
- 十六进制：`#rgb`、`#rgba`、`#rrggbb`、`#rrggbbaa`（`#` 可省略）
- `rgb()`、`rgba()`、`srgb()` 函数字符串
- HTML/CSS 名称（`red`、`mediumseagreen` 等）
- `"transparent"`

## 常量

```go
var (
    ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
    ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
    ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)
```

## 示例

```go
c, err := goimage.ParseColor("rgba(16, 185, 129, 0.5)")
if err != nil {
    log.Fatal(err)
}
h, s, l := c.HSL()
log.Println(c.ToHex("#"), h, s, l)

img := goimage.New(4, 4).Fill("tomato")
```

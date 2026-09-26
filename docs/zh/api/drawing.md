# 绘制

像素、图形、填充、水印与文字。

## 签名

```go
func (img *Image) DrawPixel(x, y int, col any) *Image
func (img *Image) Fill(col any, xy ...int) *Image
func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawPolygon(init func(*Drawable)) *Image
func (img *Image) DrawLine(init func(*Drawable)) *Image
func (img *Image) DrawBezier(init func(*Drawable)) *Image
func (img *Image) Place(src *Image, position string, offsetX, offsetY, opacity int) *Image
func (img *Image) Text(text string, x, y int, fontInit any) *Image
```

## Drawable

```go
func (d *Drawable) Size(w, h int) *Drawable
func (d *Drawable) SetWidth(w int) *Drawable
func (d *Drawable) SetHeight(h int) *Drawable
func (d *Drawable) SetRadius(r int) *Drawable
func (d *Drawable) SetBackground(c any) *Drawable
func (d *Drawable) SetBorder(size int, col any) *Drawable
func (d *Drawable) Line(x1, y1, x2, y2 int) *Drawable
func (d *Drawable) AddPoint(x, y int) *Drawable
```

## Font

```go
func NewFont(filename ...string) *Font
func (f *Font) Filename(path string) *Font
func (f *Font) File(path string) *Font
func (f *Font) Size(v float64) *Font
func (f *Font) Angle(v float64) *Font
func (f *Font) Color(v any) *Font
func (f *Font) Align(v string) *Font
func (f *Font) Valign(v string) *Font
func (f *Font) LineHeight(v float64) *Font
func (f *Font) Wrap(width int) *Font
func (f *Font) Stroke(col any, width int) *Font
```

`Text` 接受 `func(*Font)`、`*Font` 或 `nil`（内置 `basicfont`）。

## 参数

| 方法 | 说明 |
|------|------|
| `Fill` | 无 `xy` 时填充整张画布。两个整数则从该点洪水填充 |
| `DrawRectangle` | `(x, y)` 为左上角；尺寸来自 `Drawable` |
| `DrawCircle` | `(x, y)` 为圆心；使用 `SetRadius` |
| `Place` | `element` 是 `Read` 能接受的任意输入。`opacity` 为 0–100。位置为九点对齐 |
| `Text` | `align` 为 `left`/`center`/`right`；`valign` 为 `top`/`middle`/`bottom` |
| `Stroke` | 宽度限制在 0–10 |

没有字体文件时使用 `golang.org/x/image/font/basicfont`。TTF/OTF 走 `opentype`。

## 示例

```go
img := goimage.New(320, 180).Fill("#0f172a")
img.DrawRectangle(20, 20, func(d *goimage.Drawable) {
    d.Size(80, 40).SetBackground("#22c55e").SetBorder(2, "#ffffff")
})
img.Place(goimage.Open("logo.png"), "bottom-right", 8, 8, 80)
img.Text("go-image", 20, 160, func(f *goimage.Font) {
    f.Size(18).Color("#e2e8f0").Align("left")
})
if err := img.Err(); err != nil {
    log.Fatal(err)
}
if err := img.ToPNG().Save("card.png"); err != nil {
    log.Fatal(err)
}
```

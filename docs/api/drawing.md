# Drawing

Pixels, shapes, fill, watermarks, and text.

## Signatures

```go
func (img *Image) DrawPixel(x, y int, col any) *Image
func (img *Image) Fill(col any, xy ...int) *Image
func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image
func (img *Image) DrawPolygon(init func(*Drawable)) *Image
func (img *Image) DrawLine(init func(*Drawable)) *Image
func (img *Image) DrawBezier(init func(*Drawable)) *Image
func (img *Image) Place(element any, position string, offsetX, offsetY, opacity int) *Image
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

`Text` accepts `func(*Font)`, `*Font`, or `nil` (built-in `basicfont` face).

## Parameters

| Method | Notes |
|--------|-------|
| `Fill` | No `xy` fills the whole canvas. Two ints flood-fill from that point |
| `DrawRectangle` | `(x, y)` is the top-left; size comes from `Drawable` |
| `DrawCircle` | `(x, y)` is the center; `SetRadius` |
| `Place` | `element` is anything `Read` accepts. `opacity` is 0–100. Position is 9-point |
| `Text` | `align` `left`/`center`/`right`; `valign` `top`/`middle`/`bottom` |
| `Stroke` | Width clamped 0–10 |

Without a font file, text uses `golang.org/x/image/font/basicfont`. TTF/OTF files go through `opentype`.

## Example

```go
img := goimage.Create(320, 180).Fill("#0f172a")
img.DrawRectangle(20, 20, func(d *goimage.Drawable) {
    d.Size(80, 40).SetBackground("#22c55e").SetBorder(2, "#ffffff")
})
img.Place("logo.png", "bottom-right", 8, 8, 80)
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

# 效果

颜色、滤镜与方向修改器。

## 签名

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

## 参数

| 方法 | 参数 | 范围 / 说明 |
|------|------|-------------|
| `Brightness` | `level` | 限制在 -100…100 |
| `Contrast` | `level` | 限制在 -100…100 |
| `Gamma` | `gamma` | 必须 `> 0` |
| `Colorize` | `red`, `green`, `blue` | 限制在 -100…100 |
| `Pixelate` | `size` | 像素块大小；`<= 1` 为空操作 |
| `Blur` | `amount` | 盒式模糊半径；`<= 0` 为空操作 |
| `Sharpen` | `amount` | 锐化强度；`<= 0` 为空操作 |
| `ReduceColors` | `limit` | 中位切分调色板大小 |
| `Rotate` | `angle` | 角度，**逆时针**（PHP GD） |
| `Rotate` | `background` | 新角落填充色 |
| `Flip` | | 垂直翻转 |
| `Flop` | | 水平翻转 |
| `Orient` | | 应用 EXIF 方向 2–8，然后标记为 1 |

## 示例

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

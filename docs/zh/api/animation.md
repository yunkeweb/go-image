# 动画

GIF 多帧辅助方法。

## 签名

```go
func Animate(init func(*Animation), opts ...Option) *Image
func (m *Manager) Animate(init func(*Animation)) *Image

func (a *Animation) Add(src *Image, delaySeconds float64) *Animation
func (a *Animation) AddImage(src *Image, delaySeconds float64) *Animation
func (a *Animation) AddFile(path string, delaySeconds float64) *Animation
func (a *Animation) SetLoops(n int) *Animation
func (a *Animation) Loops(n int) *Animation

func (img *Image) IsAnimated() bool
func (img *Image) Count() int
func (img *Image) Loops() int
func (img *Image) SetLoops(n int) *Image
func (img *Image) RemoveAnimation(position any) *Image
func (img *Image) SliceAnimation(offset int, length int) *Image
```

## 参数

| 名称 | 说明 |
|------|------|
| `delaySeconds` | 帧延迟，单位为**秒** |
| `SetLoops(0)` | 无限循环 |
| `RemoveAnimation` | `int` 索引或百分比字符串，如 `"50%"` |
| `SliceAnimation` | `offset` 为起始索引；`length` 为保留帧数 |

`WithDecodeAnimation(false)` 解码 GIF 时只保留第一帧。

`Crop`、`Resize` 等几何修改器会把每一帧重基准到原点 `(0,0)` 并清零偏移。不透明全屏帧保留 `DisposalNone`；含 Alpha 或未铺满画布的帧使用 `DisposalBackground`。

## 示例

```go
anim := goimage.Animate(func(a *goimage.Animation) {
    a.Add(goimage.New(8, 8).Fill("#ff0000"), 0.2)
    a.Add(goimage.New(8, 8).Fill("#0000ff"), 0.2)
    a.SetLoops(0)
})
if err := anim.Err(); err != nil {
    log.Fatal(err)
}
if err := anim.ToGIF().Save("blink.gif"); err != nil {
    log.Fatal(err)
}

still := anim.Clone().RemoveAnimation(0)
_ = still.Count() // 1
```

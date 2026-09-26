# Animation

GIF multi-frame helpers.

## Signatures

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

## Parameters

| Name | Notes |
|------|-------|
| `delaySeconds` | Frame delay in **seconds** |
| `SetLoops(0)` | Loop forever |
| `RemoveAnimation` | `int` index or percent string such as `"50%"` |
| `SliceAnimation` | `offset` is the start index; `length` is how many frames to keep |

`WithDecodeAnimation(false)` keeps only the first GIF frame on decode.

`Crop`, `Resize`, and the other geometry modifiers rebase every frame to origin `(0,0)` and clear offsets. Opaque full-canvas frames keep `DisposalNone`; frames with alpha or incomplete coverage use `DisposalBackground`.

## Example

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

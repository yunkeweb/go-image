# 设计理念

go-image 是原生的 Go 图像库：包级函数、Functional Options、强类型入口，以及像素缓冲池。

## 包级函数优先

```go
img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
canvas := goimage.New(800, 600)
raw := goimage.Decode(reader)
```

只有多图共享同一 `Config` 时才需要 `NewManager`。

## Functional Options

`opts ...Option` 可用于 `Open`、`Decode`、`New`、`Animate` 和 `NewManager`。每次调用复制默认配置再应用选项，并发调用不共享可变全局状态。

## 强类型输入

Go 没有函数重载。入口按类型命名：

| 函数 | 类型 |
|------|------|
| `Open` | `string` 路径 |
| `Decode` | `io.Reader` |
| `DecodeBytes` | `[]byte` |
| `FromImage` | `image.Image` |
| `New` | `int, int` 画布 |

`Place` 接受 `*Image`。水印先打开：`img.Place(goimage.Open("logo.png"), "bottom-right", 16, 16, 70)`。

## 延迟错误与 goroutine

链式调用返回 `*Image`，第一次错误留在该对象上。**单个 `Image` 实例非并发安全**。并行处理互不相关的文件，或在共享同一来源前先 `Clone()`：

```go
var wg sync.WaitGroup
for _, path := range paths {
    path := path
    wg.Add(1)
    go func() {
        defer wg.Done()
        img := goimage.Open(path).Cover(400, 300)
        if err := img.Err(); err != nil {
            return
        }
        _ = img.ToJPEG(80).Save("out/" + filepath.Base(path))
    }()
}
wg.Wait()

base := goimage.Open("photo.jpg")
go func() { _ = base.Clone().Cover(400, 300).ToJPEG().Save("a.jpg") }()
go func() { _ = base.Clone().Greyscale().ToPNG().Save("b.png") }()
```

`FromImage` 与解码路径一律通过 `draw.Draw` 拷贝到库自有的 NRGBA 缓冲。JPEG 的 YCbCr、调色板图、RGBA、NRGBA 都不会因类型断言 Panic，也不会共享调用方的 `Pix`。

抛弃的 NRGBA 缓冲在容量不超过 16 MiB 时回到 `sync.Pool`，避免大图把超大 slice 长期留在池中。

## 几何约定

- `Resize(400)` / `Scale(400)` 省略高度，按原图比例计算。
- `Resize(400, 300)` 同时指定宽高。零或负数返回 `ErrInvalidDimensions`。
- Cover、Contain、Pad、Crop、Fit、ResizeCanvas 使用 `WithAnchor`、`WithBackground`、`WithOffset`。
- 九点枢轴：`center`、`top`、`top-left`、`top-right`、`left`、`right`、`bottom`、`bottom-left`、`bottom-right`。
- `Rotate` 为逆时针。
- `Cover` / `Fit` 铺满目标框并裁切溢出；`Contain` / `Pad` 留边。

绘制与文字接受 `func(*Drawable)` 和 `func(*Font)`（或直接传 `*Font`）。

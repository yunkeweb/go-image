# 核心设计

go-image 是现代化的 Go 图像库。公开面是包级函数、Functional Options、强类型，以及记在 `Image` 上的延迟错误。内部通过 `sync.Pool` 复用 NRGBA 缓冲。零 CGO，像素热路径没有全局锁。

## 包级函数优先

```go
img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
canvas := goimage.New(800, 600)
raw := goimage.Decode(reader)
```

共享的解码设置放在 `Config` 里。复制 `DefaultConfig`（或自建一份 `Config`），再用 `WithConfig` 传给 `Open` / `Decode` / `New` / `Animate`。

## Functional Options

`opts ...Option` 可用于 `Open`、`Decode`、`DecodeBytes`、`FromImage`、`New`、`Animate`。每次调用都会复制 `DefaultConfig` 再应用函数。并发调用方不会共享可变全局状态。

几何辅助方法（`Cover`、`Contain`、`Pad`、`Crop`、`Fit`、`ResizeCanvas`）接受 `...GeometryOption`：

| 选项 | 作用对象 | 默认值 |
|------|----------|--------|
| `WithAnchor(name)` | Cover、Contain、Pad、Crop、画布 | `center`（Crop 为 `top-left`） |
| `WithBackground(color)` | Contain、Pad、Crop、画布 | `"ffffff"` |
| `WithOffset(x, y)` | Crop | `0, 0` |

Resize 一族用变长 height，而不是 option：`Resize(400)` 保持宽高比；`Resize(400, 300)` 同时指定两边。

## 按类型分流入口

Go 没有函数重载。入口按源类型命名：

| 函数 | 类型 |
|------|------|
| `Open` | `string` 路径 |
| `Decode` | `io.Reader` |
| `DecodeBytes` | `[]byte` |
| `FromImage` | `image.Image` |
| `New` | `int, int` 画布 |

`Place` 的参数是 `*Image`。先打开叠加层：

```go
img.Place(goimage.Open("logo.png"), "bottom-right", 16, 16, 70)
```

## 延迟错误与 goroutine

链式方法返回 `*Image`。第一次错误会粘在该值上，后续修饰变成空操作。`Image` **不是**并发安全的可变对象。独立文件可并行处理；同一张图先 `Clone()` 再分享。

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

`Image` 上没有 `sync.Mutex`。隔离来自所有权：每个 goroutine 要么打开自己的文件，要么操作一份 `Clone`。

## 自有缓冲与 `sync.Pool`

`FromImage` 和所有解码器都通过 `draw.Draw` 拷贝到库自有的 NRGBA。JPEG YCbCr、调色板 GIF、RGBA、NRGBA 输入不会在类型断言上 panic，也不会共享调用方的 `Pix` 切片。

丢弃的 NRGBA 在容量不超过 16 MiB 时归还 `sync.Pool`。更大的帧交给垃圾回收，避免一次 5000 万像素解码把巨大切片钉死，拖累后续 64×64 的小图工作。

## 包布局

```go
import goimage "github.com/yunkeweb/go-image"
```

| 路径 | 职责 |
|------|------|
| 根包（`image.go`、`options.go`、`api.go`） | 公开类型与薄封装 |
| `modifier/` | 几何、特效、绘制、GIF 布局 |
| `encoder/` | JPEG、PNG、GIF、WebP、BMP、TIFF |
| `internal/pool`、`internal/color` | 缓冲复用与 HTML 颜色名 |

应用代码只导入根模块。

## 几何约定

- `Resize(400)` / `Scale(400)` 省略高度，保持宽高比。
- `Resize(400, 300)` 同时指定两边。0 或负数得到 `ErrInvalidDimensions`。`0` 从不表示“自动”。
- Cover、Contain、Pad、Crop、Fit、ResizeCanvas 接受 `WithAnchor`、`WithBackground`、`WithOffset`。
- 九点锚：`center`、`top`、`top-left`、`top-right`、`left`、`right`、`bottom`、`bottom-left`、`bottom-right`。
- `Rotate` 为逆时针。
- `Cover` / `Fit` 填满盒子并裁掉溢出。`Contain` / `Pad` 留边。
- 重采样使用 Catmull-Rom（`golang.org/x/image/draw`）。

绘制与文字接受 `func(*Drawable)` 和 `func(*Font)`（或 `*Font` 值）。

# Design

go-image is a native Go image library: package-level functions, functional options, strong types, and pooled pixel buffers.

## Package-level first

```go
img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
canvas := goimage.New(800, 600)
raw := goimage.Decode(reader)
```

`NewManager` is only needed when many images share one `Config`.

## Functional options

`opts ...Option` works on `Open`, `Decode`, `New`, `Animate`, and `NewManager`. Each call copies the default config, then applies options. Concurrent callers do not share mutable global state.

## Typed inputs

Go has no function overloading. Entry points are named by type:

| Function | Type |
|----------|------|
| `Open` | `string` path |
| `Decode` | `io.Reader` |
| `DecodeBytes` | `[]byte` |
| `FromImage` | `image.Image` |
| `New` | `int, int` canvas |

`Place` takes `*Image`. Open the watermark first: `img.Place(goimage.Open("logo.png"), "bottom-right", 16, 16, 70)`.

## Delayed errors and goroutines

Chains return `*Image`. The first error sticks on that value. An `Image` is **not** safe for concurrent mutation. Process independent files in parallel, or `Clone()` before sharing one source:

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

`FromImage` and decode always copy into an owned NRGBA buffer with `draw.Draw`, so JPEG YCbCr, Paletted, RGBA, and NRGBA inputs never panic on a type assertion and never share the caller's `Pix` slice.

Discarded NRGBA buffers go back to `sync.Pool` when their capacity is at most 16 MiB, so large images do not pin huge slices for later small work.

## Package layout

Callers import one module path:

```go
import goimage "github.com/yunkeweb/go-image"
```

The public API lives in the root package (`image.go`, `options.go`, `api.go`). Algorithms sit in `modifier/` (geometry, effects, drawing, GIF layout). Codecs sit in `encoder/` (JPEG, PNG, GIF, WebP, BMP, TIFF). Buffer reuse and color-name tables live under `internal/` and cannot be imported by other modules.

## Geometry conventions

- `Resize(400)` / `Scale(400)` omit height and keep aspect ratio.
- `Resize(400, 300)` sets both sides. Zero or negative sizes yield `ErrInvalidDimensions`.
- Cover, Contain, Pad, Crop, Fit, and ResizeCanvas take `WithAnchor`, `WithBackground`, and `WithOffset`.
- 9-point pivots: `center`, `top`, `top-left`, `top-right`, `left`, `right`, `bottom`, `bottom-left`, `bottom-right`.
- `Rotate` is counter-clockwise.
- `Cover` / `Fit` fill a box and crop overflow. `Contain` / `Pad` letterbox.

Draw and text take `func(*Drawable)` and `func(*Font)` (or a `*Font` value).

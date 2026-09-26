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

Chains return `*Image`. The first error sticks on that value. Process independent files in parallel:

```go
var wg sync.WaitGroup
for _, path := range paths {
    path := path
    wg.Add(1)
    go func() {
        defer wg.Done()
        img := goimage.Open(path).Cover(400, 300, "center")
        if err := img.Err(); err != nil {
            return
        }
        _ = img.ToJPEG(80).Save("out/" + filepath.Base(path))
    }()
}
wg.Wait()
```

Discarded NRGBA buffers go back to `sync.Pool` when their capacity is at most 16 MiB, so large images do not pin huge slices for later small work.

## Geometry conventions

- Width or height `0` means unspecified (keep aspect ratio).
- Both `0` returns delayed `ErrInvalidDimensions`.
- 9-point pivots: `center`, `top`, `top-left`, `top-right`, `left`, `right`, `bottom`, `bottom-left`, `bottom-right`.
- `Rotate` is counter-clockwise.
- `Cover` fills a box and crops overflow. `Contain` / `Pad` letterbox.

Draw and text take `func(*Drawable)` and `func(*Font)` (or a `*Font` value).

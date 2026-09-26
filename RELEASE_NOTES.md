# v0.1.8 - Remove Manager driver abstraction

`Manager` and `Driver()` are gone. Package defaults live in `DefaultConfig`. Reuse a setup with `WithConfig` on `Open`, `Decode`, `New`, and `Animate`.

```go
cfg := goimage.DefaultConfig
cfg.DecodeAnimation = false
img := goimage.Open("photo.jpg", goimage.WithConfig(cfg))
```

---

# v0.1.7 - Idiomatic Go Redesign & Intervention V4 Standard Docs

A fluent, zero-CGO image processing library for Go. Callers import one module path (`github.com/yunkeweb/go-image`) and chain `Open` → geometry / effects / drawing → `ToJPEG` / `ToWebP`. This release closes the idiomatic API, package layout, and documentation work through **v0.1.7**.

## Features

- **Package-level APIs** — `Open`, `Decode`, `DecodeBytes`, `FromImage`, `New`, and `Animate` live at package scope. Functional options (`WithAutoOrientation`, `WithDecodeAnimation`, `WithBlendingColor`, `WithStrip`) apply to a single call; `NewManager` reuses one `Config` across many files.
- **Variadic options for `Resize` and encoders** — `Resize(400)` keeps aspect ratio; `Resize(400, 300)` sets both sides. `ToJPEG()` encodes at quality **80**; `ToJPEG(95)` sets an explicit quality. Cover / Crop / Pad take `WithAnchor`, `WithBackground`, and `WithOffset`.
- **`Image.Clone()` for safe concurrency** — each `Image` owns its NRGBA frames and delayed `Err()`. Independent files run in parallel goroutines; one source shared across pipelines is cloned first so pixel buffers never race.

```go
img := goimage.Open("photo.jpg").
    Cover(400, 300, goimage.WithAnchor("center")).
    Sharpen(8)
enc := img.ToJPEG(85)
_ = enc.Save("thumb.jpg")
```

## Performance

- **`sync.Pool` memory cap** — discarded NRGBA buffers return to the pool only when capacity is **≤ 16 MiB**. Larger frames go to the garbage collector so a 50-megapixel decode cannot pin a huge slice for later thumbnail work.
- **Lock-free high throughput** — there is no `sync.Mutex` on `Image`. Isolation is ownership: each goroutine opens its own file or mutates a `Clone()`. Decode always copies into library-owned NRGBA via `draw.Draw`.
- **Zero CGO** — JPEG, PNG, GIF (including animation), lossless WebP (VP8L), BMP, and TIFF use the Go standard library and official `golang.org/x/image`.

## Refactor

Standard Go subpackage layout. Application code still imports only the root module.

| Path | Role |
|------|------|
| Root (`image.go`, `options.go`, `api.go`, `doc.go`) | Public types and thin wrappers |
| `modifier/` | Geometry, effects, drawing, GIF frame layout |
| `encoder/` | JPEG, PNG, GIF, WebP, BMP, TIFF |
| `internal/pool`, `internal/color` | Buffer reuse and HTML color names |

Geometry that receives `0` or a negative size sets delayed `ErrInvalidDimensions`. GIF encode keeps `DisposalNone` for opaque full-canvas frames and `DisposalBackground` when a frame has alpha or partial coverage.

## Documentation

Fully restructured VitePress docs matching **Intervention Image V4** directory standards, in English and 简体中文:

1. **Getting Started** — install, quickstart, core design
2. **Concepts** — `Image`, colors, encoders/decoders, `Err()`
3. **Modifying Images** — geometry, effects, drawing (overview, parameter tables, annotated Go examples)
4. **Animation** — GIF frame control and disposal
5. **Cookbook** — avatar crop, dynamic watermark, concurrent WebP thumbnails, HTTP `WriteTo`

Docs: [https://yunkeweb.github.io/go-image/](https://yunkeweb.github.io/go-image/)

```bash
go get github.com/yunkeweb/go-image@v0.1.7
```

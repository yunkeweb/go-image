# v0.3.0 - Decode Limits, strict EncodeOptions, and quality gates

`v0.3.0` is the server-oriented stability release. Decode paths honor resource caps, unimplemented encode options fail instead of being ignored, Data URI / JPEG EXIF / GIF inputs are parsed with explicit errors, and CI runs format, vet, shuffle, race, coverage, short benchmarks, fuzz smoke tests, and the VitePress docs build.

```bash
go get github.com/yunkeweb/go-image@v0.3.0
```

## Added

- `Limits` / `WithLimits` on `Open`, `Decode`, `DecodeBytes`, `DecodeDataURI`, `New`, and `FromImage`:
  - `MaxInputBytes` — encoded input size (`io.Reader` is drained with `io.LimitReader`)
  - `MaxWidth` / `MaxHeight`
  - `MaxPixels` — `width × height` for still images, `width × height × frames` for GIF
  - `MaxFrames` — GIF frame count, counted before pixel allocation
  - Integer overflow in the pixel product is treated as over-limit
  - Zero fields remain unlimited (the default)
- Public sentinel `ErrLimit` (`errors.Is`)
- WebP lossless VP8L encode/decode round-trip tests (opaque, alpha, 1×1, non-zero bounds, `x/image/webp`, truncated/corrupt)
- Benchmarks: `BenchmarkDecodeJPEG`, `BenchmarkDecodePNG`, `BenchmarkResize`, `BenchmarkClone`, `BenchmarkEncodeJPEG`, `BenchmarkEncodePNG`, `BenchmarkEncodeWebP`, `BenchmarkGIFDecode`, `BenchmarkGIFEncode`
- Fuzz targets: `FuzzDecodeDataURI`, `FuzzParseJPEGExif`, `FuzzDecodeGIF`, `FuzzDecodeWebP`, `FuzzParseColor` (plus `FuzzDecodeBytes` / `FuzzVP8LRoundTrip`)
- CI matrix: Go 1.22 and `stable`; `gofmt`, `go vet`, `go test -shuffle=on`, `-race`, coverage, short benchmarks, fuzz smoke, docs build
- Dependabot for Go modules, GitHub Actions, and npm

## Changed

- `EncodeOptions` support matrix is enforced. Only JPEG `Quality` is implemented (0 → 80, clamp 1–100). `Progressive`, `Indexed`, `Interlaced`, non-zero `Bitdepth`, and `Quality` on PNG/GIF/WebP/BMP/TIFF return `ErrNotSupported`.
- WebP encode remains lossless VP8L; `Quality` is not applied.
- Data URI parsing follows `data:[mediatype][;base64],data`:
  - `;base64` is a case-insensitive flag parameter, not a substring match
  - standard Base64, no-padding Base64, and newline-wrapped Base64
  - URL percent-decoding of the payload
  - `image/*` media types only; unknown parameters, illegal percent-encoding, illegal Base64, and empty payloads return `ErrDecoder`
- If GIF frame counting fails, decode returns an error. Truncated or corrupt GIF is not treated as a still image.
- JPEG EXIF orientation reads little- and big-endian TIFF, orientations 1–8, and ignores truncated APP1, illegal lengths, missing/illegal Orientation, wrong types, extra APP segments, and later APP1 after a non-Exif APP1.

## Compatibility

- Existing callers that never set `Limits` are unchanged (all zeros = unlimited).
- Callers that passed unimplemented `EncodeOptions` fields as `true` / non-zero now receive `ErrNotSupported` instead of a silent encode.
- `DecodeDataURI` now rejects non-`image/*` types and unknown parameters. Previously a `text/plain` or `;charset=notbase64` URI could be misread.
- Truncated GIF that still starts with `GIF8xa` now fails closed via `ErrDecoder` rather than a one-frame fallback.

## Known limitations

- WebP encode is lossless VP8L only; there is no lossy WebP, no WebP animation, and `Quality` is unused.
- JPEG encode supports quality only. Progressive JPEG, PNG palette/bit depth/interlace, and ICC/EXIF rewrite are not implemented.
- EXIF handling is orientation (tag 0x0112) on JPEG decode. Other EXIF tags are not preserved through encode.
- Decode of `io.Reader` still buffers up to `MaxInputBytes` (or the full stream when unlimited). This is not a pixel-streaming pipeline.
- `Image` is not safe for concurrent mutation; `Clone()` before sharing one source across goroutines.
- AVIF, HEIC, and JPEG 2000 remain `ErrNotSupported`.
- Default `Limits` are unlimited. Server handlers should set caps explicitly (see the HTTP cookbook).

## Documentation

README (English and 简体中文), GoDoc, and VitePress describe Limits, the encode-option matrix, Data URI rules, GIF/EXIF failure modes, and the known limitations above.

---

# v0.1.8 - Remove Manager driver abstraction

`Manager` and `Driver()` are gone. Package defaults live in `DefaultConfig()`. Reuse a setup with `WithConfig` on `Open`, `Decode`, `New`, and `Animate`.

```go
cfg := goimage.DefaultConfig()
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

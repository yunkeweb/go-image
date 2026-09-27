# Core Design

go-image is a modern Go image library. The public surface is package-level functions, functional options, strong types, and a delayed error on `Image`. Internals reuse NRGBA buffers through `sync.Pool`. There is no CGO and no global lock on the pixel path.

## Package-level first

```go
img := goimage.Open("photo.jpg", goimage.WithAutoOrientation(true))
canvas := goimage.New(800, 600)
raw := goimage.Decode(reader)
```

Shared decode settings live in `Config`. Call `DefaultConfig()` (or build a `Config` value) and pass `WithConfig` into `Open` / `Decode` / `New` / `Animate`. `DefaultConfig` is a function that returns a copy; the package default is immutable.

`Fill(c)` paints the whole canvas. Flood fill is `FloodFill(x, y, c)`. `Frame.Image()` returns an independent copy of frame pixels.

## Functional options

`opts ...Option` works on `Open`, `Decode`, `DecodeBytes`, `FromImage`, `New`, and `Animate`. Each call copies `DefaultConfig()` and applies the functions. Concurrent callers never share mutable global state. `WithLimits` caps input bytes, width, height, total pixels, and GIF frames; zero fields are unlimited.

Geometry helpers (`Cover`, `Contain`, `Pad`, `Crop`, `Fit`, `ResizeCanvas`) take `...GeometryOption`:

| Option | Applies to | Default |
|--------|------------|---------|
| `WithAnchor(AnchorCenter)` | Cover, Contain, Pad, Crop, canvas | `AnchorCenter` (Crop: `AnchorTopLeft`) |
| `WithBackground(color)` | Contain, Pad, Crop, canvas | `"ffffff"` |
| `WithOffset(x, y)` | Crop | `0, 0` |

Resize-family methods use a variadic height instead of an option: `Resize(400)` keeps aspect ratio; `Resize(400, 300)` sets both sides.

## Typed inputs

Go has no function overloading. Entry points are named by the source type:

| Function | Type |
|----------|------|
| `Open` | `string` path |
| `Decode` | `io.Reader` |
| `DecodeBytes` | `[]byte` |
| `DecodeDataURI` | `data:image/...` URI |
| `FromImage` | `image.Image` |
| `New` | `int, int` canvas |

`Place` takes any `image.Image` and an `Anchor`. Offset and opacity are options:

```go
img.Place(goimage.Open("logo.png"), goimage.AnchorBottomRight, goimage.WithOffset(16, 16), goimage.WithOpacity(70))
```

Unknown anchors and illegal colors set `Image.Err()`. They are never silently replaced with `top-left` or white.

Preferred names are `New`, `Cover`, `Add`, `SetLoops`, `ToBMP`, `ToJPEG`, `Orient`. Aliases such as `Create`, `Fit`, `AddImage`, `ToBitmap`, and `Orientate` remain but are deprecated.

## Delayed errors and goroutines

Chains return `*Image`. The first error sticks on that value; later modifiers become no-ops. An `Image` is **not** safe for concurrent mutation. Process independent files in parallel, or `Clone()` before sharing one source.

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

There is no `sync.Mutex` on `Image`. Isolation comes from ownership: each goroutine either opens its own file or works on a `Clone`.

## Owned buffers and `sync.Pool`

`FromImage` and every decoder copy into a library-owned NRGBA with `draw.Draw`. JPEG YCbCr, paletted GIF, RGBA, and NRGBA inputs never panic on a type assertion and never share the caller's `Pix` slice.

Discarded NRGBA buffers return to `sync.Pool` when their capacity is at most 16 MiB. Larger frames are released to the garbage collector so a 50-megapixel decode cannot pin a huge slice for later 64×64 work.

## Package layout

```go
import goimage "github.com/yunkeweb/go-image"
```

| Path | Role |
|------|------|
| Root (`image.go`, `options.go`, `api.go`) | Public types and thin wrappers |
| `modifier/` | Geometry, effects, drawing, GIF layout |
| `encoder/` | JPEG, PNG, GIF, WebP, BMP, TIFF |
| `internal/pool`, `internal/color` | Buffer reuse and HTML color names |

Application code imports only the root module.

## Geometry conventions

- `Resize(400)` / `Scale(400)` omit height and keep aspect ratio.
- `Resize(400, 300)` sets both sides. Zero or negative sizes yield `ErrInvalidDimensions`. `0` is never “auto”.
- Cover, Contain, Pad, Crop, Fit, and ResizeCanvas take `WithAnchor`, `WithBackground`, and `WithOffset`.
- Nine-point pivots: `AnchorCenter`, `AnchorTop`, `AnchorTopLeft`, `AnchorTopRight`, `AnchorLeft`, `AnchorRight`, `AnchorBottom`, `AnchorBottomLeft`, `AnchorBottomRight`. String literals still compile.
- `Rotate` is counter-clockwise.
- `Cover` / `Fit` fill a box and crop overflow. `Contain` / `Pad` letterbox.
- Resampling uses Catmull-Rom (`golang.org/x/image/draw`).

Draw and text take `func(*Drawable)` and `func(*Font)` (or a `*Font` value).

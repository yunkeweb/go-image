# Errors

go-image stores the first failure on the `Image` value so chains stay fluent. That model is safe for goroutines: each `*Image` owns its own error and pixel buffers.

## Signature

```go
func (img *Image) Err() error
```

## Behavior

1. Mutating methods return the same `*Image`.
2. If `Err()` is already set, later modifiers are no-ops.
3. Encode helpers return `EncodedImage`. Use `enc.Err()` or the error from `Save`.
4. Sentinel kinds can be inspected with `errors.Is`.

```go
img := goimage.Open("missing.jpg").Cover(200, 200)
if err := img.Err(); err != nil {
    if errors.Is(err, goimage.ErrDecoder) {
        log.Fatal("could not decode:", err)
    }
    log.Fatal(err)
}
```

## Sentinel errors

| Variable | Typical cause |
|----------|----------------|
| `ErrRuntime` | Nil image or unexpected state |
| `ErrDecoder` | Unreadable input |
| `ErrEncoder` | Encode or save path failure |
| `ErrGeometry` | Invalid size (width/height) |
| `ErrInvalidDimensions` | Width or height is `0` or negative |
| `ErrColor` | Unparseable color |
| `ErrInput` | Bad argument (animation index, etc.) |
| `ErrNotSupported` | AVIF / HEIC / JPEG 2000, or unknown format |
| `ErrNotWritable` | Filesystem write failure |
| `ErrAnimation` | Empty animation builder |
| `ErrFont` | Font file load failure |
| `ErrDriver` | Driver-level failure |

## EncodedImage

```go
enc := img.ToWebP()
if err := enc.Err(); err != nil {
    log.Fatal(err)
}
_ = enc.Bytes()
_ = enc.MimeType()
_ = enc.ToDataURI()
```

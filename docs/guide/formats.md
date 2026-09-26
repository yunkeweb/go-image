# Formats

Go uses a single NRGBA driver. There is no GD or Imagick backend.

## Decode / encode matrix

| Format | Decode | Encode | Notes |
|--------|--------|--------|-------|
| JPEG | yes | yes | Quality 1–100, default 75. Progressive is accepted and ignored (stdlib writes baseline). Auto-orientation applies EXIF 2–8. |
| PNG | yes | yes | |
| GIF | still + animated | still + animated | Frame delay is seconds. |
| WebP | yes (`x/image/webp`) | lossless VP8L | Custom encoder, no CGO. |
| BMP | yes | yes | `x/image/bmp` |
| TIFF | yes | yes | `x/image/tiff` |
| AVIF | no | no | Delayed `ErrNotSupported` |
| HEIC | no | no | Delayed `ErrNotSupported` |
| JPEG 2000 | no | no | Delayed `ErrNotSupported` |

## Identifiers

`Encode`, `EncodeByMediaType`, `EncodeByExtension`, and `EncodeByPath` accept aliases such as `jpg`, `image/jpeg`, `.png`, `tif`.

```go
enc := img.EncodeByExtension("webp")
if err := enc.Save("out.webp"); err != nil {
    log.Fatal(err)
}
```

## Transparency

JPEG has no alpha. Encoding JPEG composites transparent pixels over `BlendingColor` (default white). PNG, GIF, WebP, BMP, and TIFF keep alpha when present.

## WebP

Decode uses `golang.org/x/image/webp`. Encode writes a lossless VP8L bitstream wrapped in RIFF/`WEBP`/`VP8L`. Lossy VP8 is not produced.

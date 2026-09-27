# Encoders & Decoders

go-image decodes JPEG, PNG, GIF, lossless WebP (VP8L), BMP, and TIFF with the Go standard library and `golang.org/x/image`. Encoding writes the same set. AVIF, HEIC, and JPEG 2000 return `ErrNotSupported`. There is no CGO.

## Overview

`Open` / `Decode` / `DecodeBytes` sniff the magic bytes, then copy pixels into owned NRGBA. GIF may keep every frame when `WithDecodeAnimation(true)` (the default). JPEG EXIF orientation 2–8 is applied when `WithAutoOrientation(true)` (also the default).

Encode helpers return `EncodedImage`: bytes, media type, `WriteTo`, `Save`, and `ToDataURI`. `ToJPEG` takes a variadic quality (`ToJPEG()` → 80, `ToJPEG(95)` → 95). Other formats take `EncodeOptions`.

## Decode options

```go
func Open(path string, opts ...Option) *Image
func Decode(r io.Reader, opts ...Option) *Image
func DecodeBytes(data []byte, opts ...Option) *Image
func DecodeDataURI(uri string, opts ...Option) *Image
func FromImage(src image.Image, opts ...Option) *Image
func New(width, height int, opts ...Option) *Image
```

| Option | Type | Default | Meaning |
|--------|------|---------|---------|
| `WithAutoOrientation` | `bool` | `true` | Apply JPEG EXIF orientation 2–8 after decode |
| `WithDecodeAnimation` | `bool` | `true` | Keep all GIF frames; `false` keeps the first frame |
| `WithBlendingColor` | `any` | `"ffffff"` | Flatten color for JPEG encode |
| `WithStrip` | `bool` | `false` | Drop ICC profile bytes on encode |
| `WithLimits` | `Limits` | all zeros | Decode resource caps; zero fields are unlimited |

`Limits` is checked in `Open`, `Decode`, `DecodeBytes`, `DecodeDataURI`, `New`, and `FromImage` before pixel buffers are allocated. `MaxPixels` is `width × height` for still images and `width × height × frames` for GIF. Reader input is capped with `io.LimitReader`. Integer overflow in the pixel product is over-limit. Over-limit input returns `ErrLimit`.

If GIF frame counting fails (truncated or corrupt GIF), decode returns `ErrDecoder`. The input is not treated as a still image.

## Data URI

`DecodeDataURI` accepts `data:[mediatype][;base64],data`.

| Rule | Behavior |
|------|----------|
| Media type | Must be `image/*` (case-insensitive type). Other types return `ErrDecoder` |
| `;base64` | Flag parameter, case-insensitive. Detected by splitting parameters, not substring search |
| Payload | URL percent-decoded. Illegal `%` sequences return `ErrDecoder` |
| Base64 | Standard, no-padding, and newline-wrapped forms |
| Empty payload | `ErrDecoder` |
| Unknown parameters | `ErrDecoder` (`charset` is allowed) |

## Encode API

```go
func (img *Image) ToJPEG(quality ...int) EncodedImage
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage
func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage

func (e EncodedImage) Err() error
func (e EncodedImage) Result() ([]byte, error)
func (e EncodedImage) Bytes() []byte
func (e EncodedImage) MimeType() string
func (e EncodedImage) WriteTo(w io.Writer) (int64, error)
func (e EncodedImage) Save(path string) error
func (e EncodedImage) ToDataURI() string
```

`Err`, `Result`, `WriteTo`, and `Save` report the same delayed encode failure. Check one of them before using `Bytes` or `ToDataURI`.

| Field / arg | Type | Default | Notes |
|-------------|------|---------|-------|
| `ToJPEG` quality | `...int` | `80` | Clamped 1–100; `<= 0` uses 80 |
| `EncodeOptions.Quality` | `int` | `80` | JPEG only. Non-zero on PNG/GIF/WebP/BMP/TIFF is `ErrNotSupported` |
| `EncodeOptions.Progressive` | `bool` | `false` | Unsupported (`ErrNotSupported` when true) |
| `EncodeOptions.Indexed` | `bool` | `false` | Unsupported when true |
| `EncodeOptions.Interlaced` | `bool` | `false` | Unsupported when true |
| `EncodeOptions.Bitdepth` | `int` | `0` | Unsupported when non-zero |
| WebP | | | Lossless VP8L only; `Quality` is not applied |

`ToAVIF`, `ToHEIC`, and `ToJPEG2000` return an `EncodedImage` whose `Err()` is `ErrNotSupported`.

## Example: package-level JPEG with quality 85

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg").Cover(800, 600)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	enc := img.ToJPEG(85) // explicit quality
	if err := enc.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println(enc.MimeType(), enc.Size())
	if err := enc.Save("out.jpg"); err != nil {
		log.Fatal(err)
	}
}
```

## Example: functional options + WebP / PNG

```go
package main

import (
	"log"

	goimage "github.com/yunkeweb/go-image"
)

func main() {
	img := goimage.Open("photo.jpg", goimage.WithStrip(true)).
		Scale(1200)
	if err := img.Err(); err != nil {
		log.Fatal(err)
	}

	webp := img.ToWebP() // lossless VP8L
	if err := webp.Save("out.webp"); err != nil {
		log.Fatal(err)
	}

	png := img.ToPNG()
	if err := png.Save("out.png"); err != nil {
		log.Fatal(err)
	}
}
```

## Notes

- JPEG encode writes no EXIF APP1 segment, so orientation tags cannot double-apply after `Orient()`.
- GIF encode keeps `DisposalNone` for opaque full-canvas frames and uses `DisposalBackground` when a frame has alpha or incomplete coverage.
- `WriteTo` implements `io.WriterTo` and is the HTTP path; see [HTTP Handler](/cookbook/http).

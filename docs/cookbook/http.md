# HTTP Handler

`Decode` reads any `io.Reader` (an upload body). `EncodedImage.WriteTo` implements `io.WriterTo` and streams bytes to `http.ResponseWriter`. Set `Content-Type` from `MimeType()`.

## Pipeline

1. Limit the request body (`http.MaxBytesReader`).
2. `Decode(r, WithAutoOrientation(true), WithDecodeAnimation(false))`.
3. `Cover` / `Sharpen` for the variant you need.
4. `ToJPEG(85)` or `ToWebP()`.
5. On `enc.Err()`, return 400/500; otherwise `WriteTo(w)`.

## Example: package-level thumbnail endpoint

```go
package main

import (
	"log"
	"net/http"

	goimage "github.com/yunkeweb/go-image"
)

func thumbnail(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	img := goimage.Decode(r.Body, goimage.WithLimits(goimage.Limits{
		MaxInputBytes: 12 << 20,
		MaxWidth:      8000,
		MaxHeight:     8000,
		MaxPixels:     20_000_000,
		MaxFrames:     24,
	})).Cover(400, 300).Sharpen(8)
	if err := img.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	enc := img.ToJPEG(85)
	if enc.Err() != nil {
		http.Error(w, enc.Err().Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", enc.MimeType())
	if _, err := enc.WriteTo(w); err != nil {
		log.Println("write", err)
	}
}

func main() {
	http.HandleFunc("/thumb", thumbnail)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Example: functional options + WebP query flag

```go
package main

import (
	"net/http"

	goimage "github.com/yunkeweb/go-image"
)

func variant(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
	img := goimage.Decode(r.Body,
		goimage.WithAutoOrientation(true),
		goimage.WithDecodeAnimation(false),
	).Cover(800, 600, goimage.WithAnchor(goimage.AnchorCenter))
	if err := img.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var enc goimage.EncodedImage
	if r.URL.Query().Get("fmt") == "webp" {
		enc = img.ToWebP()
	} else {
		enc = img.ToJPEG(85)
	}
	if enc.Err() != nil {
		http.Error(w, enc.Err().Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", enc.MimeType())
	_, _ = enc.WriteTo(w)
}
```

## Notes

- `WriteTo` returns the sticky encode error without writing if `enc.Err() != nil`.
- Prefer `Decode` over `DecodeBytes(io.ReadAll(...))` so the library reads the body once internally.
- Data URIs: `enc.ToDataURI()` for JSON APIs that embed images.

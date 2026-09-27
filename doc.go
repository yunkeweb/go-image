// Package goimage is a fluent image processing library for Go.
//
// Algorithms live in the modifier subpackage, codecs in encoder, and
// buffer reuse under internal. Callers import this root package only.
//
// It creates, decodes, transforms, draws on, and encodes raster images using
// the Go standard library and golang.org/x/image. The first failure in a
// chain is stored on the Image value and retrieved with Image.Err().
//
// # Decoding
//
// Open a file, decode a reader or byte slice, wrap a standard-library image,
// or allocate a canvas:
//
//	img := goimage.Open("photo.jpg")
//	img = goimage.Decode(r)
//	img = goimage.DecodeBytes(raw)
//	img = goimage.FromImage(stdImg)
//	canvas := goimage.New(800, 600)
//
// DecodeDataURI accepts data:image/... URIs. The media type must be image/*.
// ;base64 is a case-insensitive flag parameter. The payload is percent-decoded.
//
// WithLimits caps MaxInputBytes, MaxWidth, MaxHeight, MaxPixels, and MaxFrames
// on Open, Decode, DecodeBytes, DecodeDataURI, New, and FromImage. Zero fields
// are unlimited. Over-limit input sets ErrLimit before large pixel buffers are
// allocated. Truncated GIF is not decoded as a still image.
//
// Decode and FromImage always copy pixels into an owned NRGBA buffer with
// draw.Draw. JPEG YCbCr, Paletted GIF frames, NRGBA, and RGBA inputs never
// panic on a type assertion and never share the caller's Pix slice.
//
// # Concurrency
//
// Image implements image.Image using the first frame as the static view.
// Frames, Native, Exif, and Profile return copies; use the Unsafe* methods
// only when the caller will not mutate the result.
//
// An Image value is not safe for concurrent mutation. Independent images may
// be processed in parallel. When several goroutines must work from the same
// source, clone first:
//
//	base := goimage.Open("photo.jpg")
//	go func() { _ = base.Clone().Cover(400, 300).ToJPEG().Save("a.jpg") }()
//	go func() { _ = base.Clone().Greyscale().ToPNG().Save("b.png") }()
//
// # HTTP
//
// EncodedImage implements io.WriterTo. Stream directly to an http.ResponseWriter
// or a Gin context:
//
//	enc := img.ToJPEG(85)
//	if enc.Err() != nil {
//	    http.Error(w, enc.Err().Error(), http.StatusInternalServerError)
//	    return
//	}
//	w.Header().Set("Content-Type", enc.MimeType())
//	_, _ = enc.WriteTo(w)
//
// # Example
//
//	img := goimage.Open("photo.jpg").
//		Cover(400, 300).
//		Greyscale().
//		Sharpen(10)
//	if err := img.Err(); err != nil {
//	    log.Fatal(err)
//	}
//	if err := img.ToJPEG(85).Save("out.jpg"); err != nil {
//	    log.Fatal(err)
//	}
package goimage

// Package goimage is a fluent image processing library for Go.
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
// Decode and FromImage always copy pixels into an owned NRGBA buffer with
// draw.Draw. JPEG YCbCr, Paletted GIF frames, NRGBA, and RGBA inputs never
// panic on a type assertion and never share the caller's Pix slice.
//
// # Concurrency
//
// An Image value is not safe for concurrent mutation. Independent images may
// be processed in parallel. When several goroutines must work from the same
// source, clone first:
//
//	base := goimage.Open("photo.jpg")
//	go func() { _ = base.Clone().Cover(400, 300, "center").ToJPEG().Save("a.jpg") }()
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
//		Cover(400, 300, "center").
//		Greyscale().
//		Sharpen(10)
//	if err := img.Err(); err != nil {
//	    log.Fatal(err)
//	}
//	if err := img.ToJPEG(85).Save("out.jpg"); err != nil {
//	    log.Fatal(err)
//	}
package goimage

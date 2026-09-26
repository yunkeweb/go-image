// Package goimage is an idiomatic Go port of Intervention Image (PHP).
//
// It provides a fluent API for creating, decoding, transforming, drawing on,
// and encoding raster images using only the Go standard library and
// golang.org/x/image. PHP exceptions are represented as a delayed error on
// the Image value, retrieved with Image.Err().
//
//	img := goimage.New().Read("photo.jpg").
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

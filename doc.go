// Package goimage is a fluent image processing library for Go.
//
// It creates, decodes, transforms, draws on, and encodes raster images using
// the Go standard library and golang.org/x/image. The first failure in a
// chain is stored on the Image value and retrieved with Image.Err().
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

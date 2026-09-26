// Package encoder implements image codecs used by the goimage root package.
package encoder

import (
	"bytes"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/yunkeweb/go-image/internal/errs"
	"github.com/yunkeweb/go-image/internal/pool"
)

// Options controls encoder-specific knobs such as JPEG quality.
type Options struct {
	Quality     int
	Progressive bool
	Indexed     bool
	Interlaced  bool
	Bitdepth    int
}

// QualityOrDefault returns Quality clamped to 1..100, defaulting to 80.
func (o Options) QualityOrDefault() int {
	if o.Quality <= 0 {
		return 80
	}
	if o.Quality > 100 {
		return 100
	}
	return o.Quality
}

// GIFFrame is one animation frame ready for gif.EncodeAll.
type GIFFrame struct {
	Img        *image.NRGBA
	DelayCS    int
	Disposal   byte
	OffsetLeft int
	OffsetTop  int
}

// StillForOpaque composites src over bg when src has transparent pixels.
func StillForOpaque(src *image.NRGBA, bg image.Image) image.Image {
	if src == nil {
		return pool.Acquire(1, 1)
	}
	if !pool.HasTransparency(src) {
		return src
	}
	dst := pool.Acquire(src.Bounds().Dx(), src.Bounds().Dy())
	if bg != nil {
		draw.Draw(dst, dst.Bounds(), bg, image.Point{}, draw.Src)
	}
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Over)
	return dst
}

// EncodeJPEG writes a baseline JPEG at quality 1..100 (default 80).
func EncodeJPEG(src image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if quality <= 0 {
		quality = 80
	}
	if quality > 100 {
		quality = 100
	}
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: quality}); err != nil {
		return nil, errs.Wrap(errs.ErrEncoder, "jpeg encode: %v", err)
	}
	return buf.Bytes(), nil
}

// EncodePNG writes a PNG.
func EncodePNG(src image.Image) ([]byte, error) {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(&buf, src); err != nil {
		return nil, errs.Wrap(errs.ErrEncoder, "png encode: %v", err)
	}
	return buf.Bytes(), nil
}

// EncodeBMP writes a BMP.
func EncodeBMP(src image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, src); err != nil {
		return nil, errs.Wrap(errs.ErrEncoder, "bmp encode: %v", err)
	}
	return buf.Bytes(), nil
}

// EncodeTIFF writes a TIFF.
func EncodeTIFF(src image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := tiff.Encode(&buf, src, nil); err != nil {
		return nil, errs.Wrap(errs.ErrEncoder, "tiff encode: %v", err)
	}
	return buf.Bytes(), nil
}

// EncodeGIF writes an animated GIF of logical size w×h. Palettes are capped at 256 colors.
func EncodeGIF(frames []GIFFrame, loops, w, h int) ([]byte, error) {
	g := &gif.GIF{
		LoopCount: loops,
		Config:    image.Config{Width: w, Height: h},
	}
	for _, f := range frames {
		src := f.Img
		if src == nil {
			continue
		}
		canvas := pool.Acquire(w, h)
		srcB := src.Bounds()
		pt := image.Pt(f.OffsetLeft, f.OffsetTop)
		draw.Draw(canvas, srcB.Add(pt).Sub(srcB.Min), src, srcB.Min, draw.Over)
		p := QuantizePaletted(canvas, 256)
		pool.Release(canvas)
		g.Image = append(g.Image, p)
		g.Delay = append(g.Delay, f.DelayCS)
		g.Disposal = append(g.Disposal, f.Disposal)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, errs.Wrap(errs.ErrEncoder, "gif encode: %v", err)
	}
	return buf.Bytes(), nil
}

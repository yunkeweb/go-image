package goimage

import (
	"bytes"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
)

func (img *Image) encodeOpts(opts []EncodeOptions) EncodeOptions {
	if len(opts) > 0 {
		return opts[0]
	}
	return EncodeOptions{}
}

func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage {
	if img.fail() {
		return encodedErr(img.err)
	}
	if !format.supportedEncode() {
		return encodedErr(wrap(ErrNotSupported, "encoding %s is not supported by the Go driver", format))
	}
	o := img.encodeOpts(opts)
	src := img
	if img.cfg.Strip {
		src = img.Clone().RemoveProfile()
	}
	var (
		data []byte
		err  error
	)
	switch format {
	case FormatJPEG:
		data, err = encodeJPEG(src, o)
	case FormatPNG:
		data, err = encodePNG(src, o)
	case FormatGIF:
		data, err = encodeGIF(src, o)
	case FormatWEBP:
		data, err = encodeWebP(src, o)
	case FormatBMP:
		data, err = encodeBMP(src)
	case FormatTIFF:
		data, err = encodeTIFF(src)
	default:
		err = wrap(ErrNotSupported, "encoding %s is not supported", format)
	}
	if err != nil {
		return encodedErr(err)
	}
	return EncodedImage{Data: data, MediaType: format.MediaType()}
}

func (img *Image) EncodeByMediaType(mediaType string, opts ...EncodeOptions) EncodedImage {
	if mediaType == "" && img != nil {
		mediaType = img.origin.MediaType
	}
	f, err := parseFormat(mediaType)
	if err != nil {
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

func (img *Image) EncodeByExtension(ext string, opts ...EncodeOptions) EncodedImage {
	if ext == "" && img != nil {
		ext = img.origin.FileExtension()
	}
	f, err := parseFormat(ext)
	if err != nil {
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

func (img *Image) EncodeByPath(path string, opts ...EncodeOptions) EncodedImage {
	if path == "" && img != nil {
		path = img.origin.FilePath
	}
	f, err := formatFromPath(path)
	if err != nil {
		if img != nil && img.origin.MediaType != "" && img.origin.MediaType != "application/octet-stream" {
			return img.EncodeByMediaType(img.origin.MediaType, opts...)
		}
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

func parseEncodeArgs(args []any) EncodeOptions {
	var o EncodeOptions
	for _, a := range args {
		switch v := a.(type) {
		case EncodeOptions:
			o = v
		case *EncodeOptions:
			if v != nil {
				o = *v
			}
		case int:
			o.Quality = v
		case int8:
			o.Quality = int(v)
		case int16:
			o.Quality = int(v)
		case int32:
			o.Quality = int(v)
		case int64:
			o.Quality = int(v)
		case uint:
			o.Quality = int(v)
		case uint8:
			o.Quality = int(v)
		case uint16:
			o.Quality = int(v)
		case uint32:
			o.Quality = int(v)
		case uint64:
			o.Quality = int(v)
		case float64:
			o.Quality = int(v)
		case float32:
			o.Quality = int(v)
		}
	}
	return o
}

// ToJPEG encodes JPEG. Pass an int quality (e.g. ToJPEG(85)) or EncodeOptions.
func (img *Image) ToJPEG(args ...any) EncodedImage {
	return img.Encode(FormatJPEG, parseEncodeArgs(args))
}
func (img *Image) ToJPG(args ...any) EncodedImage            { return img.ToJPEG(args...) }
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage  { return img.Encode(FormatPNG, opts...) }
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage  { return img.Encode(FormatGIF, opts...) }
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatWEBP, opts...) }
func (img *Image) ToBitmap(opts ...EncodeOptions) EncodedImage {
	return img.Encode(FormatBMP, opts...)
}
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage  { return img.ToBitmap(opts...) }
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatTIFF, opts...) }
func (img *Image) ToTIF(opts ...EncodeOptions) EncodedImage  { return img.ToTIFF(opts...) }

func (img *Image) ToJPEG2000(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "JPEG 2000 encoding is not supported by the Go driver"))
}
func (img *Image) ToJP2(opts ...EncodeOptions) EncodedImage { return img.ToJPEG2000(opts...) }
func (img *Image) ToAVIF(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "AVIF encoding is not supported by the Go driver"))
}
func (img *Image) ToHEIC(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "HEIC encoding is not supported by the Go driver"))
}

func stillForOpaque(img *Image) image.Image {
	src := img.primary()
	if src == nil {
		return acquireNRGBA(1, 1)
	}
	if !hasTransparency(src) {
		return src
	}
	bg := img.BlendingColor()
	dst := newBlank(src.Bounds().Dx(), src.Bounds().Dy(), bg)
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Over)
	return dst
}

func hasTransparency(n *image.NRGBA) bool {
	b := n.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if n.NRGBAAt(x, y).A < 255 {
				return true
			}
		}
	}
	return false
}

func encodeJPEG(img *Image, o EncodeOptions) ([]byte, error) {
	var buf bytes.Buffer
	q := o.qualityOrDefault()
	// stdlib jpeg.Encode writes no APP1/EXIF, so orientation tags cannot
	// double-apply after Orient()/auto-orientation reset the in-memory flag to 1.
	if err := jpeg.Encode(&buf, stillForOpaque(img), &jpeg.Options{Quality: q}); err != nil {
		return nil, wrap(ErrEncoder, "jpeg encode: %v", err)
	}
	return buf.Bytes(), nil
}

func encodePNG(img *Image, o EncodeOptions) ([]byte, error) {
	src := img.primary()
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := enc.Encode(&buf, src); err != nil {
		return nil, wrap(ErrEncoder, "png encode: %v", err)
	}
	_ = o
	return buf.Bytes(), nil
}

func encodeBMP(img *Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, stillForOpaque(img)); err != nil {
		return nil, wrap(ErrEncoder, "bmp encode: %v", err)
	}
	return buf.Bytes(), nil
}

func encodeTIFF(img *Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := tiff.Encode(&buf, img.primary(), nil); err != nil {
		return nil, wrap(ErrEncoder, "tiff encode: %v", err)
	}
	return buf.Bytes(), nil
}

func encodeGIF(img *Image, o EncodeOptions) ([]byte, error) {
	_ = o
	w, h := img.Width(), img.Height()
	g := &gif.GIF{
		LoopCount: img.loops,
		Config:    image.Config{Width: w, Height: h},
	}
	for _, f := range img.frames {
		src := f.Img
		if src == nil {
			continue
		}
		canvas := acquireNRGBA(w, h)
		srcB := src.Bounds()
		pt := image.Pt(f.OffsetLeft, f.OffsetTop)
		draw.Draw(canvas, srcB.Add(pt).Sub(srcB.Min), src, srcB.Min, draw.Over)
		p := quantizePaletted(canvas, 256)
		releaseNRGBA(canvas)
		g.Image = append(g.Image, p)
		g.Delay = append(g.Delay, delayToGIF(f.Delay))
		d := byte(f.Dispose)
		if d == 0 {
			d = gif.DisposalBackground
		}
		g.Disposal = append(g.Disposal, d)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, wrap(ErrEncoder, "gif encode: %v", err)
	}
	return buf.Bytes(), nil
}

func quantizePaletted(src *image.NRGBA, limit int) *image.Paletted {
	b := src.Bounds()
	pal := medianCutPalette(src, limit)
	p := image.NewPaletted(b, pal)
	draw.Draw(p, b, src, b.Min, draw.Src)
	return p
}

func medianCutPalette(src *image.NRGBA, limit int) color.Palette {
	if limit < 2 {
		limit = 2
	}
	b := src.Bounds()
	type pix struct{ r, g, bl, a uint8 }
	seen := map[uint32]pix{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			key := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
			if _, ok := seen[key]; !ok {
				seen[key] = pix{c.R, c.G, c.B, c.A}
			}
		}
	}
	pts := make([]pix, 0, len(seen))
	for _, p := range seen {
		pts = append(pts, p)
	}
	if len(pts) == 0 {
		return color.Palette{color.NRGBA{A: 0}}
	}
	if len(pts) <= limit {
		pal := make(color.Palette, 0, len(pts)+1)
		pal = append(pal, color.NRGBA{A: 0})
		for _, p := range pts {
			pal = append(pal, color.NRGBA{R: p.r, G: p.g, B: p.bl, A: p.a})
		}
		return pal
	}
	boxes := [][]pix{pts}
	for len(boxes) < limit {
		bi, channel, spread := -1, 0, 0
		for i, box := range boxes {
			if len(box) < 2 {
				continue
			}
			minR, maxR := 255, 0
			minG, maxG := 255, 0
			minB, maxB := 255, 0
			for _, p := range box {
				if int(p.r) < minR {
					minR = int(p.r)
				}
				if int(p.r) > maxR {
					maxR = int(p.r)
				}
				if int(p.g) < minG {
					minG = int(p.g)
				}
				if int(p.g) > maxG {
					maxG = int(p.g)
				}
				if int(p.bl) < minB {
					minB = int(p.bl)
				}
				if int(p.bl) > maxB {
					maxB = int(p.bl)
				}
			}
			ranges := []int{maxR - minR, maxG - minG, maxB - minB}
			ch, sp := 0, ranges[0]
			for c, r := range ranges {
				if r > sp {
					ch, sp = c, r
				}
			}
			if sp > spread {
				bi, channel, spread = i, ch, sp
			}
		}
		if bi < 0 {
			break
		}
		box := boxes[bi]
		val := func(p pix) int {
			switch channel {
			case 0:
				return int(p.r)
			case 1:
				return int(p.g)
			default:
				return int(p.bl)
			}
		}
		// insertion-ish partition by median
		for i := 1; i < len(box); i++ {
			j := i
			for j > 0 && val(box[j-1]) > val(box[j]) {
				box[j-1], box[j] = box[j], box[j-1]
				j--
			}
		}
		mid := len(box) / 2
		if mid == 0 {
			mid = 1
		}
		left, right := box[:mid], box[mid:]
		boxes[bi] = left
		boxes = append(boxes, right)
	}
	pal := color.Palette{color.NRGBA{A: 0}}
	for _, box := range boxes {
		var r, g, bl, a, n int
		for _, p := range box {
			r += int(p.r)
			g += int(p.g)
			bl += int(p.bl)
			a += int(p.a)
			n++
		}
		if n == 0 {
			continue
		}
		pal = append(pal, color.NRGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(bl / n), A: uint8(a / n)})
	}
	return pal
}

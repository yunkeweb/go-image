package encoder

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/gif"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"

	"github.com/yunkeweb/go-image/internal/errs"
)

// ReadFile reads path and wraps I/O failures as ErrDecoder.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to read file: %v", err)
	}
	return data, nil
}

// ReadFileLimited is ReadFile with an optional MaxInputBytes cap (0 = unlimited).
func ReadFileLimited(path string, maxBytes int64) ([]byte, error) {
	if maxBytes > 0 {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, errs.Wrap(errs.ErrDecoder, "unable to read file: %v", err)
		}
		if fi.Size() > maxBytes {
			return nil, errs.Wrap(errs.ErrLimit, "input exceeds MaxInputBytes (%d)", maxBytes)
		}
	}
	return ReadFile(path)
}

// ReadAll drains r and wraps I/O failures as ErrDecoder.
func ReadAll(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, errs.Wrap(errs.ErrDecoder, "nil reader")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to read input: %v", err)
	}
	return data, nil
}

// ReadAllLimited drains r up to maxBytes+1 (0 = unlimited) and rejects overflow.
func ReadAllLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	if r == nil {
		return nil, errs.Wrap(errs.ErrDecoder, "nil reader")
	}
	if maxBytes <= 0 {
		return ReadAll(r)
	}
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to read input: %v", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, errs.Wrap(errs.ErrLimit, "input exceeds MaxInputBytes (%d)", maxBytes)
	}
	return data, nil
}

// DecodeDataURIPayload extracts the payload of a data:image/... URI.
func DecodeDataURIPayload(s string) ([]byte, error) {
	comma := strings.Index(s, ",")
	if comma < 0 {
		return nil, errs.Wrap(errs.ErrDecoder, "invalid data URI")
	}
	meta := s[:comma]
	payload := s[comma+1:]
	if strings.Contains(meta, ";base64") {
		data, err := decodeBase64Flexible(payload)
		if err != nil {
			return nil, errs.Wrap(errs.ErrDecoder, "invalid data URI base64")
		}
		return data, nil
	}
	return []byte(payload), nil
}

func decodeBase64Flexible(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		return decoded, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}

// DecodeGIF decodes an animated or still GIF.
func DecodeGIF(data []byte) (*gif.GIF, error) {
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to decode gif: %v", err)
	}
	return g, nil
}

// DecodeConfig reads width and height without allocating a pixel buffer.
func DecodeConfig(data []byte) (image.Config, string, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil {
		return cfg, format, nil
	}
	if cfg, err := webp.DecodeConfig(bytes.NewReader(data)); err == nil {
		return cfg, "webp", nil
	}
	if cfg, err := bmp.DecodeConfig(bytes.NewReader(data)); err == nil {
		return cfg, "bmp", nil
	}
	if cfg, err := tiff.DecodeConfig(bytes.NewReader(data)); err == nil {
		return cfg, "tiff", nil
	}
	return image.Config{}, "", err
}

// CountGIFFrames walks GIF image descriptors without decoding pixels.
func CountGIFFrames(data []byte) (int, error) {
	if !IsGIF(data) || len(data) < 13 {
		return 0, errs.Wrap(errs.ErrDecoder, "not a gif")
	}
	packed := data[10]
	i := 13
	if packed&0x80 != 0 {
		i += 3 * (1 << (1 + int(packed&7)))
	}
	n := 0
	for i < len(data) {
		switch data[i] {
		case 0x3b:
			return n, nil
		case 0x21:
			i++
			if i >= len(data) {
				return 0, errs.Wrap(errs.ErrDecoder, "truncated gif")
			}
			i++ // label
			var err error
			i, err = skipGIFSubBlocks(data, i)
			if err != nil {
				return 0, err
			}
		case 0x2c:
			if i+10 > len(data) {
				return 0, errs.Wrap(errs.ErrDecoder, "truncated gif")
			}
			imgPacked := data[i+9]
			i += 10
			if imgPacked&0x80 != 0 {
				i += 3 * (1 << (1 + int(imgPacked&7)))
			}
			if i >= len(data) {
				return 0, errs.Wrap(errs.ErrDecoder, "truncated gif")
			}
			i++ // LZW minimum code size
			var err error
			i, err = skipGIFSubBlocks(data, i)
			if err != nil {
				return 0, err
			}
			n++
		default:
			return 0, errs.Wrap(errs.ErrDecoder, "invalid gif block")
		}
	}
	return n, nil
}

func skipGIFSubBlocks(data []byte, i int) (int, error) {
	for {
		if i >= len(data) {
			return 0, errs.Wrap(errs.ErrDecoder, "truncated gif")
		}
		sz := int(data[i])
		i++
		if sz == 0 {
			return i, nil
		}
		if i+sz > len(data) {
			return 0, errs.Wrap(errs.ErrDecoder, "truncated gif")
		}
		i += sz
	}
}

// DecodeStill decodes a single-frame raster (JPEG, PNG, GIF, WebP, BMP, TIFF).
func DecodeStill(data []byte) (image.Image, string, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	_ = cfg
	if err == nil && (format == "jpeg" || format == "png" || format == "gif") {
		im, format, err := image.Decode(bytes.NewReader(data))
		return im, format, err
	}
	if im, err := webp.Decode(bytes.NewReader(data)); err == nil {
		return im, "webp", nil
	}
	if im, err := bmp.Decode(bytes.NewReader(data)); err == nil {
		return im, "bmp", nil
	}
	if im, err := tiff.Decode(bytes.NewReader(data)); err == nil {
		return im, "tiff", nil
	}
	im, format, err := image.Decode(bytes.NewReader(data))
	return im, format, err
}

// ParseJPEGExif reads Orientation from a JPEG APP1 Exif segment.
// The returned map is nil when no valid Exif orientation is present.
// The integer is always in 1..8 (1 when missing or invalid).
func ParseJPEGExif(data []byte) (map[string]any, int) {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return nil, 1
	}
	i := 2
	for i+4 < len(data) {
		if data[i] != 0xff {
			break
		}
		marker := data[i+1]
		if marker == 0xda {
			break
		}
		if i+4 > len(data) {
			break
		}
		size := int(data[i+2])<<8 | int(data[i+3])
		if size < 2 || i+2+size > len(data) {
			break
		}
		if marker == 0xe1 {
			seg := data[i+4 : i+2+size]
			if bytes.HasPrefix(seg, []byte("Exif\x00\x00")) {
				orient := readExifOrientation(seg[6:])
				if orient < 1 || orient > 8 {
					orient = 1
				}
				m := map[string]any{"Orientation": orient, "IFD0.Orientation": orient}
				return m, orient
			}
		}
		i += 2 + size
	}
	return nil, 1
}

// readExifOrientation parses a TIFF IFD0 Orientation tag (0x0112).
// The TIFF header must be II*\x00 or MM\x00*, the IFD offset and entry count
// must lie inside the buffer, the tag type must be SHORT (3) with count 1,
// and the value must be in 1..8. Any violation returns 1.
func readExifOrientation(tiffData []byte) int {
	if len(tiffData) < 8 {
		return 1
	}
	var le bool
	switch {
	case bytes.HasPrefix(tiffData, []byte("II*\x00")):
		le = true
	case bytes.HasPrefix(tiffData, []byte("MM\x00*")):
		le = false
	default:
		return 1
	}
	u16 := func(b []byte) int {
		if le {
			return int(b[0]) | int(b[1])<<8
		}
		return int(b[0])<<8 | int(b[1])
	}
	u32 := func(b []byte) int {
		if le {
			return int(b[0]) | int(b[1])<<8 | int(b[2])<<16 | int(b[3])<<24
		}
		return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	}
	ifd0 := u32(tiffData[4:8])
	if ifd0 < 8 || ifd0+2 > len(tiffData) {
		return 1
	}
	n := u16(tiffData[ifd0 : ifd0+2])
	if n < 0 {
		return 1
	}
	entryBytes := n * 12
	if ifd0+2+entryBytes > len(tiffData) {
		return 1
	}
	off := ifd0 + 2
	for i := 0; i < n; i++ {
		tag := u16(tiffData[off : off+2])
		typ := u16(tiffData[off+2 : off+4])
		count := u32(tiffData[off+4 : off+8])
		if tag == 0x0112 {
			if typ != 3 || count != 1 {
				return 1
			}
			orient := u16(tiffData[off+8 : off+10])
			if orient < 1 || orient > 8 {
				return 1
			}
			return orient
		}
		off += 12
	}
	return 1
}

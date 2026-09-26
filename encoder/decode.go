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

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to read file: %v", err)
	}
	return data, nil
}

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

func DecodeGIF(data []byte) (*gif.GIF, error) {
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return nil, errs.Wrap(errs.ErrDecoder, "unable to decode gif: %v", err)
	}
	return g, nil
}

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
				m := map[string]any{"Orientation": orient, "IFD0.Orientation": orient}
				return m, orient
			}
		}
		i += 2 + size
	}
	return nil, 1
}

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
	off := ifd0 + 2
	for i := 0; i < n; i++ {
		if off+12 > len(tiffData) {
			break
		}
		tag := u16(tiffData[off : off+2])
		typ := u16(tiffData[off+2 : off+4])
		if tag == 0x0112 {
			if typ == 3 {
				return u16(tiffData[off+8 : off+10])
			}
			return u16(tiffData[off+8 : off+10])
		}
		off += 12
	}
	return 1
}

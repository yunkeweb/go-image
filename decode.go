package goimage

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
)

func decodeFile(path string, cfg Config) (*Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, wrap(ErrDecoder, "unable to read file: %v", err)
	}
	return decodeBytes(data, path, cfg)
}

func decodeReader(r io.Reader, cfg Config) (*Image, error) {
	if r == nil {
		return nil, wrap(ErrDecoder, "nil reader")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, wrap(ErrDecoder, "unable to read input: %v", err)
	}
	return decodeBytes(data, "", cfg)
}

func fromStdImage(src image.Image, cfg Config) (*Image, error) {
	if src == nil {
		return nil, wrap(ErrDecoder, "nil image")
	}
	img := newImage([]Frame{{Img: asNRGBA(src)}}, cfg)
	img.origin = Origin{MediaType: "application/octet-stream"}
	return img, nil
}

func decodeDataURI(s string, cfg Config) (*Image, error) {
	comma := strings.Index(s, ",")
	if comma < 0 {
		return nil, wrap(ErrDecoder, "invalid data URI")
	}
	meta := s[:comma]
	payload := s[comma+1:]
	var data []byte
	var err error
	if strings.Contains(meta, ";base64") {
		data, err = decodeBase64Flexible(payload)
		if err != nil {
			return nil, wrap(ErrDecoder, "invalid data URI base64")
		}
	} else {
		data = []byte(payload)
	}
	return decodeBytes(data, "", cfg)
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

func decodeBytes(data []byte, path string, cfg Config) (*Image, error) {
	if len(data) == 0 {
		return nil, wrap(ErrDecoder, "empty input")
	}
	origin := Origin{FilePath: path, MediaType: sniffMediaType(data)}

	if isGIF(data) {
		g, err := gif.DecodeAll(bytes.NewReader(data))
		if err != nil {
			return nil, wrap(ErrDecoder, "unable to decode gif: %v", err)
		}
		img := imageFromGIF(g, cfg, origin)
		if img.Err() != nil {
			return nil, img.Err()
		}
		return img, nil
	}

	decoded, format, err := decodeStill(data)
	if err != nil {
		return nil, wrap(ErrDecoder, "unable to decode input: %v", err)
	}
	if origin.MediaType == "application/octet-stream" && format != "" {
		if f, ferr := parseFormat(format); ferr == nil {
			origin.MediaType = f.MediaType()
		}
	}
	img := newImage([]Frame{{Img: asNRGBA(decoded)}}, cfg)
	img.origin = origin
	if format == "jpeg" {
		if exif, orient := parseJPEGExif(data); exif != nil {
			img.exif = exif
			if cfg.AutoOrientation {
				applyOrientation(img, orient)
				img.markOrientationNormal()
			}
		}
	}
	return img, nil
}

func decodeStill(data []byte) (image.Image, string, error) {
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

func isGIF(data []byte) bool {
	return len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")))
}

func sniffMediaType(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G'}):
		return "image/png"
	case isGIF(data):
		return "image/gif"
	case bytes.HasPrefix(data, []byte("RIFF")) && len(data) >= 12 && bytes.Equal(data[8:12], []byte("WEBP")):
		return "image/webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return "image/bmp"
	case bytes.HasPrefix(data, []byte("II*\x00")) || bytes.HasPrefix(data, []byte("MM\x00*")):
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

func parseJPEGExif(data []byte) (map[string]any, int) {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return nil, 1
	}
	i := 2
	for i+4 < len(data) {
		if data[i] != 0xff {
			break
		}
		marker := data[i+1]
		if marker == 0xda { // SOS
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
		if tag == 0x0112 { // Orientation
			if typ == 3 {
				return u16(tiffData[off+8 : off+10])
			}
			return u16(tiffData[off+8 : off+10])
		}
		off += 12
	}
	return 1
}

func applyOrientation(img *Image, orient int) {
	switch orient {
	case 2:
		img.Flop()
	case 3:
		img.Rotate(180, "ffffff")
	case 4:
		img.Rotate(180, "ffffff").Flop()
	case 5:
		img.Rotate(270, "ffffff").Flop()
	case 6:
		img.Rotate(270, "ffffff")
	case 7:
		img.Rotate(90, "ffffff").Flop()
	case 8:
		img.Rotate(90, "ffffff")
	}
}

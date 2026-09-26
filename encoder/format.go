package encoder

import (
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-image/internal/errs"
)

const (
	JPEG = "jpeg"
	PNG  = "png"
	GIF  = "gif"
	WEBP = "webp"
	BMP  = "bmp"
	TIFF = "tiff"
	AVIF = "avif"
	HEIC = "heic"
	JP2  = "jp2"
)

func MediaType(format string) string {
	switch format {
	case JPEG:
		return "image/jpeg"
	case PNG:
		return "image/png"
	case GIF:
		return "image/gif"
	case WEBP:
		return "image/webp"
	case BMP:
		return "image/bmp"
	case TIFF:
		return "image/tiff"
	case AVIF:
		return "image/avif"
	case HEIC:
		return "image/heic"
	case JP2:
		return "image/jp2"
	default:
		return "application/octet-stream"
	}
}

func FileExtension(format string) string {
	switch format {
	case JPEG:
		return "jpg"
	case PNG:
		return "png"
	case GIF:
		return "gif"
	case WEBP:
		return "webp"
	case BMP:
		return "bmp"
	case TIFF:
		return "tiff"
	case AVIF:
		return "avif"
	case HEIC:
		return "heic"
	case JP2:
		return "jp2"
	default:
		return ""
	}
}

func SupportedEncode(format string) bool {
	switch format {
	case JPEG, PNG, GIF, WEBP, BMP, TIFF:
		return true
	default:
		return false
	}
}

func ParseFormat(identifier string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(identifier))
	s = strings.TrimPrefix(s, ".")
	switch s {
	case "jpg", "jpeg", "pjpg", "pjpeg", "image/jpeg", "image/jpg", "image/pjpeg", "image/x-jpeg":
		return JPEG, nil
	case "png", "image/png", "image/x-png":
		return PNG, nil
	case "gif", "image/gif":
		return GIF, nil
	case "webp", "image/webp", "image/x-webp":
		return WEBP, nil
	case "bmp", "image/bmp", "image/ms-bmp", "image/x-bitmap", "image/x-bmp",
		"image/x-ms-bmp", "image/x-windows-bmp", "image/x-win-bitmap",
		"image/x-xbitmap", "image/x-bmp3":
		return BMP, nil
	case "tif", "tiff", "image/tiff":
		return TIFF, nil
	case "avif", "image/avif", "image/x-avif":
		return AVIF, nil
	case "heic", "heif", "image/heic", "image/x-heic", "image/heif":
		return HEIC, nil
	case "jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx",
		"image/jp2", "image/x-jp2-codestream", "image/jpx", "image/jpm":
		return JP2, nil
	default:
		return "", errs.Wrap(errs.ErrNotSupported, "unable to create format from %q", identifier)
	}
}

func FormatFromPath(path string) (string, error) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		return "", errs.Wrap(errs.ErrEncoder, "could not determine encoding format from path")
	}
	return ParseFormat(ext)
}

func SniffMediaType(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "image/jpeg"
	case len(data) >= 4 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return "image/png"
	case IsGIF(data):
		return "image/gif"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp"
	case len(data) >= 2 && data[0] == 'B' && data[1] == 'M':
		return "image/bmp"
	case len(data) >= 4 && ((data[0] == 'I' && data[1] == 'I' && data[2] == '*' && data[3] == 0) ||
		(data[0] == 'M' && data[1] == 'M' && data[2] == 0 && data[3] == '*')):
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}

func IsGIF(data []byte) bool {
	return len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a")
}

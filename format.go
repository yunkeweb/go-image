package goimage

import (
	"path/filepath"
	"strings"
)

// Format is an output image format.
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatGIF  Format = "gif"
	FormatWEBP Format = "webp"
	FormatBMP  Format = "bmp"
	FormatTIFF Format = "tiff"
	FormatAVIF Format = "avif"
	FormatHEIC Format = "heic"
	FormatJP2  Format = "jp2"
)

func (f Format) MediaType() string {
	switch f {
	case FormatJPEG:
		return "image/jpeg"
	case FormatPNG:
		return "image/png"
	case FormatGIF:
		return "image/gif"
	case FormatWEBP:
		return "image/webp"
	case FormatBMP:
		return "image/bmp"
	case FormatTIFF:
		return "image/tiff"
	case FormatAVIF:
		return "image/avif"
	case FormatHEIC:
		return "image/heic"
	case FormatJP2:
		return "image/jp2"
	default:
		return "application/octet-stream"
	}
}

func (f Format) FileExtension() string {
	switch f {
	case FormatJPEG:
		return "jpg"
	case FormatPNG:
		return "png"
	case FormatGIF:
		return "gif"
	case FormatWEBP:
		return "webp"
	case FormatBMP:
		return "bmp"
	case FormatTIFF:
		return "tiff"
	case FormatAVIF:
		return "avif"
	case FormatHEIC:
		return "heic"
	case FormatJP2:
		return "jp2"
	default:
		return ""
	}
}

func (f Format) supportedEncode() bool {
	switch f {
	case FormatJPEG, FormatPNG, FormatGIF, FormatWEBP, FormatBMP, FormatTIFF:
		return true
	default:
		return false
	}
}

func parseFormat(identifier string) (Format, error) {
	s := strings.ToLower(strings.TrimSpace(identifier))
	s = strings.TrimPrefix(s, ".")
	switch s {
	case "jpg", "jpeg", "pjpg", "pjpeg", "image/jpeg", "image/jpg", "image/pjpeg", "image/x-jpeg":
		return FormatJPEG, nil
	case "png", "image/png", "image/x-png":
		return FormatPNG, nil
	case "gif", "image/gif":
		return FormatGIF, nil
	case "webp", "image/webp", "image/x-webp":
		return FormatWEBP, nil
	case "bmp", "image/bmp", "image/ms-bmp", "image/x-bitmap", "image/x-bmp",
		"image/x-ms-bmp", "image/x-windows-bmp", "image/x-win-bitmap",
		"image/x-xbitmap", "image/x-bmp3":
		return FormatBMP, nil
	case "tif", "tiff", "image/tiff":
		return FormatTIFF, nil
	case "avif", "image/avif", "image/x-avif":
		return FormatAVIF, nil
	case "heic", "heif", "image/heic", "image/x-heic", "image/heif":
		return FormatHEIC, nil
	case "jp2", "j2k", "jp2k", "jpf", "jpm", "jpg2", "j2c", "jpc", "jpx",
		"image/jp2", "image/x-jp2-codestream", "image/jpx", "image/jpm":
		return FormatJP2, nil
	default:
		return "", wrap(ErrNotSupported, "unable to create format from %q", identifier)
	}
}

func formatFromPath(path string) (Format, error) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if ext == "" {
		return "", wrap(ErrEncoder, "could not determine encoding format from path")
	}
	return parseFormat(ext)
}

// EncodeOptions controls format-specific encoding, matching PHP encoder constructors.
type EncodeOptions struct {
	Quality     int  // JPEG/WebP 0–100; default 80
	Progressive bool // JPEG (best-effort; stdlib writes baseline)
	Indexed     bool // PNG palette
	Interlaced  bool // PNG/GIF
	Bitdepth    int  // PNG
}

func (o EncodeOptions) qualityOrDefault() int {
	if o.Quality <= 0 {
		return 80
	}
	if o.Quality > 100 {
		return 100
	}
	return o.Quality
}

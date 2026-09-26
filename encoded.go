package goimage

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
)

// EncodedImage is encoded binary output plus a media type.
type EncodedImage struct {
	Data      []byte
	MediaType string
	err       error
}

func (e EncodedImage) Err() error { return e.err }

func (e EncodedImage) Bytes() []byte { return e.Data }

func (e EncodedImage) MimeType() string { return e.MediaType }

func (e EncodedImage) Size() int { return len(e.Data) }

func (e EncodedImage) ToDataURI() string {
	return "data:" + e.MediaType + ";base64," + base64.StdEncoding.EncodeToString(e.Data)
}

func (e EncodedImage) String() string { return string(e.Data) }

// WriteTo writes the encoded bytes to w. It implements io.WriterTo so HTTP
// handlers can stream with img.ToJPEG().WriteTo(w).
func (e EncodedImage) WriteTo(w io.Writer) (int64, error) {
	if e.err != nil {
		return 0, e.err
	}
	if w == nil {
		return 0, wrap(ErrEncoder, "nil writer")
	}
	n, err := w.Write(e.Data)
	return int64(n), err
}

func (e EncodedImage) Save(path string) error {
	if e.err != nil {
		return e.err
	}
	if path == "" {
		return wrap(ErrEncoder, "could not determine file path to save")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		if filepath.Dir(path) != "." {
			return wrap(ErrNotWritable, "unable to write %s: %v", path, err)
		}
	}
	if err := os.WriteFile(path, e.Data, 0o644); err != nil {
		return wrap(ErrNotWritable, "unable to write %s: %v", path, err)
	}
	return nil
}

func encodedErr(err error) EncodedImage {
	return EncodedImage{err: err, MediaType: "application/octet-stream"}
}

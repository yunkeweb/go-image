package goimage

import "path/filepath"

// Origin records where an image was created from.
type Origin struct {
	MediaType string
	FilePath  string
}

func (o Origin) MimeType() string { return o.MediaType }

func (o Origin) FileExtension() string {
	if o.FilePath == "" {
		return ""
	}
	ext := filepath.Ext(o.FilePath)
	if ext == "" {
		return ""
	}
	return ext[1:]
}

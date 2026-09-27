package goimage

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"testing"
)

func pngDataURIRaw(t *testing.T) []byte {
	t.Helper()
	raw, err := New(2, 2).Fill("#112233").ToPNG().Result()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestDecodeDataURIStandardAndVariants(t *testing.T) {
	raw := pngDataURIRaw(t)
	std := base64.StdEncoding.EncodeToString(raw)
	nopad := base64.RawStdEncoding.EncodeToString(raw)
	withNL := std[:16] + "\n" + std[16:]
	percented := strings.ReplaceAll(std, "+", "%2B")
	percented = strings.ReplaceAll(percented, "/", "%2F")
	percented = strings.ReplaceAll(percented, "=", "%3D")

	cases := []string{
		"data:image/png;base64," + std,
		"data:image/png;BASE64," + std,
		"data:IMAGE/PNG;base64," + std,
		"data:image/png;charset=UTF-8;base64," + std,
		"data:image/png;base64," + nopad,
		"data:image/png;base64," + withNL,
		"data:image/png;base64," + percented,
	}
	for _, uri := range cases {
		img := DecodeDataURI(uri)
		if img.Err() != nil {
			t.Fatalf("%q: %v", uri, img.Err())
		}
		if img.Width() != 2 || img.Height() != 2 {
			t.Fatalf("%q size %dx%d", uri, img.Width(), img.Height())
		}
	}

	plain := "data:image/png," + url.PathEscape(string(raw))
	img := DecodeDataURI(plain)
	if img.Err() != nil {
		t.Fatal(img.Err())
	}
	if img.Width() != 2 {
		t.Fatalf("plain size %d", img.Width())
	}
}

func TestDecodeDataURIRejects(t *testing.T) {
	raw := pngDataURIRaw(t)
	std := base64.StdEncoding.EncodeToString(raw)
	cases := []struct {
		uri string
	}{
		{"data:text/plain;base64," + std},
		{"data:application/octet-stream;base64," + std},
		{"data:image/png;foo=bar;base64," + std},
		{"data:image/png;base64,!!!!"},
		{"data:image/png;base64,"},
		{"data:image/png,"},
		{"data:image/png,hello%zz"},
		{"data:image/png,hello%"},
		{"data:,AAAA"},
		{"not-a-data-uri"},
	}
	for _, tc := range cases {
		img := DecodeDataURI(tc.uri)
		if img.Err() == nil {
			t.Fatalf("expected error for %q", tc.uri)
		}
		if !errors.Is(img.Err(), ErrDecoder) {
			t.Fatalf("%q: %v", tc.uri, img.Err())
		}
	}
}

func TestDecodeDataURILimits(t *testing.T) {
	raw := pngDataURIRaw(t)
	uri := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
	img := DecodeDataURI(uri, WithLimits(Limits{MaxInputBytes: 8}))
	if img.Err() == nil || !errors.Is(img.Err(), ErrLimit) {
		t.Fatalf("limit: %v", img.Err())
	}
}

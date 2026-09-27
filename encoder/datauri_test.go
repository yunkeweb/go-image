package encoder

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/yunkeweb/go-image/internal/errs"
)

func TestDecodeDataURIPayloadBase64Variants(t *testing.T) {
	raw := []byte("png-bytes")
	std := base64.StdEncoding.EncodeToString(raw)
	rawEnc := base64.RawStdEncoding.EncodeToString(raw)

	cases := []string{
		"data:image/png;base64," + std,
		"data:image/png;BASE64," + std,
		"data:IMAGE/PNG;base64," + std,
		"data:image/png;charset=US-ASCII;base64," + std,
		"data:image/png;base64," + rawEnc,
		"data:image/png;base64," + std[:len(std)/2] + "\n" + std[len(std)/2:],
		"data:image/png;base64," + strings.ReplaceAll(std, "+", "%2B"),
	}
	for _, uri := range cases {
		got, err := DecodeDataURIPayload(uri)
		if err != nil {
			t.Fatalf("%q: %v", uri, err)
		}
		if string(got) != string(raw) {
			t.Fatalf("%q: got %q", uri, got)
		}
	}
}

func TestDecodeDataURIPayloadPercentAndPlain(t *testing.T) {
	got, err := DecodeDataURIPayload("data:image/png,hello%20world")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestDecodeDataURIPayloadRejects(t *testing.T) {
	cases := []string{
		"data:text/plain;base64,aaaa",
		"data:application/json;base64,e30=",
		"data:image/png;foo=bar;base64,aa==",
		"data:image/png;base64,!!!!",
		"data:image/png;base64,",
		"data:image/png,",
		"data:image/png,hello%",
		"data:image/png,hello%zz",
		"data:image/png,hello%2",
		"not-a-uri",
		"data:image/png;base64",
		"data:,aaaa",
	}
	for _, uri := range cases {
		_, err := DecodeDataURIPayload(uri)
		if err == nil {
			t.Fatalf("expected error for %q", uri)
		}
		if !errors.Is(err, errs.ErrDecoder) {
			t.Fatalf("%q: %v", uri, err)
		}
	}
}

func TestDecodeDataURIPayloadDoesNotUseContains(t *testing.T) {
	raw := []byte("abc")
	std := base64.StdEncoding.EncodeToString(raw)
	uri := "data:image/png;charset=notbase64," + std
	got, err := DecodeDataURIPayload(uri)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != std {
		t.Fatalf("charset value containing base64 must not trigger base64 decode, got %q", got)
	}
}

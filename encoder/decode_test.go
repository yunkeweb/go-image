package encoder

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func jpegWithExif(orient int, le bool) []byte {
	tiff := make([]byte, 26)
	if le {
		copy(tiff[0:4], []byte("II*\x00"))
		binary.LittleEndian.PutUint32(tiff[4:8], 8)
		binary.LittleEndian.PutUint16(tiff[8:10], 1)
		binary.LittleEndian.PutUint16(tiff[10:12], 0x0112)
		binary.LittleEndian.PutUint16(tiff[12:14], 3)
		binary.LittleEndian.PutUint32(tiff[14:18], 1)
		binary.LittleEndian.PutUint16(tiff[18:20], uint16(orient))
	} else {
		copy(tiff[0:4], []byte("MM\x00*"))
		binary.BigEndian.PutUint32(tiff[4:8], 8)
		binary.BigEndian.PutUint16(tiff[8:10], 1)
		binary.BigEndian.PutUint16(tiff[10:12], 0x0112)
		binary.BigEndian.PutUint16(tiff[12:14], 3)
		binary.BigEndian.PutUint32(tiff[14:18], 1)
		binary.BigEndian.PutUint16(tiff[18:20], uint16(orient))
	}
	payload := append([]byte("Exif\x00\x00"), tiff...)
	segLen := len(payload) + 2
	out := []byte{0xff, 0xd8, 0xff, 0xe1, byte(segLen >> 8), byte(segLen)}
	out = append(out, payload...)
	return out
}

func TestParseJPEGExifOrientations(t *testing.T) {
	for orient := 1; orient <= 8; orient++ {
		m, got := ParseJPEGExif(jpegWithExif(orient, true))
		if got != orient {
			t.Fatalf("le orient %d got %d", orient, got)
		}
		if m["Orientation"] != orient {
			t.Fatalf("map %v", m)
		}
		_, got = ParseJPEGExif(jpegWithExif(orient, false))
		if got != orient {
			t.Fatalf("be orient %d got %d", orient, got)
		}
	}
}

func jpegAPP0ThenExif(orient int) []byte {
	jfif := []byte{
		0xff, 0xd8,
		0xff, 0xe0, 0x00, 0x10,
		'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
	}
	exif := jpegWithExif(orient, true)
	return append(jfif, exif[2:]...)
}

func jpegMultiAPP1(first, second []byte) []byte {
	out := []byte{0xff, 0xd8}
	out = append(out, first[2:]...)
	out = append(out, second[2:]...)
	return out
}

func jpegXMPAPP1() []byte {
	payload := []byte("http://ns.adobe.com/xap/1.0/\x00<x:xmpmeta/>")
	segLen := len(payload) + 2
	out := []byte{0xff, 0xd8, 0xff, 0xe1, byte(segLen >> 8), byte(segLen)}
	return append(out, payload...)
}

func TestParseJPEGExifCorrupt(t *testing.T) {
	if _, got := ParseJPEGExif([]byte{0xff, 0xd8, 0xff, 0xe1, 0x00, 0x08, 'E', 'x'}); got != 1 {
		t.Fatalf("short %d", got)
	}
	badOrder := jpegWithExif(3, true)
	copy(badOrder[10:], []byte("XX"))
	if _, got := ParseJPEGExif(badOrder); got != 1 && !bytes.Contains(badOrder, []byte("Exif")) {
		t.Fatalf("bad order %d", got)
	}
	if got := readExifOrientation([]byte("notaTIFF")); got != 1 {
		t.Fatalf("bad tiff %d", got)
	}
	if got := readExifOrientation(nil); got != 1 {
		t.Fatal("nil")
	}
	tiff := make([]byte, 26)
	copy(tiff[0:4], []byte("II*\x00"))
	binary.LittleEndian.PutUint32(tiff[4:8], 8)
	binary.LittleEndian.PutUint16(tiff[8:10], 1)
	binary.LittleEndian.PutUint16(tiff[10:12], 0x0112)
	binary.LittleEndian.PutUint16(tiff[12:14], 4) // LONG
	binary.LittleEndian.PutUint32(tiff[14:18], 1)
	binary.LittleEndian.PutUint16(tiff[18:20], 6)
	if got := readExifOrientation(tiff); got != 1 {
		t.Fatalf("wrong type %d", got)
	}
	if _, got := ParseJPEGExif(jpegWithExif(9, true)); got != 1 {
		t.Fatalf("orient 9 got %d", got)
	}
	if _, got := ParseJPEGExif(jpegWithExif(0, true)); got != 1 {
		t.Fatalf("orient 0 got %d", got)
	}
	missing := jpegWithExif(1, true)
	if _, got := ParseJPEGExif(missing[:8]); got != 1 {
		t.Fatalf("truncated APP1 %d", got)
	}
	illegalLen := []byte{0xff, 0xd8, 0xff, 0xe1, 0x00, 0x01}
	if _, got := ParseJPEGExif(illegalLen); got != 1 {
		t.Fatalf("illegal length %d", got)
	}
	truncatedJPEG := []byte{0xff, 0xd8, 0xff}
	if _, got := ParseJPEGExif(truncatedJPEG); got != 1 {
		t.Fatalf("truncated jpeg %d", got)
	}
	if _, got := ParseJPEGExif(nil); got != 1 {
		t.Fatal("nil jpeg")
	}
	noOrient := jpegWithExif(1, false)
	copy(noOrient[len(noOrient)-8:], []byte{0x00, 0x10, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01})
	if _, got := ParseJPEGExif(noOrient); got != 1 {
		t.Fatalf("missing orient %d", got)
	}
}

func TestParseJPEGExifAfterOtherAPPAndMultiAPP1(t *testing.T) {
	_, got := ParseJPEGExif(jpegAPP0ThenExif(6))
	if got != 6 {
		t.Fatalf("after APP0 got %d", got)
	}
	xmpThenExif := jpegMultiAPP1(jpegXMPAPP1(), jpegWithExif(8, true))
	_, got = ParseJPEGExif(xmpThenExif)
	if got != 8 {
		t.Fatalf("second APP1 got %d", got)
	}
	firstWins := jpegMultiAPP1(jpegWithExif(3, true), jpegWithExif(7, false))
	_, got = ParseJPEGExif(firstWins)
	if got != 3 {
		t.Fatalf("first APP1 should win, got %d", got)
	}
}

func TestParseJPEGExifRandomNoPanic(t *testing.T) {
	inputs := [][]byte{
		{},
		{0xff},
		{0xff, 0xd8},
		{0xff, 0xd8, 0xff, 0xe1, 0xff, 0xff},
		bytes.Repeat([]byte{0xff}, 64),
		append([]byte{0xff, 0xd8}, bytes.Repeat([]byte{0x00, 0xff}, 40)...),
		jpegWithExif(4, true)[:len(jpegWithExif(4, true))-3],
	}
	for i, in := range inputs {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					t.Fatalf("panic on input %d: %v", i, rec)
				}
			}()
			_, _ = ParseJPEGExif(in)
		}()
	}
}

func TestCountGIFFrames(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	g := &gif.GIF{LoopCount: 0}
	for i := 0; i < 3; i++ {
		p := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
		g.Image = append(g.Image, p)
		g.Delay = append(g.Delay, 5)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatal(err)
	}
	n, err := CountGIFFrames(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("frames %d", n)
	}
	if _, err := CountGIFFrames([]byte("notgif")); err == nil {
		t.Fatal("expected error")
	}
}

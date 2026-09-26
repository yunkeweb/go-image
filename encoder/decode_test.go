package encoder

import (
	"bytes"
	"encoding/binary"
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
	// type not SHORT
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
}

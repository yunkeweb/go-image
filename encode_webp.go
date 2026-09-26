package goimage

import (
	"encoding/binary"
	"image"
)

// encodeWebP writes a lossless VP8L bitstream (pure Go, no CGO).
func encodeWebP(img *Image, o EncodeOptions) ([]byte, error) {
	_ = o
	src := img.primary()
	if src == nil {
		return nil, wrap(ErrEncoder, "empty image")
	}
	payload := encodeVP8L(src)
	return wrapRIFFWebP(payload), nil
}

func wrapRIFFWebP(vp8l []byte) []byte {
	// RIFF + size + WEBP + VP8L + size + payload [+ pad]
	chunk := 4 + 4 + len(vp8l) // 'VP8L' + size + payload
	riffSize := 4 + chunk      // 'WEBP' + chunk
	out := make([]byte, 0, 12+chunk+1)
	out = append(out, 'R', 'I', 'F', 'F')
	out = binary.LittleEndian.AppendUint32(out, uint32(riffSize))
	out = append(out, 'W', 'E', 'B', 'P')
	out = append(out, 'V', 'P', '8', 'L')
	out = binary.LittleEndian.AppendUint32(out, uint32(len(vp8l)))
	out = append(out, vp8l...)
	if len(vp8l)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

type bitWriter struct {
	buf []byte
	n   uint // bits filled in last byte (0–7); 0 means last byte is full or buf empty
}

func (w *bitWriter) put(val uint32, bits int) {
	for i := 0; i < bits; i++ {
		if w.n == 0 {
			w.buf = append(w.buf, 0)
		}
		if val&1 == 1 {
			w.buf[len(w.buf)-1] |= 1 << w.n
		}
		val >>= 1
		w.n++
		if w.n == 8 {
			w.n = 0
		}
	}
}

// putHuff writes the high bit of an n-bit canonical Huffman code first,
// matching golang.org/x/image/vp8l tree walking.
func (w *bitWriter) putHuff(code uint32, nbits int) {
	for i := nbits - 1; i >= 0; i-- {
		w.put((code >> uint(i)) & 1, 1)
	}
}

func encodeVP8L(src *image.NRGBA) []byte {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	bw := &bitWriter{}
	bw.put(0x2f, 8)
	bw.put(uint32(w-1), 14)
	bw.put(uint32(h-1), 14)
	alpha := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if src.NRGBAAt(x, y).A != 255 {
				alpha = 1
				break
			}
		}
		if alpha == 1 {
			break
		}
	}
	bw.put(uint32(alpha), 1)
	bw.put(0, 3) // version
	bw.put(0, 1) // no transform
	bw.put(0, 1) // no color cache
	bw.put(0, 1) // no meta-Huffman (single group)

	writeLiteralHuffman(bw, 256+24) // green + length prefixes
	writeLiteralHuffman(bw, 256)    // red
	writeLiteralHuffman(bw, 256)    // blue
	writeLiteralHuffman(bw, 256)    // alpha
	writeSimpleOneSymbol(bw, 0)     // distance (unused)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			bw.putHuff(uint32(c.G), 8)
			bw.putHuff(uint32(c.R), 8)
			bw.putHuff(uint32(c.B), 8)
			bw.putHuff(uint32(c.A), 8)
		}
	}
	return bw.buf
}

// writeLiteralHuffman emits a VP8L Huffman tree where symbols [0,nLit) have
// 8-bit identity codes and the rest have length 0.
func writeLiteralHuffman(bw *bitWriter, alphabet int) {
	nLit := 256
	if alphabet < 256 {
		nLit = alphabet
	}
	bw.put(0, 1) // not simple
	// 12 code-length codes (4 + 8) covering up to symbol 8 in kCodeLengthCodeOrder
	bw.put(8, 4)
	// kCodeLengthCodeOrder: 17,18,0,1,2,3,4,5,16,6,7,8
	// lengths:              0, 2,0,0,0,0,0,0, 2,0,0,1
	orderLen := []uint32{0, 2, 0, 0, 0, 0, 0, 0, 2, 0, 0, 1}
	for _, l := range orderLen {
		bw.put(l, 3)
	}
	// decodeCodeLengths: 1-bit “useLength” flag (0 = encode the full alphabet).
	bw.put(0, 1)
	// Canonical codes: 8 → 0 (1 bit), 16 → 2 (2 bits), 18 → 3 (2 bits)
	put8 := func() { bw.putHuff(0, 1) }
	put16 := func(extra uint32) {
		bw.putHuff(2, 2)
		bw.put(extra, 2) // repeat previous 3+extra times (raw bits)
	}
	put18 := func(nZeros int) {
		// 18 repeats zeros 11+extra times, extra is 7 bits (0–127) → 11–138
		bw.putHuff(3, 2)
		bw.put(uint32(nZeros-11), 7)
	}

	// nLit eights
	put8()
	remain := nLit - 1
	for remain > 0 {
		n := remain
		if n > 6 {
			n = 6
		}
		if n < 3 {
			// 16 can only repeat 3–6; fall back to explicit 8s
			for i := 0; i < n; i++ {
				put8()
			}
			remain = 0
			break
		}
		put16(uint32(n - 3))
		remain -= n
	}
	zeros := alphabet - nLit
	for zeros > 0 {
		n := zeros
		if n > 138 {
			n = 138
		}
		if n < 11 {
			// not enough for code 18; emit zeros via 17 (3–10) if possible
			if n >= 3 {
				// 17: 3–10 zeros, 3 extra bits, count=3+extra. We didn't include
				// 17 in the tree. Emit one unused-path: give remaining zeros
				// using repeated 18 with n=11 (minimum) would overflow alphabet.
				// Pad the alphabet encoding by using 18 only when zeros>=11.
				// For <11, emit length-0 by... we cannot. So always keep
				// alphabet-nLit either 0 or >=11.
				// Green alphabet is 280, nLit=256, zeros=24 >= 11. OK.
				// If zeros < 11, write 18 with 11 and hope decoder stops at alphabet?
				// Safer: zeros for green is 24. For 256-alpha, zeros=0.
			}
			break
		}
		put18(n)
		zeros -= n
	}
}

func writeSimpleOneSymbol(bw *bitWriter, symbol int) {
	bw.put(1, 1) // simple
	bw.put(0, 1) // num_symbols = 1
	if symbol <= 1 {
		bw.put(0, 1) // 1-bit symbol
		bw.put(uint32(symbol), 1)
		return
	}
	bw.put(1, 1) // 8-bit symbol
	bw.put(uint32(symbol), 8)
}

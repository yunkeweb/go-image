package intcolor

import (
	"image/color"
	"math"
)

func Clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func AbsInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func OverNRGBA(dst, src color.NRGBA) color.NRGBA {
	if src.A == 0 {
		return dst
	}
	if src.A == 255 {
		return src
	}
	sa := float64(src.A) / 255
	da := float64(dst.A) / 255
	outA := sa + da*(1-sa)
	if outA == 0 {
		return color.NRGBA{}
	}
	r := (float64(src.R)*sa + float64(dst.R)*da*(1-sa)) / outA
	g := (float64(src.G)*sa + float64(dst.G)*da*(1-sa)) / outA
	b := (float64(src.B)*sa + float64(dst.B)*da*(1-sa)) / outA
	return color.NRGBA{
		R: Clamp8(int(math.Round(r))),
		G: Clamp8(int(math.Round(g))),
		B: Clamp8(int(math.Round(b))),
		A: Clamp8(int(math.Round(outA * 255))),
	}
}

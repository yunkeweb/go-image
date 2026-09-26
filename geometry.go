package goimage

import (
	"math"
	"strings"
)

// Point is an integer coordinate.
type Point struct {
	X, Y int
}

// Size is a rectangle with an optional 9-point pivot.
type Size struct {
	Width, Height int
	Pivot         Point
}

func (s Size) AspectRatio() float64 {
	if s.Height == 0 {
		return 0
	}
	return float64(s.Width) / float64(s.Height)
}

func (s Size) FitsInto(other Size) bool {
	return s.Width <= other.Width && s.Height <= other.Height
}

func (s Size) IsLandscape() bool { return s.Width > s.Height }
func (s Size) IsPortrait() bool  { return s.Width < s.Height }

func (s Size) MovePivot(position string, offsetX, offsetY int) Size {
	s.Pivot = pivotPoint(s.Width, s.Height, position, offsetX, offsetY)
	return s
}

func (s Size) RelativePositionTo(other Size) Point {
	return Point{X: s.Pivot.X - other.Pivot.X, Y: s.Pivot.Y - other.Pivot.Y}
}

func (s Size) AlignPivotTo(ref Size, position string) Size {
	reference := Size{Width: ref.Width, Height: ref.Height}.MovePivot(position, 0, 0)
	moved := s.MovePivot(position, 0, 0)
	moved.Pivot = reference.RelativePositionTo(moved)
	return moved
}

type resizer struct {
	width  int
	height int
	hasW   bool
	hasH   bool
}

func newResizer(width, height int) (resizer, error) {
	r := resizer{}
	if width == 0 && height == 0 {
		return r, wrap(ErrInvalidDimensions, "width and height cannot both be 0")
	}
	if width != 0 {
		if width < 1 {
			return r, wrap(ErrGeometry, "the width you specify must be greater than or equal to 1")
		}
		r.width = width
		r.hasW = true
	}
	if height != 0 {
		if height < 1 {
			return r, wrap(ErrGeometry, "the height you specify must be greater than or equal to 1")
		}
		r.height = height
		r.hasH = true
	}
	return r, nil
}

func (r resizer) proportionalWidth(size Size) int {
	if !r.hasH {
		return size.Width
	}
	ar := size.AspectRatio()
	if ar == 0 {
		return max(1, r.width)
	}
	return max(1, int(math.Round(float64(r.height)*ar)))
}

func (r resizer) proportionalHeight(size Size) int {
	if !r.hasW {
		return size.Height
	}
	ar := size.AspectRatio()
	if ar == 0 {
		return max(1, r.height)
	}
	return max(1, int(math.Round(float64(r.width)/ar)))
}

func (r resizer) resize(size Size) Size {
	out := size
	if r.hasW {
		out.Width = r.width
	}
	if r.hasH {
		out.Height = r.height
	}
	return out
}

func (r resizer) resizeDown(size Size) Size {
	out := size
	if r.hasW {
		out.Width = min(r.width, size.Width)
	}
	if r.hasH {
		out.Height = min(r.height, size.Height)
	}
	return out
}

func (r resizer) scale(size Size) Size {
	out := size
	switch {
	case r.hasW && r.hasH:
		out.Width = min(r.proportionalWidth(size), r.width)
		out.Height = min(r.proportionalHeight(size), r.height)
	case r.hasW:
		out.Width = r.width
		out.Height = r.proportionalHeight(size)
	case r.hasH:
		out.Width = r.proportionalWidth(size)
		out.Height = r.height
	}
	return out
}

func (r resizer) scaleDown(size Size) Size {
	out := size
	switch {
	case r.hasW && r.hasH:
		out.Width = min(r.proportionalWidth(size), r.width, size.Width)
		out.Height = min(r.proportionalHeight(size), r.height, size.Height)
	case r.hasW:
		out.Width = min(r.width, size.Width)
		out.Height = min(r.proportionalHeight(size), size.Height)
	case r.hasH:
		out.Width = min(r.proportionalWidth(size), size.Width)
		out.Height = min(r.height, size.Height)
	}
	return out
}

func (r resizer) cover(size Size) (Size, error) {
	if !r.hasW || !r.hasH {
		return Size{}, wrap(ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.width, Height: r.height}
	out := size
	out.Width = r.width
	out.Height = r.proportionalHeight(size)
	if out.FitsInto(target) {
		out.Width = r.proportionalWidth(size)
		out.Height = r.height
	}
	return out, nil
}

func (r resizer) contain(size Size) (Size, error) {
	if !r.hasW || !r.hasH {
		return Size{}, wrap(ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.width, Height: r.height}
	out := size
	out.Width = r.width
	out.Height = r.proportionalHeight(size)
	if !out.FitsInto(target) {
		out.Width = r.proportionalWidth(size)
		out.Height = r.height
	}
	return out, nil
}

func (r resizer) containDown(size Size) (Size, error) {
	if !r.hasW || !r.hasH {
		return Size{}, wrap(ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.width, Height: r.height}
	out := size
	out.Width = min(size.Width, r.width)
	out.Height = min(size.Height, r.proportionalHeight(size))
	if !out.FitsInto(target) {
		out.Width = min(size.Width, r.proportionalWidth(size))
		out.Height = min(size.Height, r.height)
	}
	return out, nil
}

func pivotPoint(w, h int, position string, ox, oy int) Point {
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "top", "top-center", "top-middle", "center-top", "middle-top":
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: oy}
	case "top-right", "right-top":
		return Point{X: w - ox, Y: oy}
	case "left", "left-center", "left-middle", "center-left", "middle-left":
		return Point{X: ox, Y: int(math.Round(float64(h)/2)) + oy}
	case "right", "right-center", "right-middle", "center-right", "middle-right":
		return Point{X: w - ox, Y: int(math.Round(float64(h)/2)) + oy}
	case "bottom-left", "left-bottom":
		return Point{X: ox, Y: h - oy}
	case "bottom", "bottom-center", "bottom-middle", "center-bottom", "middle-bottom":
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: h - oy}
	case "bottom-right", "right-bottom":
		return Point{X: w - ox, Y: h - oy}
	case "center", "middle", "center-center", "middle-middle":
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: int(math.Round(float64(h)/2)) + oy}
	default: // top-left / left-top
		return Point{X: ox, Y: oy}
	}
}

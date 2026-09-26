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

type geometrySettings struct {
	anchor     string
	background any
	offsetX    int
	offsetY    int
}

// GeometryOption configures Cover, Contain, Pad, Crop, Fit, and canvas helpers.
type GeometryOption func(*geometrySettings)

// WithAnchor sets the 9-point pivot (center, top-left, bottom-right, …).
func WithAnchor(anchor string) GeometryOption {
	return func(s *geometrySettings) {
		if strings.TrimSpace(anchor) != "" {
			s.anchor = anchor
		}
	}
}

// WithBackground sets the fill color for new canvas pixels.
func WithBackground(color any) GeometryOption {
	return func(s *geometrySettings) {
		if color != nil {
			s.background = color
		}
	}
}

// WithOffset shifts the crop origin after the anchor is applied.
func WithOffset(x, y int) GeometryOption {
	return func(s *geometrySettings) {
		s.offsetX = x
		s.offsetY = y
	}
}

func applyGeometryOptions(base geometrySettings, opts []GeometryOption) geometrySettings {
	for _, o := range opts {
		if o != nil {
			o(&base)
		}
	}
	return base
}

type resizer struct {
	width  int
	height int
	hasW   bool
	hasH   bool
}

func invalidDimensions() error {
	return wrap(ErrInvalidDimensions, "invalid dimensions")
}

// sizeFromArgs builds a resizer from Resize/Scale-style arguments.
// Omitting height keeps the original aspect ratio. Zero and negative values
// are errors; 0 is never treated as “auto”.
func sizeFromArgs(width int, height []int) (resizer, error) {
	if width < 1 {
		return resizer{}, invalidDimensions()
	}
	r := resizer{width: width, hasW: true}
	if len(height) == 0 {
		return r, nil
	}
	if height[0] < 1 {
		return resizer{}, invalidDimensions()
	}
	r.height = height[0]
	r.hasH = true
	return r, nil
}

// newResizer requires both width and height to be >= 1.
func newResizer(width, height int) (resizer, error) {
	if width < 1 || height < 1 {
		return resizer{}, invalidDimensions()
	}
	return resizer{width: width, hasW: true, height: height, hasH: true}, nil
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
	switch {
	case r.hasW && r.hasH:
		out.Width = r.width
		out.Height = r.height
	case r.hasW:
		out.Width = r.width
		out.Height = r.proportionalHeight(size)
	case r.hasH:
		out.Width = r.proportionalWidth(size)
		out.Height = r.height
	}
	return out
}

func (r resizer) resizeDown(size Size) Size {
	out := r.resize(size)
	out.Width = min(out.Width, size.Width)
	out.Height = min(out.Height, size.Height)
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

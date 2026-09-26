// Package modifier implements raster algorithms used by the goimage root package.
package modifier

import (
	"math"
	"strings"

	"github.com/yunkeweb/go-image/internal/errs"
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

// AspectRatio returns Width/Height, or 0 when Height is 0.
func (s Size) AspectRatio() float64 {
	if s.Height == 0 {
		return 0
	}
	return float64(s.Width) / float64(s.Height)
}

// FitsInto reports whether s is smaller than or equal to other on both axes.
func (s Size) FitsInto(other Size) bool {
	return s.Width <= other.Width && s.Height <= other.Height
}

// IsLandscape reports whether Width is greater than Height.
func (s Size) IsLandscape() bool { return s.Width > s.Height }

// IsPortrait reports whether Width is less than Height.
func (s Size) IsPortrait() bool { return s.Width < s.Height }

// MovePivot sets the 9-point pivot named by position, plus an extra offset.
func (s Size) MovePivot(position string, offsetX, offsetY int) Size {
	s.Pivot = PivotPoint(s.Width, s.Height, position, offsetX, offsetY)
	return s
}

// RelativePositionTo returns the vector from other.Pivot to s.Pivot.
func (s Size) RelativePositionTo(other Size) Point {
	return Point{X: s.Pivot.X - other.Pivot.X, Y: s.Pivot.Y - other.Pivot.Y}
}

// AlignPivotTo moves s so its named pivot matches ref's named pivot.
func (s Size) AlignPivotTo(ref Size, position string) Size {
	reference := Size{Width: ref.Width, Height: ref.Height}.MovePivot(position, 0, 0)
	moved := s.MovePivot(position, 0, 0)
	moved.Pivot = reference.RelativePositionTo(moved)
	return moved
}

// Resizer computes target sizes for Resize, Scale, Cover, and Contain.
type Resizer struct {
	Width  int
	Height int
	HasW   bool
	HasH   bool
}

func invalidDimensions() error {
	return errs.Wrap(errs.ErrInvalidDimensions, "invalid dimensions")
}

// SizeFromArgs builds a Resizer from Resize/Scale-style arguments.
// Omitting height keeps the original aspect ratio. Zero and negative values
// are errors; 0 is never treated as “auto”.
func SizeFromArgs(width int, height []int) (Resizer, error) {
	if width < 1 {
		return Resizer{}, invalidDimensions()
	}
	r := Resizer{Width: width, HasW: true}
	if len(height) == 0 {
		return r, nil
	}
	if height[0] < 1 {
		return Resizer{}, invalidDimensions()
	}
	r.Height = height[0]
	r.HasH = true
	return r, nil
}

// NewResizer requires both width and height to be >= 1.
func NewResizer(width, height int) (Resizer, error) {
	if width < 1 || height < 1 {
		return Resizer{}, invalidDimensions()
	}
	return Resizer{Width: width, HasW: true, Height: height, HasH: true}, nil
}

func (r Resizer) proportionalWidth(size Size) int {
	if !r.HasH {
		return size.Width
	}
	ar := size.AspectRatio()
	if ar == 0 {
		return max(1, r.Width)
	}
	return max(1, int(math.Round(float64(r.Height)*ar)))
}

func (r Resizer) proportionalHeight(size Size) int {
	if !r.HasW {
		return size.Height
	}
	ar := size.AspectRatio()
	if ar == 0 {
		return max(1, r.Height)
	}
	return max(1, int(math.Round(float64(r.Width)/ar)))
}

// Resize returns the target size for an absolute resize of size.
func (r Resizer) Resize(size Size) Size {
	out := size
	switch {
	case r.HasW && r.HasH:
		out.Width = r.Width
		out.Height = r.Height
	case r.HasW:
		out.Width = r.Width
		out.Height = r.proportionalHeight(size)
	case r.HasH:
		out.Width = r.proportionalWidth(size)
		out.Height = r.Height
	}
	return out
}

// ResizeDown is Resize that never enlarges size.
func (r Resizer) ResizeDown(size Size) Size {
	out := r.Resize(size)
	out.Width = min(out.Width, size.Width)
	out.Height = min(out.Height, size.Height)
	return out
}

// Scale fits size inside the resizer box, keeping aspect ratio.
func (r Resizer) Scale(size Size) Size {
	out := size
	switch {
	case r.HasW && r.HasH:
		out.Width = min(r.proportionalWidth(size), r.Width)
		out.Height = min(r.proportionalHeight(size), r.Height)
	case r.HasW:
		out.Width = r.Width
		out.Height = r.proportionalHeight(size)
	case r.HasH:
		out.Width = r.proportionalWidth(size)
		out.Height = r.Height
	}
	return out
}

// ScaleDown is Scale that never enlarges size.
func (r Resizer) ScaleDown(size Size) Size {
	out := size
	switch {
	case r.HasW && r.HasH:
		out.Width = min(r.proportionalWidth(size), r.Width, size.Width)
		out.Height = min(r.proportionalHeight(size), r.Height, size.Height)
	case r.HasW:
		out.Width = min(r.Width, size.Width)
		out.Height = min(r.proportionalHeight(size), size.Height)
	case r.HasH:
		out.Width = min(r.proportionalWidth(size), size.Width)
		out.Height = min(r.Height, size.Height)
	}
	return out
}

// Cover returns the size that fills the resizer box, keeping aspect ratio.
func (r Resizer) Cover(size Size) (Size, error) {
	if !r.HasW || !r.HasH {
		return Size{}, errs.Wrap(errs.ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.Width, Height: r.Height}
	out := size
	out.Width = r.Width
	out.Height = r.proportionalHeight(size)
	if out.FitsInto(target) {
		out.Width = r.proportionalWidth(size)
		out.Height = r.Height
	}
	return out, nil
}

// Contain returns the size that fits inside the resizer box, keeping aspect ratio.
func (r Resizer) Contain(size Size) (Size, error) {
	if !r.HasW || !r.HasH {
		return Size{}, errs.Wrap(errs.ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.Width, Height: r.Height}
	out := size
	out.Width = r.Width
	out.Height = r.proportionalHeight(size)
	if !out.FitsInto(target) {
		out.Width = r.proportionalWidth(size)
		out.Height = r.Height
	}
	return out, nil
}

// ContainDown is Contain that never enlarges size.
func (r Resizer) ContainDown(size Size) (Size, error) {
	if !r.HasW || !r.HasH {
		return Size{}, errs.Wrap(errs.ErrGeometry, "target size needs width and height")
	}
	target := Size{Width: r.Width, Height: r.Height}
	out := size
	out.Width = min(size.Width, r.Width)
	out.Height = min(size.Height, r.proportionalHeight(size))
	if !out.FitsInto(target) {
		out.Width = min(size.Width, r.proportionalWidth(size))
		out.Height = min(size.Height, r.Height)
	}
	return out, nil
}

// PivotPoint returns the 9-point anchor for a w×h rectangle.
func PivotPoint(w, h int, position string, ox, oy int) Point {
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

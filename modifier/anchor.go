package modifier

import (
	"math"
	"strings"

	"github.com/yunkeweb/go-image/internal/errs"
)

// Anchor is a 9-point pivot used by Cover, Crop, Place, and related helpers.
type Anchor string

const (
	AnchorTopLeft     Anchor = "top-left"
	AnchorTop         Anchor = "top"
	AnchorTopRight    Anchor = "top-right"
	AnchorLeft        Anchor = "left"
	AnchorCenter      Anchor = "center"
	AnchorRight       Anchor = "right"
	AnchorBottomLeft  Anchor = "bottom-left"
	AnchorBottom      Anchor = "bottom"
	AnchorBottomRight Anchor = "bottom-right"
)

// String returns the canonical hyphenated name.
func (a Anchor) String() string { return string(a) }

// ParseAnchor canonicalizes a 9-point position name or a documented synonym.
// Unknown values return ErrGeometry; they are never treated as top-left.
func ParseAnchor(s string) (Anchor, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "top-left", "left-top":
		return AnchorTopLeft, nil
	case "top", "top-center", "top-middle", "center-top", "middle-top":
		return AnchorTop, nil
	case "top-right", "right-top":
		return AnchorTopRight, nil
	case "left", "left-center", "left-middle", "center-left", "middle-left":
		return AnchorLeft, nil
	case "center", "middle", "center-center", "middle-middle":
		return AnchorCenter, nil
	case "right", "right-center", "right-middle", "center-right", "middle-right":
		return AnchorRight, nil
	case "bottom-left", "left-bottom":
		return AnchorBottomLeft, nil
	case "bottom", "bottom-center", "bottom-middle", "center-bottom", "middle-bottom":
		return AnchorBottom, nil
	case "bottom-right", "right-bottom":
		return AnchorBottomRight, nil
	default:
		return "", errs.Wrap(errs.ErrGeometry, "invalid anchor %q", s)
	}
}

// Canonical returns the named constant for a, or an error when a is unknown.
func (a Anchor) Canonical() (Anchor, error) {
	return ParseAnchor(string(a))
}

// PivotPoint returns the 9-point anchor for a w×h rectangle.
func PivotPoint(w, h int, position Anchor, ox, oy int) Point {
	a, err := position.Canonical()
	if err != nil {
		a = AnchorTopLeft
	}
	switch a {
	case AnchorTop:
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: oy}
	case AnchorTopRight:
		return Point{X: w - ox, Y: oy}
	case AnchorLeft:
		return Point{X: ox, Y: int(math.Round(float64(h)/2)) + oy}
	case AnchorRight:
		return Point{X: w - ox, Y: int(math.Round(float64(h)/2)) + oy}
	case AnchorBottomLeft:
		return Point{X: ox, Y: h - oy}
	case AnchorBottom:
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: h - oy}
	case AnchorBottomRight:
		return Point{X: w - ox, Y: h - oy}
	case AnchorCenter:
		return Point{X: int(math.Round(float64(w)/2)) + ox, Y: int(math.Round(float64(h)/2)) + oy}
	default:
		return Point{X: ox, Y: oy}
	}
}

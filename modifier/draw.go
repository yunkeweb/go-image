package modifier

import (
	"image"
	"image/color"
	"math"

	intcolor "github.com/yunkeweb/go-image/internal/color"
	"github.com/yunkeweb/go-image/internal/pool"
)

// DrawPixel sets (x, y) on a clone of n.
func DrawPixel(n *image.NRGBA, x, y int, c color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	b := dst.Bounds()
	if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
		dst.SetNRGBA(x, y, c)
	}
	return dst
}

// Fill paints every pixel of n with c.
func Fill(n *image.NRGBA, c color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	pool.FillRect(dst, dst.Bounds(), c)
	return dst
}

// FloodFill replaces the 4-connected region of equal color at (x, y).
func FloodFill(n *image.NRGBA, x, y int, c color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	floodFill(dst, x, y, c)
	return dst
}

// DrawRectangle paints a filled and/or stroked rectangle onto a clone of n.
func DrawRectangle(n *image.NRGBA, x, y, w, h int, bg *color.NRGBA, borderSize int, border *color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	r := image.Rect(x, y, x+w, y+h)
	if bg != nil {
		pool.FillRect(dst, r, *bg)
	}
	if borderSize > 0 && border != nil {
		strokeRect(dst, r, borderSize, *border)
	}
	return dst
}

// DrawEllipse paints a filled and/or stroked ellipse onto a clone of n.
func DrawEllipse(n *image.NRGBA, cx, cy, rx, ry int, bg *color.NRGBA, borderSize int, border *color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	if bg != nil {
		fillEllipse(dst, cx, cy, rx, ry, *bg)
	}
	if borderSize > 0 && border != nil {
		for i := 0; i < borderSize; i++ {
			strokeEllipse(dst, cx, cy, rx-i, ry-i, *border)
		}
	}
	return dst
}

// DrawPolygon paints a filled and/or stroked polygon onto a clone of n.
func DrawPolygon(n *image.NRGBA, pts []Point, bg *color.NRGBA, borderSize int, border *color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	if bg != nil {
		fillPolygon(dst, pts, *bg)
	}
	if borderSize > 0 && border != nil && len(pts) > 1 {
		for i := 0; i < len(pts); i++ {
			a := pts[i]
			b := pts[(i+1)%len(pts)]
			drawLineWidth(dst, a.X, a.Y, b.X, b.Y, borderSize, *border)
		}
	}
	return dst
}

// DrawLine paints a stroke of width pixels onto a clone of n.
func DrawLine(n *image.NRGBA, x1, y1, x2, y2, width int, c color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	if width < 1 {
		width = 1
	}
	drawLineWidth(dst, x1, y1, x2, y2, width, c)
	return dst
}

// DrawBezier paints a polyline through pts onto a clone of n.
func DrawBezier(n *image.NRGBA, pts []Point, width int, c color.NRGBA) *image.NRGBA {
	dst := pool.Clone(n)
	if len(pts) < 2 {
		return dst
	}
	if width < 1 {
		width = 1
	}
	steps := 64
	prev := bezierPoint(pts, 0)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		p := bezierPoint(pts, t)
		drawLineWidth(dst, prev.X, prev.Y, p.X, p.Y, width, c)
		prev = p
	}
	return dst
}

func strokeRect(dst *image.NRGBA, r image.Rectangle, width int, c color.NRGBA) {
	for i := 0; i < width; i++ {
		rr := image.Rect(r.Min.X+i, r.Min.Y+i, r.Max.X-i, r.Max.Y-i)
		if rr.Dx() <= 0 || rr.Dy() <= 0 {
			return
		}
		for x := rr.Min.X; x < rr.Max.X; x++ {
			setOver(dst, x, rr.Min.Y, c)
			setOver(dst, x, rr.Max.Y-1, c)
		}
		for y := rr.Min.Y; y < rr.Max.Y; y++ {
			setOver(dst, rr.Min.X, y, c)
			setOver(dst, rr.Max.X-1, y, c)
		}
	}
}

func fillEllipse(dst *image.NRGBA, cx, cy, rx, ry int, c color.NRGBA) {
	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}
	rx2 := float64(rx * rx)
	ry2 := float64(ry * ry)
	b := dst.Bounds()
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if dx*dx/rx2+dy*dy/ry2 <= 1 {
				if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
					setOver(dst, x, y, c)
				}
			}
		}
	}
}

func strokeEllipse(dst *image.NRGBA, cx, cy, rx, ry int, c color.NRGBA) {
	if rx < 1 || ry < 1 {
		return
	}
	steps := int(2 * math.Pi * math.Max(float64(rx), float64(ry)))
	if steps < 16 {
		steps = 16
	}
	var px, py int
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps) * 2 * math.Pi
		x := cx + int(math.Round(float64(rx)*math.Cos(t)))
		y := cy + int(math.Round(float64(ry)*math.Sin(t)))
		if i > 0 {
			drawLineWidth(dst, px, py, x, y, 1, c)
		}
		px, py = x, y
	}
}

func fillPolygon(dst *image.NRGBA, pts []Point, c color.NRGBA) {
	if len(pts) < 3 {
		return
	}
	minY, maxY := pts[0].Y, pts[0].Y
	for _, p := range pts {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	b := dst.Bounds()
	n := len(pts)
	for y := minY; y <= maxY; y++ {
		var nodes []int
		j := n - 1
		for i := 0; i < n; i++ {
			pi, pj := pts[i], pts[j]
			if (pi.Y < y && pj.Y >= y) || (pj.Y < y && pi.Y >= y) {
				x := pi.X + int(float64(y-pi.Y)*float64(pj.X-pi.X)/float64(pj.Y-pi.Y))
				nodes = append(nodes, x)
			}
			j = i
		}
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				if nodes[j] < nodes[i] {
					nodes[i], nodes[j] = nodes[j], nodes[i]
				}
			}
		}
		for i := 0; i+1 < len(nodes); i += 2 {
			a, b2 := nodes[i], nodes[i+1]
			if a > b2 {
				a, b2 = b2, a
			}
			for x := a; x <= b2; x++ {
				if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
					setOver(dst, x, y, c)
				}
			}
		}
	}
}

func drawLineWidth(dst *image.NRGBA, x0, y0, x1, y1, width int, c color.NRGBA) {
	dx := intcolor.AbsInt(x1 - x0)
	dy := -intcolor.AbsInt(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	x, y := x0, y0
	r := width / 2
	if r < 0 {
		r = 0
	}
	for {
		if width <= 1 {
			setOver(dst, x, y, c)
		} else {
			fillEllipse(dst, x, y, r, r, c)
		}
		if x == x1 && y == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
}

func bezierPoint(pts []Point, t float64) Point {
	n := len(pts)
	tmp := make([]Point, n)
	copy(tmp, pts)
	for k := 1; k < n; k++ {
		for i := 0; i < n-k; i++ {
			tmp[i] = Point{
				X: int(math.Round((1-t)*float64(tmp[i].X) + t*float64(tmp[i+1].X))),
				Y: int(math.Round((1-t)*float64(tmp[i].Y) + t*float64(tmp[i+1].Y))),
			}
		}
	}
	return tmp[0]
}

func floodFill(dst *image.NRGBA, x, y int, repl color.NRGBA) {
	b := dst.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return
	}
	target := dst.NRGBAAt(x, y)
	if target == repl {
		return
	}
	stack := []Point{{x, y}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if p.X < b.Min.X || p.Y < b.Min.Y || p.X >= b.Max.X || p.Y >= b.Max.Y {
			continue
		}
		if dst.NRGBAAt(p.X, p.Y) != target {
			continue
		}
		dst.SetNRGBA(p.X, p.Y, repl)
		stack = append(stack, Point{p.X + 1, p.Y}, Point{p.X - 1, p.Y}, Point{p.X, p.Y + 1}, Point{p.X, p.Y - 1})
	}
}

func setOver(dst *image.NRGBA, x, y int, c color.NRGBA) {
	b := dst.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return
	}
	dst.SetNRGBA(x, y, intcolor.OverNRGBA(dst.NRGBAAt(x, y), c))
}

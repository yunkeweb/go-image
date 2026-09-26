package goimage

import (
	"image"
	"image/color"
	"math"
)

// Drawable configures a vector primitive (rectangle, ellipse, line, …).
type Drawable struct {
	Width, Height  int
	Radius         int
	Background     any
	BorderColor    any
	BorderSize     int
	Points         []Point
	X1, Y1, X2, Y2 int
}

func (d *Drawable) Size(w, h int) *Drawable   { d.Width, d.Height = w, h; return d }
func (d *Drawable) SetWidth(w int) *Drawable  { d.Width = w; return d }
func (d *Drawable) SetHeight(h int) *Drawable { d.Height = h; return d }
func (d *Drawable) SetRadius(r int) *Drawable { d.Radius = r; return d }
func (d *Drawable) SetBackground(c any) *Drawable {
	d.Background = c
	return d
}
func (d *Drawable) SetBorder(size int, col any) *Drawable {
	d.BorderSize = size
	d.BorderColor = col
	return d
}
func (d *Drawable) Line(x1, y1, x2, y2 int) *Drawable {
	d.X1, d.Y1, d.X2, d.Y2 = x1, y1, x2, y2
	return d
}
func (d *Drawable) AddPoint(x, y int) *Drawable {
	d.Points = append(d.Points, Point{X: x, Y: y})
	return d
}

func (img *Image) DrawPixel(x, y int, col any) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		b := dst.Bounds()
		if x >= b.Min.X && y >= b.Min.Y && x < b.Max.X && y < b.Max.Y {
			dst.SetNRGBA(x, y, c.NRGBA())
		}
		return dst, nil
	})
}

func (img *Image) Fill(col any, xy ...int) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	flood := len(xy) >= 2
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		if flood {
			floodFill(dst, xy[0], xy[1], c.NRGBA())
		} else {
			fillRect(dst, dst.Bounds(), c.NRGBA())
		}
		return dst, nil
	})
}

func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		w, h := d.Width, d.Height
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		r := image.Rect(x, y, x+w, y+h)
		if d.Background != nil {
			bg, err := ParseColor(d.Background)
			if err == nil {
				fillRect(dst, r, bg.NRGBA())
			}
		}
		if d.BorderSize > 0 && d.BorderColor != nil {
			bc, err := ParseColor(d.BorderColor)
			if err == nil {
				strokeRect(dst, r, d.BorderSize, bc.NRGBA())
			}
		}
		return dst, nil
	})
}

func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	rx, ry := d.Width/2, d.Height/2
	if d.Radius > 0 {
		rx, ry = d.Radius, d.Radius
	}
	return img.drawEllipse(x, y, rx, ry, d)
}

func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	r := d.Radius
	if r == 0 {
		r = d.Width / 2
	}
	return img.drawEllipse(x, y, r, r, d)
}

func (img *Image) drawEllipse(cx, cy, rx, ry int, d *Drawable) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		if d.Background != nil {
			if bg, err := ParseColor(d.Background); err == nil {
				fillEllipse(dst, cx, cy, rx, ry, bg.NRGBA())
			}
		}
		if d.BorderSize > 0 && d.BorderColor != nil {
			if bc, err := ParseColor(d.BorderColor); err == nil {
				for i := 0; i < d.BorderSize; i++ {
					strokeEllipse(dst, cx, cy, rx-i, ry-i, bc.NRGBA())
				}
			}
		}
		return dst, nil
	})
}

func (img *Image) DrawPolygon(init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		if d.Background != nil {
			if bg, err := ParseColor(d.Background); err == nil {
				fillPolygon(dst, d.Points, bg.NRGBA())
			}
		}
		if d.BorderSize > 0 && d.BorderColor != nil && len(d.Points) > 1 {
			if bc, err := ParseColor(d.BorderColor); err == nil {
				for i := 0; i < len(d.Points); i++ {
					a := d.Points[i]
					b := d.Points[(i+1)%len(d.Points)]
					drawLineWidth(dst, a.X, a.Y, b.X, b.Y, d.BorderSize, bc.NRGBA())
				}
			}
		}
		return dst, nil
	})
}

func (img *Image) DrawLine(init func(*Drawable)) *Image {
	d := &Drawable{BorderSize: 1, BorderColor: "#000000"}
	if init != nil {
		init(d)
	}
	col := d.BorderColor
	if col == nil {
		col = d.Background
	}
	if col == nil {
		col = "#000000"
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	w := d.BorderSize
	if w < 1 {
		w = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		drawLineWidth(dst, d.X1, d.Y1, d.X2, d.Y2, w, c.NRGBA())
		return dst, nil
	})
}

func (img *Image) DrawBezier(init func(*Drawable)) *Image {
	d := &Drawable{BorderSize: 1, BorderColor: "#000000"}
	if init != nil {
		init(d)
	}
	col := d.BorderColor
	if col == nil {
		col = "#000000"
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	w := d.BorderSize
	if w < 1 {
		w = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		pts := d.Points
		if len(pts) < 2 {
			return dst, nil
		}
		steps := 64
		prev := bezierPoint(pts, 0)
		for i := 1; i <= steps; i++ {
			t := float64(i) / float64(steps)
			p := bezierPoint(pts, t)
			drawLineWidth(dst, prev.X, prev.Y, p.X, p.Y, w, c.NRGBA())
			prev = p
		}
		return dst, nil
	})
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
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
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
	dst.SetNRGBA(x, y, overNRGBA(dst.NRGBAAt(x, y), c))
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

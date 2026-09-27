package modifier

import (
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"

	"github.com/yunkeweb/go-image/internal/pool"
)

// Resample scales srcRect of src into a new dw×dh NRGBA.
func Resample(src *image.NRGBA, srcRect image.Rectangle, dw, dh int) *image.NRGBA {
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	dst := pool.Acquire(dw, dh)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, srcRect, draw.Src, nil)
	return dst
}

// SizeOf returns the pixel size of n.
func SizeOf(n *image.NRGBA) Size {
	if n == nil {
		return Size{}
	}
	b := n.Bounds()
	return Size{Width: b.Dx(), Height: b.Dy()}
}

// CoverSizes computes the crop box and target size for a cover fit.
func CoverSizes(imagesize Size, width, height int, pos Anchor, _ bool) (crop Size, resizeTo Size, err error) {
	r, err := NewResizer(width, height)
	if err != nil {
		return Size{}, Size{}, err
	}
	cropBox := Size{Width: width, Height: height}
	contained, err := NewResizer(imagesize.Width, imagesize.Height)
	if err != nil {
		return Size{}, Size{}, err
	}
	crop, err = contained.Contain(cropBox)
	if err != nil {
		return Size{}, Size{}, err
	}
	crop = crop.AlignPivotTo(imagesize, pos)
	resizeTo = r.Resize(crop)
	return crop, resizeTo, nil
}

// ApplyCover crops n to crop and resamples to resizeTo.
func ApplyCover(n *image.NRGBA, crop, resizeTo Size) *image.NRGBA {
	px, py := crop.Pivot.X, crop.Pivot.Y
	cw, ch := crop.Width, crop.Height
	b := n.Bounds()
	sr := image.Rect(b.Min.X+px, b.Min.Y+py, b.Min.X+px+cw, b.Min.Y+py+ch).Intersect(b)
	return Resample(n, sr, resizeTo.Width, resizeTo.Height)
}

// PlaceOnCanvas draws n onto a width×height canvas filled with bg.
func PlaceOnCanvas(n *image.NRGBA, width, height int, crop Size, bg color.NRGBA) *image.NRGBA {
	scaled := Resample(n, n.Bounds(), crop.Width, crop.Height)
	dst := pool.Blank(width, height, bg)
	pt := image.Pt(crop.Pivot.X, crop.Pivot.Y)
	draw.Draw(dst, scaled.Bounds().Add(pt), scaled, scaled.Bounds().Min, draw.Over)
	if scaled != n {
		pool.Release(scaled)
	}
	return dst
}

// Crop extracts width×height from n using anchor and optional offset.
func Crop(n *image.NRGBA, width, height int, anchor Anchor, bg color.NRGBA, ox, oy int) (*image.NRGBA, error) {
	if width < 1 || height < 1 {
		return nil, invalidDimensions()
	}
	orig := SizeOf(n)
	crop := Size{Width: width, Height: height}.MovePivot(anchor, 0, 0)
	crop = crop.AlignPivotTo(orig, anchor)
	px := crop.Pivot.X + ox
	py := crop.Pivot.Y + oy
	dst := pool.Blank(width, height, bg)
	srcRect := n.Bounds()
	dp := image.Pt(-px, -py)
	draw.Draw(dst, srcRect.Add(dp), n, srcRect.Min, draw.Over)
	return dst, nil
}

// ResizeCanvas changes the canvas size without scaling pixels.
func ResizeCanvas(n *image.NRGBA, width, height int, anchor Anchor, bg color.NRGBA) (*image.NRGBA, error) {
	if width < 1 || height < 1 {
		return nil, invalidDimensions()
	}
	orig := SizeOf(n)
	canvas := Size{Width: width, Height: height}
	placed := orig.AlignPivotTo(canvas, anchor)
	dst := pool.Blank(width, height, bg)
	pt := image.Pt(placed.Pivot.X, placed.Pivot.Y)
	draw.Draw(dst, n.Bounds().Add(pt), n, n.Bounds().Min, draw.Over)
	return dst, nil
}

// Trim crops uniform border pixels within tolerance.
func Trim(n *image.NRGBA, tolerance int) *image.NRGBA {
	b := n.Bounds()
	ref := averageCorners(n)
	thresh := tolerance
	if thresh < 0 {
		thresh = 0
	}
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := n.NRGBAAt(x, y)
			if colorDist(c, ref) > thresh {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x+1 > maxX {
					maxX = x + 1
				}
				if y+1 > maxY {
					maxY = y + 1
				}
			}
		}
	}
	if maxX <= minX || maxY <= minY {
		return pool.Blank(1, 1, ref)
	}
	cropped := pool.Acquire(maxX-minX, maxY-minY)
	draw.Draw(cropped, cropped.Bounds(), n, image.Pt(minX, minY), draw.Src)
	return cropped
}

func averageCorners(n *image.NRGBA) color.NRGBA {
	b := n.Bounds()
	pts := []image.Point{
		{b.Min.X, b.Min.Y},
		{b.Max.X - 1, b.Min.Y},
		{b.Min.X, b.Max.Y - 1},
		{b.Max.X - 1, b.Max.Y - 1},
	}
	var r, g, bl, a int
	for _, p := range pts {
		c := n.NRGBAAt(p.X, p.Y)
		r += int(c.R)
		g += int(c.G)
		bl += int(c.B)
		a += int(c.A)
	}
	return color.NRGBA{R: uint8(r / 4), G: uint8(g / 4), B: uint8(bl / 4), A: uint8(a / 4)}
}

func colorDist(a, b color.NRGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	if dr < 0 {
		dr = -dr
	}
	if dg < 0 {
		dg = -dg
	}
	if db < 0 {
		db = -db
	}
	return dr + dg + db
}

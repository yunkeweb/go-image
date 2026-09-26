package goimage

import (
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"
)

func resample(src *image.NRGBA, srcRect image.Rectangle, dw, dh int) *image.NRGBA {
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	dst := acquireNRGBA(dw, dh)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, srcRect, draw.Src, nil)
	return dst
}

// Resize stretches to width/height. A zero dimension keeps the original.
func (img *Image) Resize(width, height int) *Image {
	if img.fail() {
		return img
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.resize(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) ResizeDown(width, height int) *Image {
	if img.fail() {
		return img
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.resizeDown(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) Scale(width, height int) *Image {
	if img.fail() {
		return img
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.scale(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) ScaleDown(width, height int) *Image {
	if img.fail() {
		return img
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.scaleDown(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) Cover(width, height int, position ...string) *Image {
	if img.fail() {
		return img
	}
	pos := "center"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	crop, resizeTo, err := coverSizes(img.Size(), width, height, pos, false)
	if err != nil {
		return img.setErr(err)
	}
	return img.applyCover(crop, resizeTo)
}

func (img *Image) CoverDown(width, height int, position ...string) *Image {
	if img.fail() {
		return img
	}
	pos := "center"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	crop, _, err := coverSizes(img.Size(), width, height, pos, true)
	if err != nil {
		return img.setErr(err)
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	resizeTo := r.resizeDown(crop)
	return img.applyCover(crop, resizeTo)
}

func coverSizes(imagesize Size, width, height int, pos string, _ bool) (crop Size, resizeTo Size, err error) {
	r, err := newResizer(width, height)
	if err != nil {
		return Size{}, Size{}, err
	}
	cropBox := Size{Width: width, Height: height}
	contained, err := newResizer(imagesize.Width, imagesize.Height)
	if err != nil {
		return Size{}, Size{}, err
	}
	crop, err = contained.contain(cropBox)
	if err != nil {
		return Size{}, Size{}, err
	}
	crop = crop.AlignPivotTo(imagesize, pos)
	resizeTo = r.resize(crop)
	return crop, resizeTo, nil
}

func (img *Image) applyCover(crop, resizeTo Size) *Image {
	px, py := crop.Pivot.X, crop.Pivot.Y
	cw, ch := crop.Width, crop.Height
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		b := n.Bounds()
		sr := image.Rect(b.Min.X+px, b.Min.Y+py, b.Min.X+px+cw, b.Min.Y+py+ch).Intersect(b)
		return resample(n, sr, resizeTo.Width, resizeTo.Height), nil
	})
}

func (img *Image) Contain(width, height int, background any, position ...string) *Image {
	if img.fail() {
		return img
	}
	pos := "center"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorWhite
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.contain(img.Size())
	if err != nil {
		return img.setErr(err)
	}
	canvas := Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, pos)
	return img.placeOnCanvas(width, height, crop, bg)
}

func (img *Image) Pad(width, height int, background any, position ...string) *Image {
	if img.fail() {
		return img
	}
	pos := "center"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorWhite
	}
	r, err := newResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.containDown(img.Size())
	if err != nil {
		return img.setErr(err)
	}
	canvas := Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, pos)
	return img.placeOnCanvas(width, height, crop, bg)
}

func (img *Image) placeOnCanvas(width, height int, crop Size, bg Color) *Image {
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		scaled := resample(n, n.Bounds(), crop.Width, crop.Height)
		dst := newBlank(width, height, bg)
		pt := image.Pt(crop.Pivot.X, crop.Pivot.Y)
		draw.Draw(dst, scaled.Bounds().Add(pt), scaled, scaled.Bounds().Min, draw.Over)
		if scaled != n {
			releaseNRGBA(scaled)
		}
		return dst, nil
	})
}

func (img *Image) Crop(width, height, offsetX, offsetY int, background any, position ...string) *Image {
	if img.fail() {
		return img
	}
	if width == 0 && height == 0 {
		return img.setErr(wrap(ErrInvalidDimensions, "width and height cannot both be 0"))
	}
	pos := "top-left"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorWhite
	}
	orig := img.Size()
	crop := Size{Width: width, Height: height}.MovePivot(pos, 0, 0)
	crop = crop.AlignPivotTo(orig, pos)
	px := crop.Pivot.X + offsetX
	py := crop.Pivot.Y + offsetY
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := newBlank(width, height, bg)
		srcRect := n.Bounds()
		dp := image.Pt(-px, -py)
		draw.Draw(dst, srcRect.Add(dp), n, srcRect.Min, draw.Over)
		return dst, nil
	})
}

func (img *Image) ResizeCanvas(width, height int, background any, position ...string) *Image {
	if img.fail() {
		return img
	}
	pos := "center"
	if len(position) > 0 && position[0] != "" {
		pos = position[0]
	}
	if width == 0 {
		width = img.Width()
	}
	if height == 0 {
		height = img.Height()
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorWhite
	}
	orig := img.Size()
	canvas := Size{Width: width, Height: height}
	placed := orig.AlignPivotTo(canvas, pos)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := newBlank(width, height, bg)
		pt := image.Pt(placed.Pivot.X, placed.Pivot.Y)
		draw.Draw(dst, n.Bounds().Add(pt), n, n.Bounds().Min, draw.Over)
		return dst, nil
	})
}

func (img *Image) ResizeCanvasRelative(width, height int, background any, position ...string) *Image {
	if img.fail() {
		return img
	}
	return img.ResizeCanvas(img.Width()+width, img.Height()+height, background, position...)
}

func (img *Image) Trim(tolerance int) *Image {
	if img.fail() {
		return img
	}
	if img.IsAnimated() {
		return img.setErr(wrap(ErrNotSupported, "trim modifier cannot be applied to animated images"))
	}
	n := img.primary()
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
		releaseNRGBA(n)
		img.frames[0].Img = newBlank(1, 1, colorFromNRGBA(ref))
		img.resetGIFFrameLayout()
		return img
	}
	cropped := acquireNRGBA(maxX-minX, maxY-minY)
	draw.Draw(cropped, cropped.Bounds(), n, image.Pt(minX, minY), draw.Src)
	releaseNRGBA(n)
	img.frames[0].Img = cropped
	img.resetGIFFrameLayout()
	return img
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

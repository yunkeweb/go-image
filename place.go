package goimage

import (
	"image"
	"image/draw"
)

// Place overlays another image (watermark) using 9-point alignment.
// opacity is 0–100, matching PHP PlaceModifier.
func (img *Image) Place(element any, position string, offsetX, offsetY, opacity int) *Image {
	if img.fail() {
		return img
	}
	if position == "" {
		position = "top-left"
	}
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 100 {
		opacity = 100
	}
	wm := img.managerRead(element)
	if wm.Err() != nil {
		return img.setErr(wm.Err())
	}
	pos := placePosition(img.Size(), wm.Size(), position, offsetX, offsetY)
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		dst := cloneNRGBA(n)
		overlay := wm.primary()
		if overlay == nil {
			return dst, nil
		}
		if opacity >= 100 {
			draw.Draw(dst, overlay.Bounds().Add(image.Pt(pos.X, pos.Y)), overlay, overlay.Bounds().Min, draw.Over)
			return dst, nil
		}
		placeTransparent(dst, overlay, pos, opacity)
		return dst, nil
	})
}

func (img *Image) managerRead(element any) *Image {
	m := &Manager{cfg: img.cfg}
	return m.Read(element)
}

func placePosition(imageSize, wm Size, position string, ox, oy int) Point {
	img := imageSize.MovePivot(position, ox, oy)
	mark := wm.MovePivot(position, 0, 0)
	return img.RelativePositionTo(mark)
}

func placeTransparent(dst, overlay *image.NRGBA, pos Point, opacity int) {
	ob := overlay.Bounds()
	db := dst.Bounds()
	alpha := float64(opacity) / 100
	for y := ob.Min.Y; y < ob.Max.Y; y++ {
		for x := ob.Min.X; x < ob.Max.X; x++ {
			dx := pos.X + (x - ob.Min.X)
			dy := pos.Y + (y - ob.Min.Y)
			if dx < db.Min.X || dy < db.Min.Y || dx >= db.Max.X || dy >= db.Max.Y {
				continue
			}
			s := overlay.NRGBAAt(x, y)
			s.A = uint8(float64(s.A) * alpha)
			dst.SetNRGBA(dx, dy, overNRGBA(dst.NRGBAAt(dx, dy), s))
		}
	}
}

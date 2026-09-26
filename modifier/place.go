package modifier

import (
	"image"
	"image/draw"

	intcolor "github.com/yunkeweb/go-image/internal/color"
	"github.com/yunkeweb/go-image/internal/pool"
)

func Place(dst, overlay *image.NRGBA, position string, offsetX, offsetY, opacity int) *image.NRGBA {
	if overlay == nil {
		return pool.Clone(dst)
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
	pos := PlacePosition(SizeOf(dst), SizeOf(overlay), position, offsetX, offsetY)
	out := pool.Clone(dst)
	if opacity >= 100 {
		draw.Draw(out, overlay.Bounds().Add(image.Pt(pos.X, pos.Y)), overlay, overlay.Bounds().Min, draw.Over)
		return out
	}
	placeTransparent(out, overlay, pos, opacity)
	return out
}

func PlacePosition(imageSize, wm Size, position string, ox, oy int) Point {
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
			dst.SetNRGBA(dx, dy, intcolor.OverNRGBA(dst.NRGBAAt(dx, dy), s))
		}
	}
}

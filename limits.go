package goimage

import "math"

func checkInputBytes(cfg Config, n int64) error {
	max := cfg.Limits.MaxInputBytes
	if max <= 0 {
		return nil
	}
	if n > max {
		return wrap(ErrLimit, "input size %d exceeds MaxInputBytes %d", n, max)
	}
	return nil
}

func checkImageLimits(cfg Config, w, h, frames int) error {
	lim := cfg.Limits
	if frames < 1 {
		frames = 1
	}
	if lim.MaxWidth > 0 && w > lim.MaxWidth {
		return wrap(ErrLimit, "width %d exceeds MaxWidth %d", w, lim.MaxWidth)
	}
	if lim.MaxHeight > 0 && h > lim.MaxHeight {
		return wrap(ErrLimit, "height %d exceeds MaxHeight %d", h, lim.MaxHeight)
	}
	if lim.MaxFrames > 0 && frames > lim.MaxFrames {
		return wrap(ErrLimit, "frame count %d exceeds MaxFrames %d", frames, lim.MaxFrames)
	}
	if lim.MaxPixels <= 0 {
		return nil
	}
	pixels, ok := totalPixels(w, h, frames)
	if !ok || pixels > lim.MaxPixels {
		return wrap(ErrLimit, "pixel count exceeds MaxPixels %d", lim.MaxPixels)
	}
	return nil
}

func totalPixels(w, h, frames int) (int64, bool) {
	if w < 1 || h < 1 || frames < 1 {
		return 0, true
	}
	if w > 0 && int64(h) > math.MaxInt64/int64(w) {
		return 0, false
	}
	area := int64(w) * int64(h)
	if int64(frames) > 0 && area > math.MaxInt64/int64(frames) {
		return 0, false
	}
	return area * int64(frames), true
}

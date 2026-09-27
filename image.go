package goimage

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-image/encoder"
	intcolor "github.com/yunkeweb/go-image/internal/color"
	"github.com/yunkeweb/go-image/internal/errs"
	"github.com/yunkeweb/go-image/internal/pool"
	"github.com/yunkeweb/go-image/modifier"
)

var (
	ErrRuntime           = errs.ErrRuntime
	ErrDecoder           = errs.ErrDecoder
	ErrEncoder           = errs.ErrEncoder
	ErrGeometry          = errs.ErrGeometry
	ErrInvalidDimensions = errs.ErrInvalidDimensions
	ErrColor             = errs.ErrColor
	ErrInput             = errs.ErrInput
	ErrNotSupported      = errs.ErrNotSupported
	ErrNotWritable       = errs.ErrNotWritable
	ErrAnimation         = errs.ErrAnimation
	ErrFont              = errs.ErrFont
	ErrDriver            = errs.ErrDriver
)

func wrap(kind error, format string, args ...any) error {
	return errs.Wrap(kind, format, args...)
}

func cloneExif(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = cloneExifValue(v)
	}
	return dst
}

func cloneExifValue(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return append([]byte(nil), t...)
	case []string:
		return append([]string(nil), t...)
	case []int:
		return append([]int(nil), t...)
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = cloneExifValue(x)
		}
		return out
	case map[string]any:
		return cloneExif(t)
	default:
		return v
	}
}

// Point is an integer pixel coordinate in the root public API.
type Point struct {
	X, Y int
}

// Size is a rectangle with an optional 9-point pivot.
type Size struct {
	Width, Height int
	Pivot         Point
}

func (s Size) toMod() modifier.Size {
	return modifier.Size{Width: s.Width, Height: s.Height, Pivot: modifier.Point{X: s.Pivot.X, Y: s.Pivot.Y}}
}

func sizeFromMod(s modifier.Size) Size {
	return Size{Width: s.Width, Height: s.Height, Pivot: Point{X: s.Pivot.X, Y: s.Pivot.Y}}
}

func pointsToMod(pts []Point) []modifier.Point {
	out := make([]modifier.Point, len(pts))
	for i, p := range pts {
		out[i] = modifier.Point{X: p.X, Y: p.Y}
	}
	return out
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
// Untyped string literals such as "center" still compile. Unknown names
// fall through to the modifier helper; Image methods report them via Err().
func (s Size) MovePivot(position Anchor, offsetX, offsetY int) Size {
	a, err := ParseAnchor(string(position))
	if err != nil {
		a = position
	}
	return sizeFromMod(s.toMod().MovePivot(string(a), offsetX, offsetY))
}

// RelativePositionTo returns the vector from other.Pivot to s.Pivot.
func (s Size) RelativePositionTo(other Size) Point {
	p := s.toMod().RelativePositionTo(other.toMod())
	return Point{X: p.X, Y: p.Y}
}

// AlignPivotTo moves s so its named pivot matches ref's named pivot.
func (s Size) AlignPivotTo(ref Size, position Anchor) Size {
	a, err := ParseAnchor(string(position))
	if err != nil {
		a = position
	}
	return sizeFromMod(s.toMod().AlignPivotTo(ref.toMod(), string(a)))
}

// Color is an 8-bit-per-channel sRGB color with alpha (255 = opaque).
type Color struct {
	R, G, B, A uint8
}

// ColorspaceName names a working colorspace. Pixel buffers stay NRGBA.
type ColorspaceName string

const (
	ColorspaceRGB  ColorspaceName = "rgb"
	ColorspaceCMYK ColorspaceName = "cmyk"
)

var (
	ColorTransparent = Color{R: 255, G: 255, B: 255, A: 0}
	ColorWhite       = Color{R: 255, G: 255, B: 255, A: 255}
	ColorBlack       = Color{R: 0, G: 0, B: 0, A: 255}
)

// ParseColor decodes v as an sRGB color. Strings may be CSS names, #hex,
// rgb()/rgba(). color.Color and Color values are accepted. Nil *Color is an error.
func ParseColor(v any) (Color, error) {
	switch c := v.(type) {
	case Color:
		return c, nil
	case *Color:
		if c == nil {
			return Color{}, wrap(ErrColor, "nil color")
		}
		return *c, nil
	case color.Color:
		r, g, b, a := c.RGBA()
		return Color{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}, nil
	case string:
		return parseColorString(c)
	default:
		return Color{}, wrap(ErrColor, "unable to decode color from %T", v)
	}
}

func parseBackground(v any) (Color, error) {
	if v == nil {
		return ColorWhite, nil
	}
	return ParseColor(v)
}

func parseColorString(s string) (Color, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Color{}, wrap(ErrColor, "empty color string")
	}
	lower := strings.ToLower(s)
	if lower == "transparent" {
		return ColorTransparent, nil
	}
	if hex, ok := intcolor.HTMLNames[lower]; ok {
		return parseHex(hex)
	}
	if strings.HasPrefix(lower, "rgb") || strings.HasPrefix(lower, "srgb") {
		return parseRGBFunc(s)
	}
	return parseHex(s)
}

func parseHex(s string) (Color, error) {
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimSpace(s)
	if s == "" {
		return Color{}, wrap(ErrColor, "unable to decode hex color")
	}
	expand := func(ch byte) uint8 {
		v, err := strconv.ParseUint(string([]byte{ch, ch}), 16, 8)
		if err != nil {
			return 0
		}
		return uint8(v)
	}
	pair := func(i int) (uint8, error) {
		v, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return 0, wrap(ErrColor, "unable to decode hex color")
		}
		return uint8(v), nil
	}
	c := Color{A: 255}
	switch len(s) {
	case 3:
		c.R, c.G, c.B = expand(s[0]), expand(s[1]), expand(s[2])
	case 4:
		c.R, c.G, c.B, c.A = expand(s[0]), expand(s[1]), expand(s[2]), expand(s[3])
	case 6:
		var err error
		if c.R, err = pair(0); err != nil {
			return Color{}, err
		}
		if c.G, err = pair(2); err != nil {
			return Color{}, err
		}
		if c.B, err = pair(4); err != nil {
			return Color{}, err
		}
	case 8:
		var err error
		if c.R, err = pair(0); err != nil {
			return Color{}, err
		}
		if c.G, err = pair(2); err != nil {
			return Color{}, err
		}
		if c.B, err = pair(4); err != nil {
			return Color{}, err
		}
		if c.A, err = pair(6); err != nil {
			return Color{}, err
		}
	default:
		return Color{}, wrap(ErrColor, "unable to decode hex color %q", s)
	}
	return c, nil
}

func parseRGBFunc(s string) (Color, error) {
	start := strings.Index(s, "(")
	end := strings.LastIndex(s, ")")
	if start < 0 || end <= start {
		return Color{}, wrap(ErrColor, "unable to decode rgb color")
	}
	parts := strings.Split(s[start+1:end], ",")
	if len(parts) < 3 || len(parts) > 4 {
		return Color{}, wrap(ErrColor, "unable to decode rgb color")
	}
	ch := func(p string) (uint8, error) {
		p = strings.TrimSpace(p)
		if strings.HasSuffix(p, "%") {
			f, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
			if err != nil {
				return 0, wrap(ErrColor, "unable to decode rgb color")
			}
			return clamp8(int(math.Round(f / 100 * 255))), nil
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return 0, wrap(ErrColor, "unable to decode rgb color")
		}
		return clamp8(int(math.Round(f))), nil
	}
	r, err := ch(parts[0])
	if err != nil {
		return Color{}, err
	}
	g, err := ch(parts[1])
	if err != nil {
		return Color{}, err
	}
	b, err := ch(parts[2])
	if err != nil {
		return Color{}, err
	}
	c := Color{R: r, G: g, B: b, A: 255}
	if len(parts) == 4 {
		p := strings.TrimSpace(parts[3])
		if strings.HasSuffix(p, "%") {
			f, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
			if err != nil {
				return Color{}, wrap(ErrColor, "unable to decode rgb color")
			}
			c.A = clamp8(int(math.Round(f / 100 * 255)))
		} else {
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return Color{}, wrap(ErrColor, "unable to decode rgb color")
			}
			if f <= 1 {
				c.A = clamp8(int(math.Round(f * 255)))
			} else {
				c.A = clamp8(int(math.Round(f)))
			}
		}
	}
	return c, nil
}

func clamp8(v int) uint8 { return intcolor.Clamp8(v) }
func clampInt(v, lo, hi int) int { return intcolor.ClampInt(v, lo, hi) }

// NRGBA converts c to color.NRGBA.
func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// RGBA implements color.Color.
func (c Color) RGBA() (r, g, b, a uint32) {
	return c.NRGBA().RGBA()
}

func colorFromNRGBA(n color.NRGBA) Color {
	return Color{R: n.R, G: n.G, B: n.B, A: n.A}
}

// IsTransparent reports whether alpha is less than 255.
func (c Color) IsTransparent() bool { return c.A < 255 }

// IsClear reports whether alpha is 0.
func (c Color) IsClear() bool { return c.A == 0 }

// IsGreyscale reports whether R, G, and B are equal.
func (c Color) IsGreyscale() bool { return c.R == c.G && c.G == c.B }

// ToHex formats c as hex with an optional prefix such as "#".
func (c Color) ToHex(prefix string) string {
	if c.IsTransparent() {
		return fmt.Sprintf("%s%02x%02x%02x%02x", prefix, c.R, c.G, c.B, c.A)
	}
	return fmt.Sprintf("%s%02x%02x%02x", prefix, c.R, c.G, c.B)
}

// String returns an rgb() or rgba() CSS color.
func (c Color) String() string {
	if c.IsTransparent() {
		return fmt.Sprintf("rgba(%d, %d, %d, %.1f)", c.R, c.G, c.B, float64(c.A)/255)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

// HSL returns hue in degrees and saturation/lightness in 0..1.
func (c Color) HSL() (h, s, l float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2
	if max == min {
		return 0, 0, l
	}
	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	return h, s, l
}

// HSV returns hue in degrees and saturation/value in 0..1.
func (c Color) HSV() (h, s, v float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	v = max
	d := max - min
	if max == 0 {
		return 0, 0, v
	}
	s = d / max
	if d == 0 {
		return 0, s, v
	}
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	return h, s, v
}

// CMYK returns cyan, magenta, yellow, and key in 0..1.
func (c Color) CMYK() (cyan, magenta, yellow, key float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	key = 1 - math.Max(r, math.Max(g, b))
	if key >= 1 {
		return 0, 0, 0, 1
	}
	cyan = (1 - r - key) / (1 - key)
	magenta = (1 - g - key) / (1 - key)
	yellow = (1 - b - key) / (1 - key)
	return cyan, magenta, yellow, key
}

// ColorFromHSL builds an sRGB Color from HSL. h is degrees; s and l are 0..1.
func ColorFromHSL(h, s, l float64, a uint8) Color {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = intcolor.Clamp01(s)
	l = intcolor.Clamp01(l)
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return Color{
		R: clamp8(int(math.Round((r + m) * 255))),
		G: clamp8(int(math.Round((g + m) * 255))),
		B: clamp8(int(math.Round((b + m) * 255))),
		A: a,
	}
}

// Format names an encode/decode image format such as jpeg or png.
type Format string

const (
	FormatJPEG Format = Format(encoder.JPEG)
	FormatPNG  Format = Format(encoder.PNG)
	FormatGIF  Format = Format(encoder.GIF)
	FormatWEBP Format = Format(encoder.WEBP)
	FormatBMP  Format = Format(encoder.BMP)
	FormatTIFF Format = Format(encoder.TIFF)
	FormatAVIF Format = Format(encoder.AVIF)
	FormatHEIC Format = Format(encoder.HEIC)
	FormatJP2  Format = Format(encoder.JP2)
)

// MediaType returns the MIME type for f, or application/octet-stream.
func (f Format) MediaType() string { return encoder.MediaType(string(f)) }

// FileExtension returns the preferred file extension without a dot.
func (f Format) FileExtension() string { return encoder.FileExtension(string(f)) }
func (f Format) supportedEncode() bool { return encoder.SupportedEncode(string(f)) }

func parseFormat(identifier string) (Format, error) {
	s, err := encoder.ParseFormat(identifier)
	return Format(s), err
}

func formatFromPath(path string) (Format, error) {
	s, err := encoder.FormatFromPath(path)
	return Format(s), err
}

// Origin records where an Image was decoded from.
type Origin struct {
	MediaType string
	FilePath  string
}

// MimeType returns Origin.MediaType.
func (o Origin) MimeType() string { return o.MediaType }

// FileExtension returns the path extension without a leading dot.
func (o Origin) FileExtension() string {
	if o.FilePath == "" {
		return ""
	}
	ext := filepath.Ext(o.FilePath)
	if ext == "" {
		return ""
	}
	return ext[1:]
}

// Frame is one animation frame on the logical canvas.
// Pixel buffers are unexported; use Image for an independent copy.
type Frame struct {
	img        *image.NRGBA
	Delay      float64
	Dispose    int
	OffsetLeft int
	OffsetTop  int
}

// Size returns the pixel size of the frame buffer.
func (f Frame) Size() Size {
	if f.img == nil {
		return Size{}
	}
	b := f.img.Bounds()
	return Size{Width: b.Dx(), Height: b.Dy()}
}

// Image returns an independent copy of the frame pixels.
// Mutating the returned image does not change the parent Image.
func (f Frame) Image() image.Image {
	if f.img == nil {
		return nil
	}
	return pool.Clone(f.img)
}

func (f Frame) clone() Frame {
	out := f
	if f.img != nil {
		out.img = pool.Clone(f.img)
	}
	return out
}

func (f Frame) asAnim() modifier.AnimFrame {
	return modifier.AnimFrame{
		Img:        f.img,
		Delay:      f.Delay,
		Dispose:    f.Dispose,
		OffsetLeft: f.OffsetLeft,
		OffsetTop:  f.OffsetTop,
	}
}

func frameFromAnim(f modifier.AnimFrame) Frame {
	return Frame{
		img:        f.Img,
		Delay:      f.Delay,
		Dispose:    f.Dispose,
		OffsetLeft: f.OffsetLeft,
		OffsetTop:  f.OffsetTop,
	}
}

// Image is a fluent, library-owned raster that may hold one or more frames.
// The first failure in a chain is stored and retrieved with Err(). An Image
// is not safe for concurrent mutation; clone first when sharing work.
//
// Image implements image.Image using the first frame as the static view.
// Error handling: Check img.Err() after call chain.
type Image struct {
	frames     []Frame
	loops      int
	origin     Origin
	exif       map[string]any
	cfg        Config
	colorspace ColorspaceName
	resX       float64
	resY       float64
	profile    []byte
	err        error
}

var _ image.Image = (*Image)(nil)

// ColorModel returns color.NRGBAModel for every Image, including nil and failed values.
func (img *Image) ColorModel() color.Model { return color.NRGBAModel }

// Bounds returns the first frame's bounds. Nil and failed images return an empty rectangle.
func (img *Image) Bounds() image.Rectangle {
	if img == nil || img.fail() || img.primary() == nil {
		return image.Rectangle{}
	}
	return img.primary().Bounds()
}

// At returns the first-frame color at (x, y). Nil, failed, and out-of-bounds
// coordinates return a zero NRGBA.
func (img *Image) At(x, y int) color.Color {
	if img == nil || img.fail() || img.primary() == nil {
		return color.NRGBA{}
	}
	return img.primary().At(x, y)
}

func newImage(frames []Frame, cfg Config) *Image {
	return &Image{
		frames:     frames,
		cfg:        cfg,
		colorspace: ColorspaceRGB,
		resX:       72,
		resY:       72,
		exif:       map[string]any{},
	}
}

func failed(err error) *Image {
	return &Image{err: err, exif: map[string]any{}, cfg: DefaultConfig()}
}

func (img *Image) fail() bool {
	return img == nil || img.err != nil
}

func (img *Image) setErr(err error) *Image {
	if img == nil {
		return failed(err)
	}
	if img.err == nil && err != nil {
		img.err = err
	}
	return img
}

// Err returns the first delayed error, or a runtime error when img is nil.
func (img *Image) Err() error {
	if img == nil {
		return wrap(ErrRuntime, "nil image")
	}
	return img.err
}

func (img *Image) eachFrame(fn func(*Frame) error) *Image {
	if img.fail() {
		return img
	}
	for i := range img.frames {
		if err := fn(&img.frames[i]); err != nil {
			return img.setErr(err)
		}
	}
	return img
}

func (img *Image) primary() *image.NRGBA {
	if img == nil || len(img.frames) == 0 {
		return nil
	}
	return img.frames[0].img
}

// Clone returns an independent copy of img, including all frame pixel buffers
// and mutable metadata. EXIF []byte, slices, maps, and nested map[string]any
// values are deep-copied. Unknown pointer types inside EXIF are not deep-copied.
// The returned image can be modified without affecting img.
func (img *Image) Clone() *Image {
	if img == nil {
		return failed(wrap(ErrRuntime, "nil image"))
	}
	cp := *img
	cp.frames = make([]Frame, len(img.frames))
	for i := range img.frames {
		cp.frames[i] = img.frames[i].clone()
	}
	cp.exif = cloneExif(img.exif)
	if img.profile != nil {
		cp.profile = append([]byte(nil), img.profile...)
	}
	return &cp
}

// Native returns an independent copy of the first frame as image.Image.
// Mutating the result does not change img. Failed images return a 1×1 canvas.
func (img *Image) Native() image.Image {
	if img.fail() {
		n := pool.Acquire(1, 1)
		if n == nil {
			return image.NewNRGBA(image.Rect(0, 0, 1, 1))
		}
		return n
	}
	if n := pool.Clone(img.primary()); n != nil {
		return n
	}
	return image.NewNRGBA(image.Rect(0, 0, 1, 1))
}

// UnsafeNative returns the internal first-frame NRGBA without copying.
// Callers must not mutate it or retain it across further Image methods.
func (img *Image) UnsafeNative() *image.NRGBA {
	if img.fail() {
		return nil
	}
	return img.primary()
}

// Frames returns a deep copy of every animation frame. Mutating the slice or
// any frame buffer does not change img.
func (img *Image) Frames() []Frame {
	if img.fail() {
		return nil
	}
	out := make([]Frame, len(img.frames))
	for i := range img.frames {
		out[i] = img.frames[i].clone()
	}
	return out
}

// UnsafeFrames returns the internal frame slice without copying.
// Callers must not mutate it or retain it across further Image methods.
func (img *Image) UnsafeFrames() []Frame {
	if img.fail() {
		return nil
	}
	return img.frames
}

// Origin returns the decode source of img.
func (img *Image) Origin() Origin {
	if img == nil {
		return Origin{}
	}
	return img.origin
}

// SetOrigin replaces the recorded decode source.
func (img *Image) SetOrigin(o Origin) *Image {
	if img.fail() {
		return img
	}
	img.origin = o
	return img
}

// Count returns the number of animation frames, or 0 on a failed image.
func (img *Image) Count() int {
	if img.fail() {
		return 0
	}
	return len(img.frames)
}

// IsAnimated reports whether img has more than one frame.
func (img *Image) IsAnimated() bool { return img.Count() > 1 }

// Loops returns the GIF Netscape loop count.
func (img *Image) Loops() int {
	if img.fail() {
		return 0
	}
	return img.loops
}

// SetLoops stores the GIF Netscape loop count (0 means loop forever).
func (img *Image) SetLoops(n int) *Image {
	if img.fail() {
		return img
	}
	img.loops = n
	return img
}

// Width returns the first frame's width in pixels.
func (img *Image) Width() int {
	if img.fail() || img.primary() == nil {
		return 0
	}
	return img.primary().Bounds().Dx()
}

// Height returns the first frame's height in pixels.
func (img *Image) Height() int {
	if img.fail() || img.primary() == nil {
		return 0
	}
	return img.primary().Bounds().Dy()
}

// Size returns the first frame's width and height.
func (img *Image) Size() Size {
	return Size{Width: img.Width(), Height: img.Height()}
}

// Colorspace returns the working colorspace name.
func (img *Image) Colorspace() ColorspaceName {
	if img.fail() {
		return ColorspaceRGB
	}
	return img.colorspace
}

// SetColorspace records a colorspace name (rgb or cmyk). Pixel data stay NRGBA.
func (img *Image) SetColorspace(name string) *Image {
	if img.fail() {
		return img
	}
	switch strings.ToLower(name) {
	case "cmyk":
		img.colorspace = ColorspaceCMYK
	default:
		img.colorspace = ColorspaceRGB
	}
	return img
}

// Resolution returns stored dots-per-inch values.
func (img *Image) Resolution() (x, y float64) {
	if img.fail() {
		return 0, 0
	}
	return img.resX, img.resY
}

// SetResolution stores dots-per-inch metadata.
func (img *Image) SetResolution(x, y float64) *Image {
	if img.fail() {
		return img
	}
	img.resX, img.resY = x, y
	return img
}

// PickColor returns the first-frame color at (x, y), or a zero Color out of bounds.
func (img *Image) PickColor(x, y int) Color {
	return img.PickColorFrame(x, y, 0)
}

// PickColorFrame returns the color at (x, y) on frame, or a zero Color out of range.
func (img *Image) PickColorFrame(x, y, frame int) Color {
	if img.fail() || frame < 0 || frame >= len(img.frames) || img.frames[frame].img == nil {
		return Color{}
	}
	n := img.frames[frame].img
	b := n.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return Color{}
	}
	return colorFromNRGBA(n.NRGBAAt(x, y))
}

// PickColors returns the color at (x, y) on every frame.
func (img *Image) PickColors(x, y int) []Color {
	if img.fail() {
		return nil
	}
	out := make([]Color, len(img.frames))
	for i := range img.frames {
		out[i] = img.PickColorFrame(x, y, i)
	}
	return out
}

// Exif returns a deep copy of the EXIF map. Mutating the result does not
// change img. Nested []byte, slices, and map[string]any values are copied;
// unknown pointer types are not deep-copied.
func (img *Image) Exif() map[string]any {
	if img.fail() {
		return nil
	}
	return cloneExif(img.exif)
}

// UnsafeExif returns the internal EXIF map without copying.
func (img *Image) UnsafeExif() map[string]any {
	if img.fail() {
		return nil
	}
	return img.exif
}

// ExifQuery returns a copied value for key, IFD0.key, or EXIF.key.
func (img *Image) ExifQuery(key string) any {
	if img.fail() || img.exif == nil {
		return nil
	}
	if v, ok := img.exif[key]; ok {
		return cloneExifValue(v)
	}
	if v, ok := img.exif["IFD0."+key]; ok {
		return cloneExifValue(v)
	}
	return cloneExifValue(img.exif["EXIF."+key])
}

// SetExif replaces EXIF with a deep copy of data.
func (img *Image) SetExif(data map[string]any) *Image {
	if img.fail() {
		return img
	}
	img.exif = cloneExif(data)
	return img
}

// BlendingColor returns the flatten-to color used by JPEG/BMP encoding.
func (img *Image) BlendingColor() Color {
	if img == nil {
		return ColorWhite
	}
	c, err := ParseColor(img.cfg.BlendingColor)
	if err != nil {
		return ColorWhite
	}
	return c
}

// SetBlendingColor stores the flatten-to color used by JPEG/BMP encoding.
func (img *Image) SetBlendingColor(v any) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(v)
	if err != nil {
		return img.setErr(err)
	}
	img.cfg.BlendingColor = c
	return img
}

// Profile returns a copy of the ICC profile bytes.
func (img *Image) Profile() []byte {
	if img.fail() {
		return nil
	}
	if img.profile == nil {
		return nil
	}
	return append([]byte(nil), img.profile...)
}

// UnsafeProfile returns the internal ICC profile slice without copying.
func (img *Image) UnsafeProfile() []byte {
	if img.fail() {
		return nil
	}
	return img.profile
}

// SetProfile stores a copy of data as the ICC profile.
func (img *Image) SetProfile(data []byte) *Image {
	if img.fail() {
		return img
	}
	img.profile = append([]byte(nil), data...)
	return img
}

// RemoveProfile drops the stored ICC profile.
func (img *Image) RemoveProfile() *Image {
	if img.fail() {
		return img
	}
	img.profile = nil
	return img
}

// Config returns a copy of the decode/encode settings used to build img.
func (img *Image) Config() Config {
	if img == nil {
		return DefaultConfig()
	}
	return img.cfg
}

func (img *Image) replaceAll(fn func(*image.NRGBA) (*image.NRGBA, error)) *Image {
	return img.eachFrame(func(f *Frame) error {
		old := f.img
		if old == nil {
			return wrap(ErrInvalidDimensions, "invalid dimensions")
		}
		n, err := fn(old)
		if err != nil {
			return err
		}
		if n == nil {
			return wrap(ErrInvalidDimensions, "invalid dimensions")
		}
		if n != old {
			pool.Release(old)
		}
		f.img = n
		return nil
	})
}

func (img *Image) replaceAllGeometry(fn func(*image.NRGBA) (*image.NRGBA, error)) *Image {
	img = img.replaceAll(fn)
	if img.fail() {
		return img
	}
	img.resetGIFFrameLayout()
	return img
}

func (img *Image) resetGIFFrameLayout() {
	if img == nil {
		return
	}
	raw := make([]modifier.AnimFrame, len(img.frames))
	for i, f := range img.frames {
		raw[i] = f.asAnim()
	}
	modifier.ResetGIFLayout(raw, img.Width(), img.Height())
	for i, f := range raw {
		img.frames[i] = frameFromAnim(f)
	}
}

func (img *Image) releaseFrames(except int) {
	if img == nil {
		return
	}
	for i := range img.frames {
		if i == except {
			continue
		}
		pool.Release(img.frames[i].img)
		img.frames[i].img = nil
	}
}

func applyGeometry(base geometrySettings, opts []GeometryOption) (geometrySettings, Anchor, Color, error) {
	cfg := applyGeometryOptions(base, opts)
	fallback := AnchorCenter
	if strings.TrimSpace(base.anchor) != "" {
		if a, err := ParseAnchor(base.anchor); err == nil {
			fallback = a
		}
	}
	a, err := resolveAnchor(cfg.anchor, fallback)
	if err != nil {
		return cfg, "", Color{}, err
	}
	bg, err := parseBackground(cfg.background)
	if err != nil {
		return cfg, a, Color{}, err
	}
	return cfg, a, bg, nil
}

// Resize scales every frame. Omitting height keeps the original aspect ratio.
// Zero and negative sizes set ErrInvalidDimensions.
// Error handling: Check img.Err() after call chain.
func (img *Image) Resize(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.Resize(img.Size().toMod())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

// ResizeDown scales down only; images already smaller than the target are unchanged.
func (img *Image) ResizeDown(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.ResizeDown(img.Size().toMod())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

// Scale resizes while fitting inside the box, keeping aspect ratio.
func (img *Image) Scale(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.Scale(img.Size().toMod())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

// ScaleDown is Scale that never enlarges the image.
func (img *Image) ScaleDown(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.ScaleDown(img.Size().toMod())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

// Cover fills width×height and crops overflow around the anchor (default center).
// Error handling: Check img.Err() after call chain.
func (img *Image) Cover(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	_, anchor, _, err := applyGeometry(geometrySettings{anchor: string(AnchorCenter)}, opts)
	if err != nil {
		return img.setErr(err)
	}
	crop, resizeTo, err := modifier.CoverSizes(img.Size().toMod(), width, height, string(anchor), false)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.ApplyCover(n, crop, resizeTo), nil
	})
}

// CoverDown is Cover that never enlarges the image.
func (img *Image) CoverDown(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	_, anchor, _, err := applyGeometry(geometrySettings{anchor: string(AnchorCenter)}, opts)
	if err != nil {
		return img.setErr(err)
	}
	crop, _, err := modifier.CoverSizes(img.Size().toMod(), width, height, string(anchor), true)
	if err != nil {
		return img.setErr(err)
	}
	r, err := modifier.NewResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	resizeTo := r.ResizeDown(crop)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.ApplyCover(n, crop, resizeTo), nil
	})
}

// Fit is an alias of Cover.
//
// Deprecated: use Cover.
func (img *Image) Fit(width, height int, opts ...GeometryOption) *Image {
	return img.Cover(width, height, opts...)
}

// Contain fits the image inside width×height and pads the remainder.
func (img *Image) Contain(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	_, anchor, bg, err := applyGeometry(geometrySettings{anchor: string(AnchorCenter), background: "ffffff"}, opts)
	if err != nil {
		return img.setErr(err)
	}
	r, err := modifier.NewResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.Contain(img.Size().toMod())
	if err != nil {
		return img.setErr(err)
	}
	canvas := modifier.Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, string(anchor))
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.PlaceOnCanvas(n, width, height, crop, bg.NRGBA()), nil
	})
}

// Pad is Contain that never enlarges the image.
func (img *Image) Pad(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	_, anchor, bg, err := applyGeometry(geometrySettings{anchor: string(AnchorCenter), background: "ffffff"}, opts)
	if err != nil {
		return img.setErr(err)
	}
	r, err := modifier.NewResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.ContainDown(img.Size().toMod())
	if err != nil {
		return img.setErr(err)
	}
	canvas := modifier.Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, string(anchor))
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.PlaceOnCanvas(n, width, height, crop, bg.NRGBA()), nil
	})
}

// Crop extracts width×height from the anchor (default top-left).
func (img *Image) Crop(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg, anchor, bg, err := applyGeometry(geometrySettings{anchor: string(AnchorTopLeft), background: "ffffff"}, opts)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Crop(n, width, height, string(anchor), bg.NRGBA(), cfg.offsetX, cfg.offsetY)
	})
}

// ResizeCanvas changes the canvas size without scaling pixels.
func (img *Image) ResizeCanvas(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	_, anchor, bg, err := applyGeometry(geometrySettings{anchor: string(AnchorCenter), background: "ffffff"}, opts)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.ResizeCanvas(n, width, height, string(anchor), bg.NRGBA())
	})
}

// ResizeCanvasRelative adds width and height to the current canvas size.
func (img *Image) ResizeCanvasRelative(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	return img.ResizeCanvas(img.Width()+width, img.Height()+height, opts...)
}

// Trim crops uniform border pixels. Animated images set ErrNotSupported.
func (img *Image) Trim(tolerance int) *Image {
	if img.fail() {
		return img
	}
	if img.IsAnimated() {
		return img.setErr(wrap(ErrNotSupported, "trim modifier cannot be applied to animated images"))
	}
	n := img.primary()
	cropped := modifier.Trim(n, tolerance)
	if cropped != n {
		pool.Release(n)
	}
	img.frames[0].img = cropped
	img.resetGIFFrameLayout()
	return img
}

// Greyscale converts every frame to luma, keeping alpha.
func (img *Image) Greyscale() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Greyscale(n), nil
	})
}

// Invert inverts RGB channels and keeps alpha.
func (img *Image) Invert() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Invert(n), nil
	})
}

// Brightness adjusts luma by level percent (-100..100).
func (img *Image) Brightness(level int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Brightness(n, level), nil
	})
}

// Contrast adjusts contrast by level percent.
func (img *Image) Contrast(level int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Contrast(n, level), nil
	})
}

// Gamma applies a gamma curve. Non-positive gamma sets ErrInput.
func (img *Image) Gamma(gamma float64) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Gamma(n, gamma)
	})
}

// Colorize tints RGB channels by signed percent deltas.
func (img *Image) Colorize(red, green, blue int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Colorize(n, red, green, blue), nil
	})
}

// Pixelate mosaics every frame with square cells of the given size.
func (img *Image) Pixelate(size int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Pixelate(n, size), nil
	})
}

// Blur applies a box blur of the given radius.
func (img *Image) Blur(amount int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Blur(n, amount), nil
	})
}

// Sharpen applies an unsharp-mask of the given amount.
func (img *Image) Sharpen(amount int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Sharpen(n, amount), nil
	})
}

// BlendTransparency composites every frame over col (or the blending color).
func (img *Image) BlendTransparency(col any) *Image {
	if img.fail() {
		return img
	}
	var bg Color
	var err error
	if col == nil {
		bg = img.BlendingColor()
	} else {
		bg, err = ParseColor(col)
		if err != nil {
			return img.setErr(err)
		}
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.BlendTransparency(n, bg.NRGBA()), nil
	})
}

// ReduceColors quantizes every frame to at most limit palette entries.
func (img *Image) ReduceColors(limit int, background any) *Image {
	if img.fail() {
		return img
	}
	if background != nil {
		if _, err := ParseColor(background); err != nil {
			return img.setErr(err)
		}
	}
	if limit < 1 {
		limit = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		pal := encoder.MedianCutPalette(n, limit)
		return modifier.ReduceColors(n, pal), nil
	})
}

// RemoveAnimation keeps one frame (index or "0%".."100%") and drops the rest.
func (img *Image) RemoveAnimation(position any) *Image {
	if img.fail() {
		return img
	}
	idx, err := parsePercentOrIndex(position, len(img.frames))
	if err != nil {
		return img.setErr(err)
	}
	if len(img.frames) == 0 {
		return img
	}
	keep := img.frames[idx].clone()
	img.releaseFrames(-1)
	img.frames = []Frame{keep}
	img.loops = 0
	return img
}

// SliceAnimation keeps length frames starting at offset.
func (img *Image) SliceAnimation(offset int, length int) *Image {
	if img.fail() {
		return img
	}
	n := len(img.frames)
	if offset < 0 {
		offset = 0
	}
	if offset >= n {
		return img.setErr(wrap(ErrAnimation, "slice offset out of range"))
	}
	end := n
	if length > 0 {
		end = min(n, offset+length)
	}
	cp := make([]Frame, 0, end-offset)
	for i := offset; i < end; i++ {
		cp = append(cp, img.frames[i].clone())
	}
	img.releaseFrames(-1)
	img.frames = cp
	return img
}

// Flip mirrors every frame vertically.
func (img *Image) Flip() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Flip(n), nil
	})
}

// Flop mirrors every frame horizontally.
func (img *Image) Flop() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Flop(n), nil
	})
}

// Rotate turns every frame by angle degrees around its center.
// Invalid background colors set Err() even when angle normalizes to 0.
func (img *Image) Rotate(angle float64, background any) *Image {
	if img.fail() {
		return img
	}
	bg, err := ParseColor(background)
	if err != nil {
		return img.setErr(err)
	}
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	if angle == 0 {
		return img
	}
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Rotate(n, angle, bg.NRGBA()), nil
	})
}

// Orient applies EXIF orientation 1..8 to pixel data, then deletes
// Orientation, IFD0.Orientation, and EXIF.Orientation so viewers do not
// rotate the image a second time. Integer and floating EXIF values are accepted.
func (img *Image) Orient() *Image {
	if img.fail() {
		return img
	}
	orient := parseOrientation(img.ExifQuery("Orientation"))
	applyOrientation(img, orient)
	img.clearOrientation()
	return img
}

// Orientate is an alias of Orient.
//
// Deprecated: use Orient.
func (img *Image) Orientate() *Image { return img.Orient() }

func (img *Image) clearOrientation() {
	if img == nil || img.exif == nil {
		return
	}
	delete(img.exif, "Orientation")
	delete(img.exif, "IFD0.Orientation")
	delete(img.exif, "EXIF.Orientation")
}

func parseOrientation(v any) int {
	n := 0
	switch t := v.(type) {
	case int:
		n = t
	case int8:
		n = int(t)
	case int16:
		n = int(t)
	case int32:
		n = int(t)
	case int64:
		n = int(t)
	case uint:
		n = int(t)
	case uint8:
		n = int(t)
	case uint16:
		n = int(t)
	case uint32:
		n = int(t)
	case uint64:
		n = int(t)
	case float32:
		n = int(t)
	case float64:
		n = int(t)
	default:
		return 1
	}
	if n < 1 || n > 8 {
		return 1
	}
	return n
}

func applyOrientation(img *Image, orient int) {
	if img == nil || orient <= 1 {
		return
	}
	img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		out := modifier.OrientPixels(n, orient)
		return out, nil
	})
}

// Place overlays src onto every frame at position.
// Offset and opacity default to (0, 0) and 100; pass WithOffset and WithOpacity
// to change them. Unknown anchors and opacity outside 0–100 set Err().
// Error handling: Check img.Err() after call chain.
func (img *Image) Place(src image.Image, position Anchor, options ...PlaceOption) *Image {
	if img.fail() {
		return img
	}
	if src == nil {
		return img.setErr(wrap(ErrInput, "nil watermark"))
	}
	anchor, err := ParseAnchor(string(position))
	if err != nil {
		return img.setErr(err)
	}
	cfg := applyGeometryOptions(geometrySettings{opacity: 100}, options)
	if cfg.opacity < 0 || cfg.opacity > 100 {
		return img.setErr(wrap(ErrInput, "opacity must be 0–100, got %d", cfg.opacity))
	}
	overlay, owned, err := overlayNRGBA(src)
	if err != nil {
		return img.setErr(err)
	}
	if owned {
		defer pool.Release(overlay)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Place(n, overlay, string(anchor), cfg.offsetX, cfg.offsetY, cfg.opacity), nil
	})
}

func overlayNRGBA(src image.Image) (n *image.NRGBA, owned bool, err error) {
	if src == nil {
		return nil, false, wrap(ErrInput, "nil watermark")
	}
	if img, ok := src.(*Image); ok {
		if img.fail() {
			return nil, false, img.Err()
		}
		out := img.primary()
		if out == nil {
			return nil, false, wrap(ErrInput, "nil watermark")
		}
		return out, false, nil
	}
	out := pool.AsNRGBA(src)
	if out == nil {
		return nil, false, wrap(ErrInvalidDimensions, "invalid dimensions")
	}
	return out, true, nil
}

// Drawable holds style for DrawRectangle, DrawEllipse, DrawCircle, DrawPolygon,
// DrawLine, and DrawBezier. It is not safe for concurrent use.
type Drawable struct {
	Width, Height  int
	Radius         int
	Background     any
	BorderColor    any
	BorderSize     int
	Points         []Point
	X1, Y1, X2, Y2 int
}

// Size sets the drawable width and height.
func (d *Drawable) Size(w, h int) *Drawable { d.Width, d.Height = w, h; return d }
// SetWidth sets the drawable width.
func (d *Drawable) SetWidth(w int) *Drawable { d.Width = w; return d }

// SetHeight sets the drawable height.
func (d *Drawable) SetHeight(h int) *Drawable { d.Height = h; return d }

// SetRadius sets the circle/ellipse radius.
func (d *Drawable) SetRadius(r int) *Drawable { d.Radius = r; return d }

// SetBackground sets the fill color.
func (d *Drawable) SetBackground(c any) *Drawable {
	d.Background = c
	return d
}

// SetBorder sets the stroke width and color.
func (d *Drawable) SetBorder(size int, col any) *Drawable {
	d.BorderSize = size
	d.BorderColor = col
	return d
}

// Line sets the endpoints for DrawLine.
func (d *Drawable) Line(x1, y1, x2, y2 int) *Drawable {
	d.X1, d.Y1, d.X2, d.Y2 = x1, y1, x2, y2
	return d
}

// AddPoint appends a vertex for DrawPolygon or DrawBezier.
func (d *Drawable) AddPoint(x, y int) *Drawable {
	d.Points = append(d.Points, Point{X: x, Y: y})
	return d
}

func nrgbaPtr(v any) (*color.NRGBA, error) {
	if v == nil {
		return nil, nil
	}
	c, err := ParseColor(v)
	if err != nil {
		return nil, err
	}
	n := c.NRGBA()
	return &n, nil
}

// DrawPixel sets the pixel at (x, y) on every frame.
func (img *Image) DrawPixel(x, y int, col any) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawPixel(n, x, y, c.NRGBA()), nil
	})
}

// Fill paints every pixel of every frame with col.
// Invalid color values set Err(); they are never replaced with white.
// Error handling: Check img.Err() after call chain.
func (img *Image) Fill(col any) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Fill(n, c.NRGBA()), nil
	})
}

// FloodFill replaces the 4-connected region of equal color at (x, y) with col
// on every frame.
func (img *Image) FloodFill(x, y int, col any) *Image {
	if img.fail() {
		return img
	}
	c, err := ParseColor(col)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.FloodFill(n, x, y, c.NRGBA()), nil
	})
}

// DrawRectangle paints a rectangle whose top-left is (x, y).
func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image {
	if img.fail() {
		return img
	}
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	bg, err := nrgbaPtr(d.Background)
	if err != nil {
		return img.setErr(err)
	}
	bd, err := nrgbaPtr(d.BorderColor)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawRectangle(n, x, y, d.Width, d.Height, bg, d.BorderSize, bd), nil
	})
}

// DrawEllipse paints an ellipse centered at (x, y).
func (img *Image) DrawEllipse(x, y int, init func(*Drawable)) *Image {
	if img.fail() {
		return img
	}
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	rx, ry := d.Width/2, d.Height/2
	if d.Radius > 0 {
		rx, ry = d.Radius, d.Radius
	}
	bg, err := nrgbaPtr(d.Background)
	if err != nil {
		return img.setErr(err)
	}
	bd, err := nrgbaPtr(d.BorderColor)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawEllipse(n, x, y, rx, ry, bg, d.BorderSize, bd), nil
	})
}

// DrawCircle paints a circle centered at (x, y).
func (img *Image) DrawCircle(x, y int, init func(*Drawable)) *Image {
	if img.fail() {
		return img
	}
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	r := d.Radius
	if r == 0 {
		r = d.Width / 2
	}
	bg, err := nrgbaPtr(d.Background)
	if err != nil {
		return img.setErr(err)
	}
	bd, err := nrgbaPtr(d.BorderColor)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawEllipse(n, x, y, r, r, bg, d.BorderSize, bd), nil
	})
}

// DrawPolygon paints the polygon described by Drawable.Points.
func (img *Image) DrawPolygon(init func(*Drawable)) *Image {
	if img.fail() {
		return img
	}
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	bg, err := nrgbaPtr(d.Background)
	if err != nil {
		return img.setErr(err)
	}
	bd, err := nrgbaPtr(d.BorderColor)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawPolygon(n, pointsToMod(d.Points), bg, d.BorderSize, bd), nil
	})
}

// DrawLine paints a stroke between Drawable.X1/Y1 and X2/Y2.
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
		return modifier.DrawLine(n, d.X1, d.Y1, d.X2, d.Y2, w, c.NRGBA()), nil
	})
}

// DrawBezier paints a quadratic/cubic path through Drawable.Points.
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
		return modifier.DrawBezier(n, pointsToMod(d.Points), w, c.NRGBA()), nil
	})
}

// Font describes a TrueType/OpenType face used by Text. Zero Font uses
// basicfont. It is not safe for concurrent mutation.
type Font struct {
	filename    string
	size        float64
	angle       float64
	color       any
	strokeColor any
	strokeWidth int
	align       string
	valign      string
	lineHeight  float64
	wrapWidth   int
}

// NewFont returns a Font. An optional filename loads a TTF/OTF file.
func NewFont(filename ...string) *Font {
	f := &Font{
		size:        12,
		color:       "000000",
		strokeColor: "ffffff",
		align:       "left",
		valign:      "bottom",
		lineHeight:  1.25,
	}
	if len(filename) > 0 && filename[0] != "" {
		f.filename = filename[0]
	}
	return f
}

// Filename sets the TTF/OTF path.
func (f *Font) Filename(path string) *Font {
	f.filename = path
	return f
}

// File is an alias of Filename.
//
// Deprecated: use Filename.
func (f *Font) File(path string) *Font { return f.Filename(path) }

// Size sets the em size in pixels.
func (f *Font) Size(v float64) *Font { f.size = v; return f }

// Angle sets clockwise rotation in degrees.
func (f *Font) Angle(v float64) *Font { f.angle = v; return f }

// Color sets the fill color.
func (f *Font) Color(v any) *Font { f.color = v; return f }

// Align sets horizontal alignment (left, center, right).
func (f *Font) Align(v string) *Font { f.align = v; return f }

// Valign sets vertical alignment (top, center, bottom).
func (f *Font) Valign(v string) *Font { f.valign = v; return f }

// LineHeight sets the multiplier used when wrapping lines.
func (f *Font) LineHeight(v float64) *Font {
	f.lineHeight = v
	return f
}

// Wrap sets the wrap width in pixels; 0 disables wrapping.
func (f *Font) Wrap(width int) *Font { f.wrapWidth = width; return f }

// Stroke sets the outline color and width (0–10).
func (f *Font) Stroke(col any, width int) *Font {
	if width < 0 {
		width = 0
	}
	if width > 10 {
		width = 10
	}
	f.strokeColor = col
	f.strokeWidth = width
	return f
}

func applyFontInit(v any) (*Font, error) {
	switch t := v.(type) {
	case nil:
		return NewFont(), nil
	case *Font:
		return t, nil
	case Font:
		return &t, nil
	case func(*Font):
		f := NewFont()
		t(f)
		return f, nil
	default:
		return nil, wrap(ErrFont, "invalid font initializer %T", v)
	}
}

// Text draws text at (x, y) using fontInit (*Font, Font, or func(*Font)).
func (img *Image) Text(text string, x, y int, fontInit any) *Image {
	if img.fail() {
		return img
	}
	fnt, err := applyFontInit(fontInit)
	if err != nil {
		return img.setErr(err)
	}
	col, err := ParseColor(fnt.color)
	if err != nil {
		return img.setErr(err)
	}
	strokeCol := ColorWhite
	if fnt.strokeWidth > 0 {
		sc, e := ParseColor(fnt.strokeColor)
		if e != nil {
			return img.setErr(e)
		}
		strokeCol = sc
	}
	style := modifier.TextStyle{
		Filename:    fnt.filename,
		Size:        fnt.size,
		Angle:       fnt.angle,
		Color:       col.NRGBA(),
		StrokeColor: strokeCol.NRGBA(),
		StrokeWidth: fnt.strokeWidth,
		Align:       fnt.align,
		Valign:      fnt.valign,
		LineHeight:  fnt.lineHeight,
		WrapWidth:   fnt.wrapWidth,
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawText(n, text, x, y, style)
	})
}

func (img *Image) encodeOpts(opts []EncodeOptions) (EncodeOptions, error) {
	if len(opts) > 1 {
		return EncodeOptions{}, wrap(ErrInput, "Encode accepts at most one EncodeOptions value")
	}
	if len(opts) == 1 {
		return opts[0], nil
	}
	return EncodeOptions{}, nil
}

// Encode writes img in format. At most one EncodeOptions value is accepted.
// Failures return an EncodedImage whose Err, WriteTo, Save, and Result report
// the same error with encode context.
func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage {
	if img.fail() {
		return encodedErr(img.err)
	}
	if !format.supportedEncode() {
		return encodedErr(wrap(ErrNotSupported, "encoding %s is not supported by the Go driver", format))
	}
	o, err := img.encodeOpts(opts)
	if err != nil {
		return encodedErr(err)
	}
	src := img
	if img.cfg.Strip {
		src = img.Clone().RemoveProfile()
	}
	var data []byte
	switch format {
	case FormatJPEG:
		data, err = encoder.EncodeJPEG(stillForOpaque(src), o.qualityOrDefault())
	case FormatPNG:
		data, err = encoder.EncodePNG(src.primary())
	case FormatGIF:
		data, err = encodeGIF(src)
	case FormatWEBP:
		data, err = encoder.EncodeWebP(src.primary())
	case FormatBMP:
		data, err = encoder.EncodeBMP(stillForOpaque(src))
	case FormatTIFF:
		data, err = encoder.EncodeTIFF(src.primary())
	default:
		err = wrap(ErrNotSupported, "encoding %s is not supported", format)
	}
	if err != nil {
		return encodedErr(annotateEncode(format, err))
	}
	return EncodedImage{Data: data, MediaType: format.MediaType()}
}

func annotateEncode(format Format, err error) error {
	if err == nil {
		return nil
	}
	var e *errs.Error
	if errors.As(err, &e) {
		return err
	}
	return wrap(ErrEncoder, "%s encode: %v", format, err)
}

func stillForOpaque(img *Image) image.Image {
	src := img.primary()
	if src == nil {
		return pool.Acquire(1, 1)
	}
	bg := image.NewUniform(img.BlendingColor().NRGBA())
	out := encoder.StillForOpaque(src, bg)
	return out
}

func encodeGIF(img *Image) ([]byte, error) {
	w, h := img.Width(), img.Height()
	frames := make([]encoder.GIFFrame, 0, len(img.frames))
	for _, f := range img.frames {
		frames = append(frames, encoder.GIFFrame{
			Img:        f.img,
			DelayCS:    modifier.DelayToGIF(f.Delay),
			Disposal:   modifier.GIFDisposal(f.asAnim(), w, h),
			OffsetLeft: f.OffsetLeft,
			OffsetTop:  f.OffsetTop,
		})
	}
	return encoder.EncodeGIF(frames, img.loops, w, h)
}

// EncodeByMediaType encodes using a MIME type, falling back to Origin.MediaType.
func (img *Image) EncodeByMediaType(mediaType string, opts ...EncodeOptions) EncodedImage {
	if mediaType == "" && img != nil {
		mediaType = img.origin.MediaType
	}
	f, err := parseFormat(mediaType)
	if err != nil {
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

// EncodeByExtension encodes using a file extension.
func (img *Image) EncodeByExtension(ext string, opts ...EncodeOptions) EncodedImage {
	if ext == "" && img != nil {
		ext = img.origin.FileExtension()
	}
	f, err := parseFormat(ext)
	if err != nil {
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

// EncodeByPath encodes using the extension of path.
func (img *Image) EncodeByPath(path string, opts ...EncodeOptions) EncodedImage {
	if path == "" && img != nil {
		path = img.origin.FilePath
	}
	f, err := formatFromPath(path)
	if err != nil {
		if img != nil && img.origin.MediaType != "" && img.origin.MediaType != "application/octet-stream" {
			return img.EncodeByMediaType(img.origin.MediaType, opts...)
		}
		return encodedErr(err)
	}
	return img.Encode(f, opts...)
}

// ToJPEG encodes a JPEG. Omitting quality uses 80; a single value sets quality.
// Extra quality arguments are an error.
// Error handling: Check the returned EncodedImage with Err, Result, Save, or WriteTo.
func (img *Image) ToJPEG(quality ...int) EncodedImage {
	if len(quality) > 1 {
		return encodedErr(wrap(ErrInput, "ToJPEG accepts at most one quality value"))
	}
	o := EncodeOptions{}
	if len(quality) == 1 {
		o.Quality = quality[0]
	}
	return img.Encode(FormatJPEG, o)
}
// ToJPG is an alias of ToJPEG.
//
// Deprecated: use ToJPEG.
func (img *Image) ToJPG(quality ...int) EncodedImage { return img.ToJPEG(quality...) }

// ToPNG encodes a PNG. At most one EncodeOptions value is accepted.
// Error handling: Check the returned EncodedImage with Err, Result, Save, or WriteTo.
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatPNG, opts...) }

// ToGIF encodes a GIF. At most one EncodeOptions value is accepted.
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatGIF, opts...) }

// ToWebP encodes a lossless VP8L WebP. At most one EncodeOptions value is accepted.
// Error handling: Check the returned EncodedImage with Err, Result, Save, or WriteTo.
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatWEBP, opts...) }

// ToBMP encodes a BMP.
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage {
	return img.Encode(FormatBMP, opts...)
}

// ToBitmap is an alias of ToBMP.
//
// Deprecated: use ToBMP.
func (img *Image) ToBitmap(opts ...EncodeOptions) EncodedImage { return img.ToBMP(opts...) }

// ToTIFF encodes a TIFF.
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatTIFF, opts...) }

// ToTIF is an alias of ToTIFF.
//
// Deprecated: use ToTIFF.
func (img *Image) ToTIF(opts ...EncodeOptions) EncodedImage { return img.ToTIFF(opts...) }

// ToJPEG2000 returns ErrNotSupported; the Go driver does not encode JPEG 2000.
func (img *Image) ToJPEG2000(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "JPEG 2000 encoding is not supported by the Go driver"))
}

// ToJP2 is an alias of ToJPEG2000.
//
// Deprecated: use ToJPEG2000.
func (img *Image) ToJP2(opts ...EncodeOptions) EncodedImage { return img.ToJPEG2000(opts...) }

// ToAVIF returns ErrNotSupported; the Go driver does not encode AVIF.
func (img *Image) ToAVIF(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "AVIF encoding is not supported by the Go driver"))
}
// ToHEIC returns ErrNotSupported; the Go driver does not encode HEIC.
func (img *Image) ToHEIC(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "HEIC encoding is not supported by the Go driver"))
}

// Save encodes img using the path extension and writes the file.
func (img *Image) Save(path string, opts ...EncodeOptions) *Image {
	if img.fail() {
		return img
	}
	if path == "" {
		path = img.origin.FilePath
	}
	if path == "" {
		return img.setErr(wrap(ErrEncoder, "could not determine file path to save"))
	}
	enc := img.EncodeByPath(path, opts...)
	if enc.err != nil {
		return img.setErr(enc.err)
	}
	if err := enc.Save(path); err != nil {
		return img.setErr(err)
	}
	return img
}

// EncodedImage is a format-encoded byte buffer. Check Err or Result before
// using Data. WriteTo and Save return the same delayed error, so callers can
// pick either the fluent Err() style or a direct error return.
type EncodedImage struct {
	Data      []byte
	MediaType string
	err       error
}

// Err returns the encode or write error, if any.
func (e EncodedImage) Err() error { return e.err }

// Result returns the encoded payload or the delayed error.
func (e EncodedImage) Result() ([]byte, error) {
	if e.err != nil {
		return nil, e.err
	}
	return e.Data, nil
}

// Bytes returns the encoded payload. It may be nil when Err is set.
func (e EncodedImage) Bytes() []byte { return e.Data }

// MimeType returns the encoded media type.
func (e EncodedImage) MimeType() string { return e.MediaType }

// Size returns the payload length in bytes.
func (e EncodedImage) Size() int { return len(e.Data) }

// ToDataURI returns a data: URI for the payload. An empty string is returned
// when Err is set.
func (e EncodedImage) ToDataURI() string {
	if e.err != nil {
		return ""
	}
	return "data:" + e.MediaType + ";base64," + base64.StdEncoding.EncodeToString(e.Data)
}

// String returns the payload as a Go string.
func (e EncodedImage) String() string { return string(e.Data) }

// WriteTo writes the payload to w. A delayed encode error is returned first.
func (e EncodedImage) WriteTo(w io.Writer) (int64, error) {
	if e.err != nil {
		return 0, e.err
	}
	if w == nil {
		return 0, wrap(ErrEncoder, "nil writer")
	}
	n, err := w.Write(e.Data)
	if err != nil {
		return int64(n), wrap(ErrEncoder, "write encoded image: %v", err)
	}
	return int64(n), nil
}

// Save writes the payload to path, creating parent directories as needed.
func (e EncodedImage) Save(path string) error {
	if e.err != nil {
		return e.err
	}
	if path == "" {
		return wrap(ErrEncoder, "could not determine file path to save")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		if filepath.Dir(path) != "." {
			return wrap(ErrNotWritable, "unable to write %s: %v", path, err)
		}
	}
	if err := os.WriteFile(path, e.Data, 0o644); err != nil {
		return wrap(ErrNotWritable, "unable to write %s: %v", path, err)
	}
	return nil
}

func encodedErr(err error) EncodedImage {
	return EncodedImage{err: err, MediaType: "application/octet-stream"}
}

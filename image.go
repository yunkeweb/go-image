package goimage

import (
	"encoding/base64"
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

type Size = modifier.Size
type Point = modifier.Point

// Color is an 8-bit-per-channel sRGB color with alpha (255 = opaque).
type Color struct {
	R, G, B, A uint8
}

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

func mustColor(v any) Color {
	c, err := ParseColor(v)
	if err != nil {
		return ColorWhite
	}
	return c
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

func (c Color) NRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func (c Color) RGBA() (r, g, b, a uint32) {
	return c.NRGBA().RGBA()
}

func colorFromNRGBA(n color.NRGBA) Color {
	return Color{R: n.R, G: n.G, B: n.B, A: n.A}
}

func (c Color) IsTransparent() bool { return c.A < 255 }
func (c Color) IsClear() bool       { return c.A == 0 }
func (c Color) IsGreyscale() bool   { return c.R == c.G && c.G == c.B }

func (c Color) ToHex(prefix string) string {
	if c.IsTransparent() {
		return fmt.Sprintf("%s%02x%02x%02x%02x", prefix, c.R, c.G, c.B, c.A)
	}
	return fmt.Sprintf("%s%02x%02x%02x", prefix, c.R, c.G, c.B)
}

func (c Color) String() string {
	if c.IsTransparent() {
		return fmt.Sprintf("rgba(%d, %d, %d, %.1f)", c.R, c.G, c.B, float64(c.A)/255)
	}
	return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
}

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

func (f Format) MediaType() string     { return encoder.MediaType(string(f)) }
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

type Origin struct {
	MediaType string
	FilePath  string
}

func (o Origin) MimeType() string { return o.MediaType }

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

type Frame struct {
	Img        *image.NRGBA
	Delay      float64
	Dispose    int
	OffsetLeft int
	OffsetTop  int
}

func (f Frame) Size() Size {
	if f.Img == nil {
		return Size{}
	}
	b := f.Img.Bounds()
	return Size{Width: b.Dx(), Height: b.Dy()}
}

func (f Frame) clone() Frame {
	out := f
	if f.Img != nil {
		out.Img = pool.Clone(f.Img)
	}
	return out
}

func (f Frame) asAnim() modifier.AnimFrame {
	return modifier.AnimFrame(f)
}

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
	return &Image{err: err, exif: map[string]any{}, cfg: defaultConfig()}
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
	return img.frames[0].Img
}

func (img *Image) Clone() *Image {
	if img == nil {
		return failed(wrap(ErrRuntime, "nil image"))
	}
	cp := *img
	cp.frames = make([]Frame, len(img.frames))
	for i := range img.frames {
		cp.frames[i] = img.frames[i].clone()
	}
	if img.exif != nil {
		cp.exif = make(map[string]any, len(img.exif))
		for k, v := range img.exif {
			cp.exif[k] = v
		}
	}
	if img.profile != nil {
		cp.profile = append([]byte(nil), img.profile...)
	}
	return &cp
}

func (img *Image) Native() image.Image {
	if img.fail() {
		return pool.Acquire(1, 1)
	}
	return img.primary()
}

func (img *Image) Frames() []Frame {
	if img.fail() {
		return nil
	}
	return img.frames
}

func (img *Image) Origin() Origin {
	if img == nil {
		return Origin{}
	}
	return img.origin
}

func (img *Image) SetOrigin(o Origin) *Image {
	if img.fail() {
		return img
	}
	img.origin = o
	return img
}

func (img *Image) Count() int {
	if img.fail() {
		return 0
	}
	return len(img.frames)
}

func (img *Image) IsAnimated() bool { return img.Count() > 1 }

func (img *Image) Loops() int {
	if img.fail() {
		return 0
	}
	return img.loops
}

func (img *Image) SetLoops(n int) *Image {
	if img.fail() {
		return img
	}
	img.loops = n
	return img
}

func (img *Image) Width() int {
	if img.fail() || img.primary() == nil {
		return 0
	}
	return img.primary().Bounds().Dx()
}

func (img *Image) Height() int {
	if img.fail() || img.primary() == nil {
		return 0
	}
	return img.primary().Bounds().Dy()
}

func (img *Image) Size() Size {
	return Size{Width: img.Width(), Height: img.Height()}
}

func (img *Image) Colorspace() ColorspaceName {
	if img.fail() {
		return ColorspaceRGB
	}
	return img.colorspace
}

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

func (img *Image) Resolution() (x, y float64) {
	if img.fail() {
		return 0, 0
	}
	return img.resX, img.resY
}

func (img *Image) SetResolution(x, y float64) *Image {
	if img.fail() {
		return img
	}
	img.resX, img.resY = x, y
	return img
}

func (img *Image) PickColor(x, y int) Color {
	return img.PickColorFrame(x, y, 0)
}

func (img *Image) PickColorFrame(x, y, frame int) Color {
	if img.fail() || frame < 0 || frame >= len(img.frames) || img.frames[frame].Img == nil {
		return Color{}
	}
	n := img.frames[frame].Img
	b := n.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return Color{}
	}
	return colorFromNRGBA(n.NRGBAAt(x, y))
}

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

func (img *Image) Exif() map[string]any {
	if img.fail() {
		return nil
	}
	return img.exif
}

func (img *Image) ExifQuery(key string) any {
	if img.fail() || img.exif == nil {
		return nil
	}
	if v, ok := img.exif[key]; ok {
		return v
	}
	if v, ok := img.exif["IFD0."+key]; ok {
		return v
	}
	return img.exif["EXIF."+key]
}

func (img *Image) SetExif(data map[string]any) *Image {
	if img.fail() {
		return img
	}
	img.exif = data
	return img
}

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

func (img *Image) Profile() []byte {
	if img.fail() {
		return nil
	}
	return img.profile
}

func (img *Image) SetProfile(data []byte) *Image {
	if img.fail() {
		return img
	}
	img.profile = append([]byte(nil), data...)
	return img
}

func (img *Image) RemoveProfile() *Image {
	if img.fail() {
		return img
	}
	img.profile = nil
	return img
}

func (img *Image) Config() Config {
	if img == nil {
		return defaultConfig()
	}
	return img.cfg
}

func (img *Image) replaceAll(fn func(*image.NRGBA) (*image.NRGBA, error)) *Image {
	return img.eachFrame(func(f *Frame) error {
		old := f.Img
		n, err := fn(old)
		if err != nil {
			return err
		}
		if n != old {
			pool.Release(old)
		}
		f.Img = n
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
		img.frames[i] = Frame(f)
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
		pool.Release(img.frames[i].Img)
		img.frames[i].Img = nil
	}
}

func (img *Image) Resize(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.Resize(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) ResizeDown(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.ResizeDown(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) Scale(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.Scale(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) ScaleDown(width int, height ...int) *Image {
	if img.fail() {
		return img
	}
	r, err := modifier.SizeFromArgs(width, height)
	if err != nil {
		return img.setErr(err)
	}
	target := r.ScaleDown(img.Size())
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Resample(n, n.Bounds(), target.Width, target.Height), nil
	})
}

func (img *Image) Cover(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "center"}, opts)
	crop, resizeTo, err := modifier.CoverSizes(img.Size(), width, height, cfg.anchor, false)
	if err != nil {
		return img.setErr(err)
	}
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.ApplyCover(n, crop, resizeTo), nil
	})
}

func (img *Image) CoverDown(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "center"}, opts)
	crop, _, err := modifier.CoverSizes(img.Size(), width, height, cfg.anchor, true)
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

func (img *Image) Fit(width, height int, opts ...GeometryOption) *Image {
	return img.Cover(width, height, opts...)
}

func (img *Image) Contain(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "center", background: "ffffff"}, opts)
	bg := mustColor(cfg.background)
	r, err := modifier.NewResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.Contain(img.Size())
	if err != nil {
		return img.setErr(err)
	}
	canvas := Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, cfg.anchor)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.PlaceOnCanvas(n, width, height, crop, bg.NRGBA()), nil
	})
}

func (img *Image) Pad(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "center", background: "ffffff"}, opts)
	bg := mustColor(cfg.background)
	r, err := modifier.NewResizer(width, height)
	if err != nil {
		return img.setErr(err)
	}
	crop, err := r.ContainDown(img.Size())
	if err != nil {
		return img.setErr(err)
	}
	canvas := Size{Width: width, Height: height}
	crop = crop.AlignPivotTo(canvas, cfg.anchor)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.PlaceOnCanvas(n, width, height, crop, bg.NRGBA()), nil
	})
}

func (img *Image) Crop(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "top-left", background: "ffffff"}, opts)
	bg := mustColor(cfg.background)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Crop(n, width, height, cfg.anchor, bg.NRGBA(), cfg.offsetX, cfg.offsetY)
	})
}

func (img *Image) ResizeCanvas(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	cfg := applyGeometryOptions(geometrySettings{anchor: "center", background: "ffffff"}, opts)
	bg := mustColor(cfg.background)
	return img.replaceAllGeometry(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.ResizeCanvas(n, width, height, cfg.anchor, bg.NRGBA())
	})
}

func (img *Image) ResizeCanvasRelative(width, height int, opts ...GeometryOption) *Image {
	if img.fail() {
		return img
	}
	return img.ResizeCanvas(img.Width()+width, img.Height()+height, opts...)
}

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
	img.frames[0].Img = cropped
	img.resetGIFFrameLayout()
	return img
}

func (img *Image) Greyscale() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Greyscale(n), nil
	})
}

func (img *Image) Invert() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Invert(n), nil
	})
}

func (img *Image) Brightness(level int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Brightness(n, level), nil
	})
}

func (img *Image) Contrast(level int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Contrast(n, level), nil
	})
}

func (img *Image) Gamma(gamma float64) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Gamma(n, gamma)
	})
}

func (img *Image) Colorize(red, green, blue int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Colorize(n, red, green, blue), nil
	})
}

func (img *Image) Pixelate(size int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Pixelate(n, size), nil
	})
}

func (img *Image) Blur(amount int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Blur(n, amount), nil
	})
}

func (img *Image) Sharpen(amount int) *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Sharpen(n, amount), nil
	})
}

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

func (img *Image) ReduceColors(limit int, background any) *Image {
	if img.fail() {
		return img
	}
	if limit < 1 {
		limit = 1
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		pal := encoder.MedianCutPalette(n, limit)
		return modifier.ReduceColors(n, pal), nil
	})
}

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

func (img *Image) Flip() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Flip(n), nil
	})
}

func (img *Image) Flop() *Image {
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Flop(n), nil
	})
}

func (img *Image) Rotate(angle float64, background any) *Image {
	if img.fail() {
		return img
	}
	bg, err := ParseColor(background)
	if err != nil {
		bg = ColorWhite
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

func (img *Image) Orient() *Image {
	if img.fail() {
		return img
	}
	v := img.ExifQuery("Orientation")
	orient := 1
	switch t := v.(type) {
	case int:
		orient = t
	case float64:
		orient = int(t)
	}
	applyOrientation(img, orient)
	img.markOrientationNormal()
	return img
}

func (img *Image) Orientate() *Image { return img.Orient() }

func (img *Image) markOrientationNormal() {
	if img == nil {
		return
	}
	if img.exif == nil {
		img.exif = map[string]any{}
	}
	img.exif["Orientation"] = 1
	img.exif["IFD0.Orientation"] = 1
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

func (img *Image) Place(src *Image, position string, offsetX, offsetY, opacity int) *Image {
	if img.fail() {
		return img
	}
	if src == nil {
		return img.setErr(wrap(ErrInput, "nil watermark"))
	}
	if src.Err() != nil {
		return img.setErr(src.Err())
	}
	overlay := src.primary()
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.Place(n, overlay, position, offsetX, offsetY, opacity), nil
	})
}

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

func nrgbaPtr(v any) *color.NRGBA {
	if v == nil {
		return nil
	}
	c, err := ParseColor(v)
	if err != nil {
		return nil
	}
	n := c.NRGBA()
	return &n
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
		return modifier.DrawPixel(n, x, y, c.NRGBA()), nil
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
		if flood {
			return modifier.FloodFill(n, xy[0], xy[1], c.NRGBA()), nil
		}
		return modifier.Fill(n, c.NRGBA()), nil
	})
}

func (img *Image) DrawRectangle(x, y int, init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawRectangle(n, x, y, d.Width, d.Height, nrgbaPtr(d.Background), d.BorderSize, nrgbaPtr(d.BorderColor)), nil
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
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawEllipse(n, x, y, rx, ry, nrgbaPtr(d.Background), d.BorderSize, nrgbaPtr(d.BorderColor)), nil
	})
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
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawEllipse(n, x, y, r, r, nrgbaPtr(d.Background), d.BorderSize, nrgbaPtr(d.BorderColor)), nil
	})
}

func (img *Image) DrawPolygon(init func(*Drawable)) *Image {
	d := &Drawable{}
	if init != nil {
		init(d)
	}
	return img.replaceAll(func(n *image.NRGBA) (*image.NRGBA, error) {
		return modifier.DrawPolygon(n, d.Points, nrgbaPtr(d.Background), d.BorderSize, nrgbaPtr(d.BorderColor)), nil
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
		return modifier.DrawLine(n, d.X1, d.Y1, d.X2, d.Y2, w, c.NRGBA()), nil
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
		return modifier.DrawBezier(n, d.Points, w, c.NRGBA()), nil
	})
}

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

func (f *Font) Filename(path string) *Font {
	f.filename = path
	return f
}
func (f *Font) File(path string) *Font { return f.Filename(path) }
func (f *Font) Size(v float64) *Font   { f.size = v; return f }
func (f *Font) Angle(v float64) *Font  { f.angle = v; return f }
func (f *Font) Color(v any) *Font      { f.color = v; return f }
func (f *Font) Align(v string) *Font   { f.align = v; return f }
func (f *Font) Valign(v string) *Font  { f.valign = v; return f }
func (f *Font) LineHeight(v float64) *Font {
	f.lineHeight = v
	return f
}
func (f *Font) Wrap(width int) *Font { f.wrapWidth = width; return f }
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
		col = ColorBlack
	}
	strokeCol := ColorWhite
	if fnt.strokeWidth > 0 {
		if sc, e := ParseColor(fnt.strokeColor); e == nil {
			strokeCol = sc
		}
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

func (img *Image) encodeOpts(opts []EncodeOptions) EncodeOptions {
	if len(opts) > 0 {
		return opts[0]
	}
	return EncodeOptions{}
}

func (img *Image) Encode(format Format, opts ...EncodeOptions) EncodedImage {
	if img.fail() {
		return encodedErr(img.err)
	}
	if !format.supportedEncode() {
		return encodedErr(wrap(ErrNotSupported, "encoding %s is not supported by the Go driver", format))
	}
	o := img.encodeOpts(opts)
	src := img
	if img.cfg.Strip {
		src = img.Clone().RemoveProfile()
	}
	var (
		data []byte
		err  error
	)
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
		return encodedErr(err)
	}
	return EncodedImage{Data: data, MediaType: format.MediaType()}
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
			Img:        f.Img,
			DelayCS:    modifier.DelayToGIF(f.Delay),
			Disposal:   modifier.GIFDisposal(f.asAnim(), w, h),
			OffsetLeft: f.OffsetLeft,
			OffsetTop:  f.OffsetTop,
		})
	}
	return encoder.EncodeGIF(frames, img.loops, w, h)
}

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

func (img *Image) ToJPEG(quality ...int) EncodedImage {
	o := EncodeOptions{}
	if len(quality) > 0 {
		o.Quality = quality[0]
	}
	return img.Encode(FormatJPEG, o)
}
func (img *Image) ToJPG(quality ...int) EncodedImage         { return img.ToJPEG(quality...) }
func (img *Image) ToPNG(opts ...EncodeOptions) EncodedImage  { return img.Encode(FormatPNG, opts...) }
func (img *Image) ToGIF(opts ...EncodeOptions) EncodedImage  { return img.Encode(FormatGIF, opts...) }
func (img *Image) ToWebP(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatWEBP, opts...) }
func (img *Image) ToBitmap(opts ...EncodeOptions) EncodedImage {
	return img.Encode(FormatBMP, opts...)
}
func (img *Image) ToBMP(opts ...EncodeOptions) EncodedImage  { return img.ToBitmap(opts...) }
func (img *Image) ToTIFF(opts ...EncodeOptions) EncodedImage { return img.Encode(FormatTIFF, opts...) }
func (img *Image) ToTIF(opts ...EncodeOptions) EncodedImage  { return img.ToTIFF(opts...) }

func (img *Image) ToJPEG2000(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "JPEG 2000 encoding is not supported by the Go driver"))
}
func (img *Image) ToJP2(opts ...EncodeOptions) EncodedImage { return img.ToJPEG2000(opts...) }
func (img *Image) ToAVIF(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "AVIF encoding is not supported by the Go driver"))
}
func (img *Image) ToHEIC(opts ...EncodeOptions) EncodedImage {
	return encodedErr(wrap(ErrNotSupported, "HEIC encoding is not supported by the Go driver"))
}

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

type EncodedImage struct {
	Data      []byte
	MediaType string
	err       error
}

func (e EncodedImage) Err() error { return e.err }

func (e EncodedImage) Bytes() []byte { return e.Data }

func (e EncodedImage) MimeType() string { return e.MediaType }

func (e EncodedImage) Size() int { return len(e.Data) }

func (e EncodedImage) ToDataURI() string {
	return "data:" + e.MediaType + ";base64," + base64.StdEncoding.EncodeToString(e.Data)
}

func (e EncodedImage) String() string { return string(e.Data) }

func (e EncodedImage) WriteTo(w io.Writer) (int64, error) {
	if e.err != nil {
		return 0, e.err
	}
	if w == nil {
		return 0, wrap(ErrEncoder, "nil writer")
	}
	n, err := w.Write(e.Data)
	return int64(n), err
}

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

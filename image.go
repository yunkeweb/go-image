package goimage

import (
	"image"
	"image/color"
	"strconv"
	"strings"
)

// Image is the fluent image object. Methods mutate the receiver and return it
// so callers can chain operations. PHP exceptions become a delayed error
// retrieved with Err().
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

// Err returns the first delayed error from a fluent chain.
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

// Clone returns a deep copy including frames.
func (img *Image) Clone() *Image {
	if img.fail() {
		out := *img
		return &out
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
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
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

func pixelAt(n *image.NRGBA, x, y int) color.NRGBA {
	b := n.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		return color.NRGBA{}
	}
	return n.NRGBAAt(x, y)
}

func parsePercentOrIndex(v any, total int) (int, error) {
	switch t := v.(type) {
	case int:
		if t < 0 {
			t = 0
		}
		if total == 0 {
			return 0, nil
		}
		if t >= total {
			return total - 1, nil
		}
		return t, nil
	case string:
		s := strings.TrimSpace(t)
		s = strings.TrimSuffix(s, "%")
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, wrap(ErrInput, "invalid animation position %q", t)
		}
		if total <= 1 {
			return 0, nil
		}
		idx := int(float64(total-1) * f / 100)
		return clampInt(idx, 0, total-1), nil
	default:
		return 0, wrap(ErrInput, "invalid animation position %T", v)
	}
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



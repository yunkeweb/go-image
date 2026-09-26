package goimage

import (
	"os"
	"strings"
)

// Font holds typography options for Text.
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

func (f *Font) hasFile() bool {
	if f == nil || f.filename == "" {
		return false
	}
	st, err := os.Stat(f.filename)
	return err == nil && !st.IsDir()
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

func wrapText(text string, maxWidth int, measure func(string) int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	raw := strings.Split(text, "\n")
	if maxWidth <= 0 {
		return raw
	}
	var lines []string
	for _, para := range raw {
		if para == "" {
			lines = append(lines, "")
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		cur := words[0]
		for _, w := range words[1:] {
			trial := cur + " " + w
			if measure(trial) <= maxWidth {
				cur = trial
			} else {
				lines = append(lines, cur)
				cur = w
			}
		}
		lines = append(lines, cur)
	}
	return lines
}

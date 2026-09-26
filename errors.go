package goimage

import "fmt"

// Sentinel errors matching PHP Intervention Image exception classes.
var (
	ErrRuntime      = fmt.Errorf("runtime error")
	ErrDecoder      = fmt.Errorf("decoder error")
	ErrEncoder      = fmt.Errorf("encoder error")
	ErrGeometry     = fmt.Errorf("geometry error")
	ErrColor        = fmt.Errorf("color error")
	ErrInput        = fmt.Errorf("input error")
	ErrNotSupported = fmt.Errorf("not supported")
	ErrNotWritable  = fmt.Errorf("not writable")
	ErrAnimation    = fmt.Errorf("animation error")
	ErrFont         = fmt.Errorf("font error")
	ErrDriver       = fmt.Errorf("driver error")
)

type imageError struct {
	kind error
	msg  string
}

func (e *imageError) Error() string {
	if e.msg == "" {
		return e.kind.Error()
	}
	return e.msg
}

func (e *imageError) Unwrap() error { return e.kind }

func wrap(kind error, format string, args ...any) error {
	return &imageError{kind: kind, msg: fmt.Sprintf(format, args...)}
}

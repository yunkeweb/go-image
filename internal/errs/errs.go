package errs

import "fmt"

// Sentinel errors shared by the root package, modifier, and encoder.
var (
	ErrRuntime           = fmt.Errorf("runtime error")
	ErrDecoder           = fmt.Errorf("decoder error")
	ErrEncoder           = fmt.Errorf("encoder error")
	ErrGeometry          = fmt.Errorf("geometry error")
	ErrInvalidDimensions = fmt.Errorf("invalid dimensions")
	ErrColor             = fmt.Errorf("color error")
	ErrInput             = fmt.Errorf("input error")
	ErrNotSupported      = fmt.Errorf("not supported")
	ErrNotWritable       = fmt.Errorf("not writable")
	ErrAnimation         = fmt.Errorf("animation error")
	ErrFont              = fmt.Errorf("font error")
	ErrDriver            = fmt.Errorf("driver error")
)

type Error struct {
	Kind error
	Msg  string
}

func (e *Error) Error() string {
	if e.Msg == "" {
		return e.Kind.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error { return e.Kind }

func Wrap(kind error, format string, args ...any) error {
	return &Error{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}

package errhelper

var _ error = (*ErrString)(nil)

// ErrString is a string type implementing error type
type ErrString string

func (s ErrString) Error() string { return string(s) }

package rs

import (
	"errors"
)

var (
	// ErrInterfaceTypeNotHandled is expected to be throwed by Options.InterfaceTypeHandler
	// when the requested interface type was not handled
	//
	// it is only considered as an error when the target interface type is not `any` (interface{}),
	// or in other words,
	// it is not considered an error when InterfaceTypeHandler returned this error with `interface{}` type
	// as expected typed
	ErrInterfaceTypeNotHandled = errors.New("interface type not handled")
)

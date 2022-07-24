package rs

import (
	"reflect"
)

// nolint:revive
var (
	type_string = reflect.TypeOf("")

	typePtr_BaseField    = reflect.TypeOf((*BaseField)(nil))
	typeStruct_BaseField = typePtr_BaseField.Elem()

	typeEface_Field = reflect.TypeOf((*Field)(nil)).Elem()
	typeEface_Any   = reflect.TypeOf((*any)(nil)).Elem()
)

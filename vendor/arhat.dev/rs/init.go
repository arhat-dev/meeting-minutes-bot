package rs

import (
	"reflect"
)

type Options struct {
	// InterfaceTypeHandler handles interface value creation for any field
	// with a interface{...} type (including []interface{...} and map[string]interface{...})
	//
	// for example, define a interface type Foo:
	// 		type Foo interface{ Bar() string }
	//
	// with InterfaceTypeHandler you can return values whose type impelemts Foo during unmarshaling
	//
	// defaults to `nil`
	InterfaceTypeHandler InterfaceTypeHandler

	// DataTagNamespace used for struct field tag parsing
	// you can set it to json to use json tag
	//
	// supported tag values are:
	// - <first tag value as data field name>
	// - inline
	// - omitempty
	//
	// unsupported tag values are ignored
	//
	// defaults to `yaml`
	DataTagNamespace string

	// AllowUnknownFields whether restrict unmarshaling to known fields
	//
	// NOTE: if there is a map field in struct with field type `rs:"other"`
	// even when AllowUnknownFields is set to false, it still gets these unknown
	// fields marshaled to that map field
	//
	// defaults to `false`
	AllowUnknownFields bool

	// AllowedRenderers limit renderers can be applied in rendering suffix
	// when this option is not set (nil), not renderer will be rejected
	// when set, only renderers with exact name matching will be allowed,
	// thus you may need to set an empty entry to allow pseudo built-in
	// empty renderer
	AllowedRenderers map[string]struct{}
}

// Init the BaseField embedded in your struct, the BaseField must be the first field
//
// 		type Foo struct {
// 			rs.BaseField // or *rs.BaseField
// 		}
//
// NOTE: if the arg `in` doesn't contain BaseField or the BaseField is not the first element
// it does nothing and will return `in` as is.
func Init(in Field, opts *Options) Field {
	_ = InitReflectValue(reflect.ValueOf(in), opts)
	return in
}

// InitReflectValue returns true when BaseField of `in` is initilized after the call
func InitReflectValue(in reflect.Value, opts *Options) bool {
	switch in.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		// no pointer to pointer support
		in = in.Elem()

		if in.Kind() != reflect.Struct {
			// the target is not a struct, not using BaseField
			return false
		}
	default:
		return false
	}

	if !in.CanAddr() {
		panic("invalid non addressable value")
	}

	if in.NumField() == 0 {
		// empty struct, no BaseField
		return false
	}

	firstField := in.Field(0)

	var baseField *BaseField
	switch firstField.Type() {
	case typeStruct_BaseField:
		// using BaseField

		baseField = firstField.Addr().Interface().(*BaseField)
	case typePtr_BaseField:
		// using *BaseField

		if firstField.IsZero() {
			// not initialized
			baseField = new(BaseField)
			firstField.Set(reflect.ValueOf(baseField))
		} else {
			baseField = firstField.Interface().(*BaseField)
		}
	default:
		// BaseField is not the first field
		return false
	}

	err := baseField.init(in, opts)
	if err != nil {
		panic(err)
	}

	return true
}

// InitRecursively trys to call Init on all fields implementing Field interface
func InitRecursively(fv reflect.Value, opts *Options) {
	switch fv.Type() {
	case typePtr_BaseField, typeStruct_BaseField:
		return
	}

findStruct:
	switch fv.Kind() {
	case reflect.Struct:
		_ = tryInit(fv, opts)
	case reflect.Ptr:
		if !fv.IsValid() || fv.IsZero() || fv.IsNil() {
			return
		}

		fv = fv.Elem()
		goto findStruct
	default:
		return
	}

	for i := 0; i < fv.NumField(); i++ {
		InitRecursively(fv.Field(i), opts)
	}
}

// return true when Init called on fieldValue
// nolint:unparam
func tryInit(fv reflect.Value, opts *Options) bool {
	if fv.CanInterface() {
		fVal, canCallInit := fv.Interface().(Field)
		if canCallInit {
			_ = Init(fVal, opts)
			return true
		}
	}

	if !fv.CanAddr() {
		return false
	}

	fv = fv.Addr()

	if !fv.CanInterface() {
		return false
	}

	fVal, canCallInit := fv.Interface().(Field)
	if canCallInit {
		_ = Init(fVal, opts)
		return true
	}

	return false
}

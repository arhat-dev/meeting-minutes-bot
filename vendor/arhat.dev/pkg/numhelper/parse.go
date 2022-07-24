package numhelper

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"arhat.dev/pkg/stringhelper"
)

type UnsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Integer interface {
	SignedInteger | UnsignedInteger
}

type Float interface {
	~float32 | ~float64
}

func ParseIntegers[T Integer, V any](arr []V) (ret []T, err error) {
	if len(arr) == 0 {
		return
	}

	ret = make([]T, len(arr))
	for i := range arr {
		ret[i], err = ParseInteger[T](arr[i])
		if err != nil {
			return
		}
	}

	return
}

func ParstFloats[T Float, V any](arr []V) (ret []T, err error) {
	if len(arr) == 0 {
		return
	}

	ret = make([]T, len(arr))
	for i := range arr {
		ret[i], err = ParseFloat[T](arr[i])
		if err != nil {
			return
		}
	}

	return
}

func ParseFloat[T Float](i any) (T, error) {
	iv, isFloat, err := ParseNumber(i)
	if err != nil {
		return 0, err
	}

	if isFloat {
		return T(math.Float64frombits(iv)), nil
	}

	return T(iv), nil
}

func ParseInteger[T Integer](i any) (T, error) {
	ret, isFloat, err := ParseNumber(i)
	if err != nil {
		return 0, err
	}

	if isFloat {
		return T(math.Float64frombits(ret)), nil
	}

	return T(ret), nil
}

// ParseNumber converts i to uint64 (with sign kept)
//
// if i is a float number (indicated by return value isFloat), return IEEE 754 bits of i
func ParseNumber(i any) (_ uint64, isFloat bool, _ error) {
	switch i := i.(type) {
	case string:
		return strToInteger(i)
	case []byte:
		return strToInteger(stringhelper.Convert[string, byte](i))

	case int:
		return uint64(i), false, nil
	case uint:
		return uint64(i), false, nil

	case int8:
		return uint64(i), false, nil
	case uint8:
		return uint64(i), false, nil

	case int16:
		return uint64(i), false, nil
	case uint16:
		return uint64(i), false, nil

	case int32:
		return uint64(i), false, nil
	case uint32:
		return uint64(i), false, nil

	case int64:
		return uint64(i), false, nil
	case uint64:
		return uint64(i), false, nil

	case uintptr:
		return uint64(i), false, nil

	case float32:
		return uint64(math.Float32bits(i)), true, nil
	case float64:
		return math.Float64bits(i), true, nil

	case bool:
		if i {
			return 1, false, nil
		}

		return 0, false, nil

	case nil:
		return 0, false, nil

	default:
		switch val := reflect.Indirect(reflect.ValueOf(i)); val.Kind() {
		case reflect.String:
			return strToInteger(val.String())
		case reflect.Slice, reflect.Array:
			switch typ := val.Elem().Type(); typ.Kind() {
			case reflect.Uint8:
				return strToInteger(*(*string)(val.Addr().UnsafePointer()))
			default:
				return 0, false, fmt.Errorf("unhandled array type %q", typ.String())
			}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			return uint64(val.Int()), false, nil
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
			return val.Uint(), false, nil
		case reflect.Float32, reflect.Float64:
			return math.Float64bits(val.Float()), true, nil
		case reflect.Bool:
			if val.Bool() {
				return 1, false, nil
			}
			return 0, false, nil
		default:
			return 0, false, fmt.Errorf("unhandled value %T", i)
		}
	}
}

func strToInteger(str string) (uint64, bool, error) {
	iv, err := strconv.ParseInt(str, 0, 64)
	if err == nil {
		return uint64(iv), false, nil
	}

	// maybe it's a float?
	fv, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return math.Float64bits(fv), true, nil
	}

	return math.Float64bits(math.NaN()), true, strconv.ErrSyntax
}

/*
Copyright 2019 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

func StringError(err string) Field {
	return String("error", err)
}

var (
	bufferPool      = buffer.NewPool()
	_stacktracePool = sync.Pool{
		New: func() interface{} {
			return newProgramCounters(64)
		},
	}
)

type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}

func takeStacktrace(skip int) string {
	if skip < 0 {
		skip = 4
	}

	buf := bufferPool.Get()
	defer buf.Free()
	programCounters := _stacktracePool.Get().(*programCounters)
	defer _stacktracePool.Put(programCounters)

	var numFrames int
	for {
		// Skip the call to runtime.Counters and takeStacktrace so that the
		// program counters start at the caller of takeStacktrace.
		numFrames = runtime.Callers(skip, programCounters.pcs)
		if numFrames < len(programCounters.pcs) {
			break
		}
		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		programCounters = newProgramCounters(len(programCounters.pcs) * 2)
	}

	i := 0
	frames := runtime.CallersFrames(programCounters.pcs[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if i != 0 {
			buf.AppendByte('\n')
		}
		i++
		buf.AppendString(frame.Function)
		buf.AppendByte('\n')
		buf.AppendByte('\t')
		buf.AppendString(frame.File)
		buf.AppendByte(':')
		buf.AppendInt(int64(frame.Line))
	}

	return buf.String()
}

// following are code from zap

var (
	_minTimeInt64 = time.Unix(0, math.MinInt64)
	_maxTimeInt64 = time.Unix(0, math.MaxInt64)
)

// Skip constructs a no-op field, which is often useful when handling invalid
// inputs in other Field constructors.
func Skip() Field {
	return Field{Type: zapcore.SkipType}
}

// nilField returns a field which will marshal explicitly as nil. See motivation
// in https://github.com/uber-go/zap/issues/753 . If we ever make breaking
// changes and add zapcore.NilType and zapcore.ObjectEncoder.AddNil, the
// implementation here should be changed to reflect that.
func nilField(key string) Field { return Reflect(key, nil) }

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func Binary(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) Field {
	var ival int64
	if val {
		ival = 1
	}
	return Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

// Boolp constructs a field that carries a *bool. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Boolp(key string, val *bool) Field {
	if val == nil {
		return nilField(key)
	}
	return Bool(key, *val)
}

// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func ByteString(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.ByteStringType, Interface: val}
}

// Complex128 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex128 to
// interface{}).
func Complex128(key string, val complex128) Field {
	return Field{Key: key, Type: zapcore.Complex128Type, Interface: val}
}

// Complex128p constructs a field that carries a *complex128. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex128p(key string, val *complex128) Field {
	if val == nil {
		return nilField(key)
	}
	return Complex128(key, *val)
}

// Complex64 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex64 to
// interface{}).
func Complex64(key string, val complex64) Field {
	return Field{Key: key, Type: zapcore.Complex64Type, Interface: val}
}

// Complex64p constructs a field that carries a *complex64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex64p(key string, val *complex64) Field {
	if val == nil {
		return nilField(key)
	}
	return Complex64(key, *val)
}

// Float64 constructs a field that carries a float64. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Float64p constructs a field that carries a *float64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float64p(key string, val *float64) Field {
	if val == nil {
		return nilField(key)
	}
	return Float64(key, *val)
}

// Float32 constructs a field that carries a float32. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float32(key string, val float32) Field {
	return Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(val))}
}

// Float32p constructs a field that carries a *float32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float32p(key string, val *float32) Field {
	if val == nil {
		return nilField(key)
	}
	return Float32(key, *val)
}

// Int constructs a field with the given key and value.
func Int(key string, val int) Field {
	return Int64(key, int64(val))
}

// Intp constructs a field that carries a *int. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Intp(key string, val *int) Field {
	if val == nil {
		return nilField(key)
	}
	return Int(key, *val)
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// Int64p constructs a field that carries a *int64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int64p(key string, val *int64) Field {
	if val == nil {
		return nilField(key)
	}
	return Int64(key, *val)
}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) Field {
	return Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

// Int32p constructs a field that carries a *int32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int32p(key string, val *int32) Field {
	if val == nil {
		return nilField(key)
	}
	return Int32(key, *val)
}

// Int16 constructs a field with the given key and value.
func Int16(key string, val int16) Field {
	return Field{Key: key, Type: zapcore.Int16Type, Integer: int64(val)}
}

// Int16p constructs a field that carries a *int16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int16p(key string, val *int16) Field {
	if val == nil {
		return nilField(key)
	}
	return Int16(key, *val)
}

// Int8 constructs a field with the given key and value.
func Int8(key string, val int8) Field {
	return Field{Key: key, Type: zapcore.Int8Type, Integer: int64(val)}
}

// Int8p constructs a field that carries a *int8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int8p(key string, val *int8) Field {
	if val == nil {
		return nilField(key)
	}
	return Int8(key, *val)
}

// String constructs a field with the given key and value.
func String(key string, val string) Field {
	return Field{Key: key, Type: zapcore.StringType, String: val}
}

// Stringp constructs a field that carries a *string. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Stringp(key string, val *string) Field {
	if val == nil {
		return nilField(key)
	}
	return String(key, *val)
}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) Field {
	return Uint64(key, uint64(val))
}

// Uintp constructs a field that carries a *uint. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintp(key string, val *uint) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint(key, *val)
}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uint64p constructs a field that carries a *uint64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint64p(key string, val *uint64) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint64(key, *val)
}

// Uint32 constructs a field with the given key and value.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(val)}
}

// Uint32p constructs a field that carries a *uint32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint32p(key string, val *uint32) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint32(key, *val)
}

// Uint16 constructs a field with the given key and value.
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(val)}
}

// Uint16p constructs a field that carries a *uint16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint16p(key string, val *uint16) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint16(key, *val)
}

// Uint8 constructs a field with the given key and value.
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: zapcore.Uint8Type, Integer: int64(val)}
}

// Uint8p constructs a field that carries a *uint8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint8p(key string, val *uint8) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint8(key, *val)
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) Field {
	return Field{Key: key, Type: zapcore.UintptrType, Integer: int64(val)}
}

// Uintptrp constructs a field that carries a *uintptr. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintptrp(key string, val *uintptr) Field {
	if val == nil {
		return nilField(key)
	}
	return Uintptr(key, *val)
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy. Outside tests, Any is always a better choice.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func Reflect(key string, val interface{}) Field {
	return Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

// Namespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func Namespace(key string) Field {
	return Field{Key: key, Type: zapcore.NamespaceType}
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Type: zapcore.StringerType, Interface: val}
}

// Time constructs a Field with the given key and value. The encoder
// controls how the time is serialized.
func Time(key string, val time.Time) Field {
	if val.Before(_minTimeInt64) || val.After(_maxTimeInt64) {
		return Field{Key: key, Type: zapcore.TimeFullType, Interface: val}
	}
	return Field{Key: key, Type: zapcore.TimeType, Integer: val.UnixNano(), Interface: val.Location()}
}

// Timep constructs a field that carries a *time.Time. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Timep(key string, val *time.Time) Field {
	if val == nil {
		return nilField(key)
	}
	return Time(key, *val)
}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under provided key. Keep in mind that taking a stacktrace is eager and
// expensive (relatively speaking); this function both makes an allocation and
// takes about two microseconds.
func Stack(key string) Field {
	return StackSkip(key, 1) // skip Stack
}

// StackSkip constructs a field similarly to Stack, but also skips the given
// number of frames from the top of the stacktrace.
func StackSkip(key string, skip int) Field {
	// Returning the stacktrace as a string costs an allocation, but saves us
	// from expanding the zapcore.Field union struct to include a byte slice. Since
	// taking a stacktrace is already so expensive (~10us), the extra allocation
	// is okay.
	return String(key, takeStacktrace(skip+1)) // skip StackSkip
}

// Duration constructs a field with the given key and value. The encoder
// controls how the duration is serialized.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: zapcore.DurationType, Integer: int64(val)}
}

// Durationp constructs a field that carries a *time.Duration. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Durationp(key string, val *time.Duration) Field {
	if val == nil {
		return nilField(key)
	}
	return Duration(key, *val)
}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func Object(key string, val zapcore.ObjectMarshaler) Field {
	return Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: val}
}

// Inline constructs a Field that is similar to Object, but it
// will add the elements of the provided ObjectMarshaler to the
// current namespace.
func Inline(val zapcore.ObjectMarshaler) Field {
	return zapcore.Field{
		Type:      zapcore.InlineMarshalerType,
		Interface: val,
	}
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
//
// Since byte/uint8 and rune/int32 are aliases, Any can't differentiate between
// them. To minimize surprises, []byte values are treated as binary blobs, byte
// values are treated as uint8, and runes are always treated as integers.
func Any(key string, value interface{}) Field {
	switch val := value.(type) {
	case zapcore.ObjectMarshaler:
		return Object(key, val)
	case zapcore.ArrayMarshaler:
		return Array(key, val)
	case bool:
		return Bool(key, val)
	case *bool:
		return Boolp(key, val)
	case []bool:
		return Bools(key, val)
	case complex128:
		return Complex128(key, val)
	case *complex128:
		return Complex128p(key, val)
	case []complex128:
		return Complex128s(key, val)
	case complex64:
		return Complex64(key, val)
	case *complex64:
		return Complex64p(key, val)
	case []complex64:
		return Complex64s(key, val)
	case float64:
		return Float64(key, val)
	case *float64:
		return Float64p(key, val)
	case []float64:
		return Float64s(key, val)
	case float32:
		return Float32(key, val)
	case *float32:
		return Float32p(key, val)
	case []float32:
		return Float32s(key, val)
	case int:
		return Int(key, val)
	case *int:
		return Intp(key, val)
	case []int:
		return Ints(key, val)
	case int64:
		return Int64(key, val)
	case *int64:
		return Int64p(key, val)
	case []int64:
		return Int64s(key, val)
	case int32:
		return Int32(key, val)
	case *int32:
		return Int32p(key, val)
	case []int32:
		return Int32s(key, val)
	case int16:
		return Int16(key, val)
	case *int16:
		return Int16p(key, val)
	case []int16:
		return Int16s(key, val)
	case int8:
		return Int8(key, val)
	case *int8:
		return Int8p(key, val)
	case []int8:
		return Int8s(key, val)
	case string:
		return String(key, val)
	case *string:
		return Stringp(key, val)
	case []string:
		return Strings(key, val)
	case uint:
		return Uint(key, val)
	case *uint:
		return Uintp(key, val)
	case []uint:
		return Uints(key, val)
	case uint64:
		return Uint64(key, val)
	case *uint64:
		return Uint64p(key, val)
	case []uint64:
		return Uint64s(key, val)
	case uint32:
		return Uint32(key, val)
	case *uint32:
		return Uint32p(key, val)
	case []uint32:
		return Uint32s(key, val)
	case uint16:
		return Uint16(key, val)
	case *uint16:
		return Uint16p(key, val)
	case []uint16:
		return Uint16s(key, val)
	case uint8:
		return Uint8(key, val)
	case *uint8:
		return Uint8p(key, val)
	case []byte:
		return Binary(key, val)
	case uintptr:
		return Uintptr(key, val)
	case *uintptr:
		return Uintptrp(key, val)
	case []uintptr:
		return Uintptrs(key, val)
	case time.Time:
		return Time(key, val)
	case *time.Time:
		return Timep(key, val)
	case []time.Time:
		return Times(key, val)
	case time.Duration:
		return Duration(key, val)
	case *time.Duration:
		return Durationp(key, val)
	case []time.Duration:
		return Durations(key, val)
	case error:
		return NamedError(key, val)
	case []error:
		return Errors(key, val)
	case fmt.Stringer:
		return Stringer(key, val)
	default:
		return Reflect(key, val)
	}
}

// Array constructs a field with the given key and ArrayMarshaler. It provides
// a flexible, but still type-safe and efficient, way to add array-like types
// to the logging context. The struct's MarshalLogArray method is called lazily.
func Array(key string, val zapcore.ArrayMarshaler) Field {
	return Field{Key: key, Type: zapcore.ArrayMarshalerType, Interface: val}
}

// Bools constructs a field that carries a slice of bools.
func Bools(key string, bs []bool) Field {
	return Array(key, bools(bs))
}

// ByteStrings constructs a field that carries a slice of []byte, each of which
// must be UTF-8 encoded text.
func ByteStrings(key string, bss [][]byte) Field {
	return Array(key, byteStringsArray(bss))
}

// Complex128s constructs a field that carries a slice of complex numbers.
func Complex128s(key string, nums []complex128) Field {
	return Array(key, complex128s(nums))
}

// Complex64s constructs a field that carries a slice of complex numbers.
func Complex64s(key string, nums []complex64) Field {
	return Array(key, complex64s(nums))
}

// Durations constructs a field that carries a slice of time.Durations.
func Durations(key string, ds []time.Duration) Field {
	return Array(key, durations(ds))
}

// Float64s constructs a field that carries a slice of floats.
func Float64s(key string, nums []float64) Field {
	return Array(key, float64s(nums))
}

// Float32s constructs a field that carries a slice of floats.
func Float32s(key string, nums []float32) Field {
	return Array(key, float32s(nums))
}

// Ints constructs a field that carries a slice of integers.
func Ints(key string, nums []int) Field {
	return Array(key, ints(nums))
}

// Int64s constructs a field that carries a slice of integers.
func Int64s(key string, nums []int64) Field {
	return Array(key, int64s(nums))
}

// Int32s constructs a field that carries a slice of integers.
func Int32s(key string, nums []int32) Field {
	return Array(key, int32s(nums))
}

// Int16s constructs a field that carries a slice of integers.
func Int16s(key string, nums []int16) Field {
	return Array(key, int16s(nums))
}

// Int8s constructs a field that carries a slice of integers.
func Int8s(key string, nums []int8) Field {
	return Array(key, int8s(nums))
}

// Strings constructs a field that carries a slice of strings.
func Strings(key string, ss []string) Field {
	return Array(key, stringArray(ss))
}

// Times constructs a field that carries a slice of time.Times.
func Times(key string, ts []time.Time) Field {
	return Array(key, times(ts))
}

// Uints constructs a field that carries a slice of unsigned integers.
func Uints(key string, nums []uint) Field {
	return Array(key, uints(nums))
}

// Uint64s constructs a field that carries a slice of unsigned integers.
func Uint64s(key string, nums []uint64) Field {
	return Array(key, uint64s(nums))
}

// Uint32s constructs a field that carries a slice of unsigned integers.
func Uint32s(key string, nums []uint32) Field {
	return Array(key, uint32s(nums))
}

// Uint16s constructs a field that carries a slice of unsigned integers.
func Uint16s(key string, nums []uint16) Field {
	return Array(key, uint16s(nums))
}

// Uint8s constructs a field that carries a slice of unsigned integers.
func Uint8s(key string, nums []uint8) Field {
	return Array(key, uint8s(nums))
}

// Uintptrs constructs a field that carries a slice of pointer addresses.
func Uintptrs(key string, us []uintptr) Field {
	return Array(key, uintptrs(us))
}

// Errors constructs a field that carries a slice of errors.
func Errors(key string, errs []error) Field {
	return Array(key, errArray(errs))
}

type bools []bool

func (bs bools) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range bs {
		arr.AppendBool(bs[i])
	}
	return nil
}

type byteStringsArray [][]byte

func (bss byteStringsArray) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range bss {
		arr.AppendByteString(bss[i])
	}
	return nil
}

type complex128s []complex128

func (nums complex128s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendComplex128(nums[i])
	}
	return nil
}

type complex64s []complex64

func (nums complex64s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendComplex64(nums[i])
	}
	return nil
}

type durations []time.Duration

func (ds durations) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range ds {
		arr.AppendDuration(ds[i])
	}
	return nil
}

type float64s []float64

func (nums float64s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendFloat64(nums[i])
	}
	return nil
}

type float32s []float32

func (nums float32s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendFloat32(nums[i])
	}
	return nil
}

type ints []int

func (nums ints) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendInt(nums[i])
	}
	return nil
}

type int64s []int64

func (nums int64s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendInt64(nums[i])
	}
	return nil
}

type int32s []int32

func (nums int32s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendInt32(nums[i])
	}
	return nil
}

type int16s []int16

func (nums int16s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendInt16(nums[i])
	}
	return nil
}

type int8s []int8

func (nums int8s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendInt8(nums[i])
	}
	return nil
}

type stringArray []string

func (ss stringArray) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range ss {
		arr.AppendString(ss[i])
	}
	return nil
}

type times []time.Time

func (ts times) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range ts {
		arr.AppendTime(ts[i])
	}
	return nil
}

type uints []uint

func (nums uints) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUint(nums[i])
	}
	return nil
}

type uint64s []uint64

func (nums uint64s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUint64(nums[i])
	}
	return nil
}

type uint32s []uint32

func (nums uint32s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUint32(nums[i])
	}
	return nil
}

type uint16s []uint16

func (nums uint16s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUint16(nums[i])
	}
	return nil
}

type uint8s []uint8

func (nums uint8s) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUint8(nums[i])
	}
	return nil
}

type uintptrs []uintptr

func (nums uintptrs) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendUintptr(nums[i])
	}
	return nil
}

// Error is shorthand for the common idiom NamedError("error", err).
func Error(err error) Field {
	return NamedError("error", err)
}

// NamedError constructs a field that lazily stores err.Error() under the
// provided key. Errors which also implement fmt.Formatter (like those produced
// by github.com/pkg/errors) will also have their verbose representation stored
// under key+"Verbose". If passed a nil error, the field is a no-op.
//
// For the common case in which the key is simply "error", the Error function
// is shorter and less repetitive.
func NamedError(key string, err error) Field {
	if err == nil {
		return Skip()
	}
	return Field{Key: key, Type: zapcore.ErrorType, Interface: err}
}

var _errArrayElemPool = sync.Pool{New: func() interface{} {
	return &errArrayElem{}
}}

type errArray []error

func (errs errArray) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range errs {
		if errs[i] == nil {
			continue
		}
		// To represent each error as an object with an "error" attribute and
		// potentially an "errorVerbose" attribute, we need to wrap it in a
		// type that implements LogObjectMarshaler. To prevent this from
		// allocating, pool the wrapper type.
		elem := _errArrayElemPool.Get().(*errArrayElem)
		elem.error = errs[i]
		arr.AppendObject(elem)
		elem.error = nil
		_errArrayElemPool.Put(elem)
	}
	return nil
}

type errArrayElem struct {
	error
}

func (e *errArrayElem) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	// Re-use the error field's logic, which supports non-standard error types.
	Error(e.error).AddTo(enc)
	return nil
}

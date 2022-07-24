package rs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML handles parsing of rendering suffix and normal yaml unmarshaling
func (f *BaseField) UnmarshalYAML(n *yaml.Node) (err error) {
	if !f.initialized() {
		return fmt.Errorf("rs: struct not intialized before unmarshaling")
	}

	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("rs: unexpected non map data %q for struct %q unmarshaling",
			n.Tag, f._parentValue.Type().String(),
		)
	}

	oneLevelMap, err := unmarshalYamlMap(n.Content)
	if err != nil {
		return fmt.Errorf("rs: data unmarshal failed for %s: %w",
			f._parentValue.Type().String(), err,
		)
	}

	var (
		hasVirtualKey = false
		hasField      bool

		suffixStart int
		field       fieldRef

		rawYamlKey, rsTag, yamlKey string
	)
	// set values
	for _, kv := range oneLevelMap {
		rawYamlKey = kv[0].Value

		// custom tag with `!rs:` prefix can also indicate rendering suffix

		suffixStart = strings.LastIndexByte(rawYamlKey, '@')
		field, hasField = f.getField(rawYamlKey)
		rsTag = strings.TrimPrefix(strings.TrimPrefix(kv[1].Tag, "!rs:"), "!tag:arhat.dev/rs:")

		switch {
		case rsTag != kv[1].Tag:
			// this field is definitely using rendering suffix as
			// indicated by its tag, which can only be set manually
			// in yaml
			//
			// In this case, user MUST NOT set @ style rendering suffix
			// at the same time

			kv[1].Tag = ""
			yamlKey = rawYamlKey
			if yamlKey == "__" {
				hasVirtualKey = true
			}
			err = f.unmarshalRS(yamlKey, rsTag, kv)
		case hasField,
			suffixStart == -1:
			// matched struct field tag `yaml:"foo@http"`
			// or having no rendering suffix
			//
			// this field is definitely not using rendering suffix

			yamlKey = rawYamlKey

			// but when there is virtual key for this object
			// we need to postpone the unmarshaling of this field
			// so its value won't be overridden by value resolved from
			// virtual key
			if hasVirtualKey {
				err = f.unmarshalRS(yamlKey, "", kv)
			} else {
				var ref *fieldRef
				if hasField {
					ref = &field
				}

				err = f.unmarshalNoRS(yamlKey, kv, ref)
			}

		default:
			// suffixStart != -1 (always true)
			// has rendering suffix suffix, and no field tag with exact match

			yamlKey = rawYamlKey[:suffixStart]
			if yamlKey == "__" {
				hasVirtualKey = true
			}
			err = f.unmarshalRS(yamlKey, rawYamlKey[suffixStart+1:], kv)
		}

		if err != nil {
			return
		}
	}

	return
}

func (f *BaseField) unmarshalNoRS(yamlKey string, kv *[2]*yaml.Node, field *fieldRef) (err error) {
	v := kv[1]
	if field == nil {
		field = f.inlineMap
		var fm yaml.Node
		fakeMap(&fm, kv[0], kv[1])
		v = &fm
	}

	if field == nil {
		if f._opts != nil && f._opts.AllowUnknownFields {
			// allows unknown fields
			return
		}

		err = fmt.Errorf("rs: unknown yaml field %q to %s",
			yamlKey, f._parentValue.Type().String(),
		)
		return
	}

	// field is not nil, fill value into inline map
	//
	// it's safe to provide nil RenderingHandler as we don't have any rendering suffix
	err = unmarshal(v, field, &field.isInlineMap, nil, yamlKey, nil /* rendering handler */)
	if err != nil {
		err = fmt.Errorf("rs: unmarshal yaml field %q to type %s.%s: %w",
			yamlKey, f._parentValue.Type().String(), field.fieldName, err,
		)
		return
	}

	return
}

func (f *BaseField) unmarshalRS(yamlKey, suffix string, kv *[2]*yaml.Node) (err error) {
	var ref *fieldRef

	field, hasField := f.getField(yamlKey)
	if hasField {
		ref = &field
	}

	v := prepareYamlNode(kv[1])
	if v == nil {
		v = kv[1]
	}

	if ref == nil {
		if yamlKey == "__" {
			// handle virtual key
			return f.addUnresolvedField_self(suffix, kv[1])
		}

		var (
			fm        yaml.Node
			clonedKey yaml.Node
		)

		cloneYamlNode(&clonedKey, kv[0], strTag, yamlKey)
		fakeMap(&fm, &clonedKey, v)
		v = &fm
		ref = f.inlineMap
	}

	if ref == nil {
		if f._opts != nil && f._opts.AllowUnknownFields {
			return nil
		}

		err = fmt.Errorf("rs: unknown yaml key %q for type %s",
			yamlKey, f._parentValue.Type().String(),
		)
		return
	}

	// field is not nil

	if ref.disableRS {
		err = fmt.Errorf("rendering suffix is not allowed to %q (type %s)",
			yamlKey, f._parentValue.Type().String(),
		)
		return
	}

	return ref.base.addUnresolvedField(ref, v, yamlKey, suffix, nil)
}

func unmarshal(
	in *yaml.Node,
	out *fieldRef,
	keepOld *bool,

	// parent of the incoming yaml node, intended to support
	// virtual key `__` for document and sequence item
	//
	// only effective if parent is nil or SequenceNode
	parent *yaml.Node,
	yamlKey string,

	// rc is used to resolve virtual key for rendered data (only happening during ResolveFileds)
	// so it can be nil when unmarhaling fields without rendering suffix
	rc RenderingHandler,
) (err error) {
	if !out.fieldValue.IsValid() {
		// no way to know what value we can set
		// NOTE: this should not happen unless user called yaml.Unmarshal with
		// 	     something nil as out
		err = fmt.Errorf(
			"invalid nil unmarshal target for yaml key %q",
			yamlKey,
		)
		return
	}

	outKind := out.fieldValue.Kind()

	in = prepareYamlNode(in)
	// reset to zero value if already set
	if isEmpty(in) {
		// only clear out ptr value to keep same behavior as
		// yaml.Unmarshal
		if out.fieldValue.CanSet() {
			switch outKind {
			case reflect.Ptr, reflect.Map, reflect.Slice:
				out.fieldValue.Set(reflect.Zero(out.fieldValue.Type()))
			}
		}
		return nil
	}

	// we are trying to set value of it, so initialize the pointer when not set before
	var outElem fieldRef
	for outKind == reflect.Ptr {
		if out.fieldValue.IsNil() {
			out.fieldValue.Set(reflect.New(out.fieldValue.Type().Elem()))
		}

		outElem = out.Elem()
		outKind = outElem.fieldValue.Kind()
		out = &outElem
	}

	// handle virtual key `__` for document node and sequence node
	//
	// mapping node should use rendering suffix directly and they can handle
	// virtual key on their own
	if (parent == nil || parent.Kind == yaml.SequenceNode) &&
		in.Kind == yaml.MappingNode &&
		// resolving inline map item
		// it should be able to handle virtual key on its own
		!out.isInlineMap {

		var pairs []*[2]*yaml.Node
		pairs, err = unmarshalYamlMap(in.Content)
		if err != nil {
			err = fmt.Errorf("invalid mapping node: %w", err)
			return
		}

		// TODO: merge multiple virtual values into one
		var (
			content []*yaml.Node
			suffix  string
			ufs     unresolvedFieldSpec
		)

		for _, pair := range pairs {
			suffix = strings.TrimPrefix(pair[0].Value, "__@")

			if suffix == pair[0].Value {
				content = append(content, pair[:]...)
				continue
			}

			if rc == nil {
				return fmt.Errorf("invalid list item using virtual key, "+
					"please add rendering suffix to your list field name %q", yamlKey,
				)
			}

			ufs = unresolvedFieldSpec{
				ref:       out,
				rawData:   pair[1],
				renderers: parseRenderingSuffix(suffix),
			}

			_, err = handleUnresolvedField(1, &ufs, nil, true, yamlKey, rc)
			if err != nil {
				return err
			}
		}

		if len(content) == 0 {
			switch outKind {
			case reflect.Map, reflect.Interface, reflect.Struct:
			default:
				return nil
			}
		}

		// temporarily replace content
		original := in.Content
		defer func() {
			in.Content = original
		}()

		in.Content = content
	}

	switch outKind {
	case reflect.Invalid:
		// TODO: this should not happen since we have already checked outVal.IsValid before

		// unreachable code
		err = fmt.Errorf("unexpected nil out value for yaml key %q", yamlKey)
		return
	case reflect.Chan, reflect.Func:
		err = fmt.Errorf("invalid out value is not data type for yaml key %q", yamlKey)
		return
	case reflect.Array:
		return unmarshalArray(in, out, yamlKey, rc)
	case reflect.Slice:
		return unmarshalSlice(in, out, keepOld, yamlKey, rc)
	case reflect.Map:
		_, err = unmarshalMap(in, out, nil, keepOld, yamlKey, rc)
		return
	case reflect.Struct:
		return unmarshalStruct(in, out, yamlKey)
	case reflect.Interface:
		handled, err := unmarshalInterface(in, out, keepOld, yamlKey, rc)
		if !handled {
			// fallback to go-yaml behavior
			return in.Decode(out.fieldValue.Addr().Interface())
		}

		return err
	default:
		return in.Decode(out.fieldValue.Addr().Interface())
	}
}

func unmarshalStruct(
	in *yaml.Node,
	outVal *fieldRef,
	yamlKey string,
) (err error) {
	tryInit(outVal.fieldValue, outVal.base._opts)

	var (
		out = outVal.fieldValue.Addr().Interface()
	)

	switch ot := out.(type) {
	case yaml.Unmarshaler:
		err = ot.UnmarshalYAML(in)
	default:
		err = in.Decode(ot)
	}

	if in.Kind == yaml.MappingNode {
		return
	}

	if isEmpty(in) {
		return nil
	}

	in, err = TypeHintNone{}.apply(in)
	if err != nil {
		return fmt.Errorf(
			"unexpected input for struct %q %s: %w",
			yamlKey, outVal.fieldValue.Type().String(), err,
		)
	}

	switch ot := out.(type) {
	case yaml.Unmarshaler:
		return ot.UnmarshalYAML(in)
	default:
		return in.Decode(ot)
	}
}

// unmarshalInterface handles interface type creation
func unmarshalInterface(
	in *yaml.Node,
	out *fieldRef,
	keepOld *bool,
	yamlKey string,
	rc RenderingHandler,
) (bool, error) {
	opts := out.base._opts
	if opts == nil || opts.InterfaceTypeHandler == nil {
		// there is no interface type handler
		// use default behavior for interface{} types
		return false, nil
	}

	val := out.fieldValue
	if !out.fieldValue.IsValid() || out.fieldValue.IsNil() {
		fVal, err := opts.InterfaceTypeHandler.Create(out.fieldValue.Type(), yamlKey)
		if err != nil {
			if errors.Is(err, ErrInterfaceTypeNotHandled) && out.fieldValue.Type() == typeEface_Any {
				// no type information provided, decode using go-yaml directly
				return false, nil
			}

			return true, fmt.Errorf(
				"create interface field: %w",
				err,
			)
		}

		val = reflect.ValueOf(fVal)
		if err := checkAssignable(yamlKey, val, out.fieldValue); err != nil {
			return true, err
		}

		if out.fieldValue.CanSet() {
			out.fieldValue.Set(val)
		} else {
			out.fieldValue.Elem().Set(val)
		}
	} else {
		val = val.Elem()
	}

	// DO NOT use outVal directly, which will always match reflect.Interface
	clonedOut := out.clone(val)
	return true, unmarshal(in, &clonedOut, keepOld, in, yamlKey, rc)
}

func unmarshalArray(
	in *yaml.Node,
	out *fieldRef,
	yamlKey string,
	rc RenderingHandler,
) (err error) {
	if in.Kind != yaml.SequenceNode {
		in, err = applyObjectsHint(in)
		if err != nil {
			err = fmt.Errorf(
				"unexpected input for array %q (%s): %w",
				yamlKey, out.fieldValue.Type().String(), err,
			)
			return
		}
	}

	size := len(in.Content)
	expectedSize := out.fieldValue.Len()
	if size < expectedSize {
		return fmt.Errorf(
			"array size not match for %s: want %d got %d",
			out.fieldValue.Type().String(), expectedSize, size,
		)
	}

	var clonedOut fieldRef
	for i := 0; i < expectedSize; i++ {
		clonedOut = out.clone(out.fieldValue.Index(i))
		err = unmarshal(
			in.Content[i], &clonedOut,
			// always drop existing inner data
			// (actually doesn't matter since it's new)
			nil,
			in,
			yamlKey,
			rc,
		)
		if err != nil {
			err = fmt.Errorf(
				"unmarshal #%d array item of yaml field %q for %s: %w",
				i, yamlKey, out.fieldValue.Type().String(), err,
			)
			return
		}
	}

	return nil
}

// when keepOld is set to true, append data from in to original slice
func unmarshalSlice(
	in *yaml.Node,
	outVal *fieldRef,
	keepOld *bool,
	yamlKey string,
	rc RenderingHandler,
) (err error) {
	if in.Kind != yaml.SequenceNode {
		if isEmpty(in) {
			return nil
		}

		in, err = applyObjectsHint(in)
		if err != nil {
			err = fmt.Errorf(
				"unexpected input for slice %q (%s): %w",
				yamlKey, outVal.fieldValue.Type().String(), err,
			)

			return
		}
	}

	size := len(in.Content)
	tmpVal := reflect.MakeSlice(outVal.fieldValue.Type(), size, size)

	var clonedOut fieldRef
	for i := 0; i < size; i++ {
		clonedOut = outVal.clone(tmpVal.Index(i))
		err = unmarshal(
			in.Content[i], &clonedOut,
			// always drop existing inner data
			// (actually doesn't matter since it's new)
			nil,
			in,
			yamlKey,
			rc,
		)
		if err != nil {
			return fmt.Errorf(
				"unmarshal #%d slice item of yaml field %q for %s: %w",
				i, yamlKey, outVal.fieldValue.Type().String(), err,
			)
		}
	}

	err = checkAssignable(yamlKey, tmpVal, outVal.fieldValue)
	if err != nil {
		return
	}

	if outVal.fieldValue.IsZero() || keepOld == nil || !*keepOld {
		outVal.fieldValue.Set(tmpVal)
	} else {
		outVal.fieldValue.Set(reflect.AppendSlice(outVal.fieldValue, tmpVal))
	}

	return
}

// map key MUST be string
// the return value is only meaningful when keepOld is set (resolving inline map pair)
func unmarshalMap(
	in *yaml.Node,
	outVal *fieldRef,
	inlineMapItemCache *reflect.Value,
	keepOld *bool,
	yamlKey string,
	rc RenderingHandler,
) (ret *reflect.Value, err error) {
	if in.Kind != yaml.MappingNode {
		in, err = applyObjectHint(in)
		if err != nil {
			err = fmt.Errorf(
				"unexpected input for map %q (%s): %w",
				yamlKey, outVal.fieldValue.Type().String(), err,
			)
			return
		}
	}

	// keepOld can only be set when resolving inline map items
	// then the `in` arg only contain a pair of key value
	//
	// when keepOld is set, there can be some data already unmarshaled
	// inside the map, so we should not create a new map in that case
	if outVal.fieldValue.IsNil() || keepOld == nil {
		outVal.fieldValue.Set(reflect.MakeMap(outVal.fieldValue.Type()))
	}

	m, err := unmarshalYamlMap(in.Content)
	if err != nil {
		return
	}

	valType := outVal.fieldValue.Type().Elem()
	var (
		k         string
		clonedOut fieldRef
	)
	for i, kv := range m {
		if i == 0 && keepOld != nil && *keepOld && inlineMapItemCache != nil {
			ret = inlineMapItemCache
		} else {
			val := reflect.New(valType).Elem()
			ret = &val
		}

		k = kv[0].Value
		clonedOut = outVal.clone(*ret)
		err = unmarshal(
			kv[1], &clonedOut, keepOld, in,
			// use k rather than `yamlKey`
			k,
			rc,
		)
		if err != nil {
			err = fmt.Errorf("unmarshal map value %s for key %q: %w",
				valType.String(), k, err,
			)

			return
		}

		outVal.fieldValue.SetMapIndex(reflect.ValueOf(k), *ret)
	}

	return
}

func checkAssignable(yamlKey string, in, out reflect.Value) (err error) {
	if !in.Type().AssignableTo(out.Type()) {
		err = fmt.Errorf(
			"unexpected value of yaml field %q: want %q, got %q",
			yamlKey, out.Type().String(), in.Type().String(),
		)
		return
	}

	return
}

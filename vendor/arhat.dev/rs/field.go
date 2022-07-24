package rs

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

// BaseField is a helper struct type implementing Field interface
// it serves as a cache storage for encoding/decoding
type BaseField struct {
	// fields with `_` prefix are supposed to be set after parent struct going through InitAny()

	_initialized uint32

	// _parentType and _parentValue represents the parent struct type and value
	// they are set in Init function call
	// _parentType  reflect.Type
	_parentValue reflect.Value

	_opts *Options

	// normalFields are those without `rs:"other"` field tag
	// key: first value of the specified field tag namespace
	// 		or lower case version of the exported field name (go-yaml's default behavior)
	normalFields map[string]fieldRef
	inlineMap    *fieldRef

	// key: yamlKey
	unresolvedNormalFields   map[string]unresolvedFieldSpec
	unresolvedInlineMapItems map[string][]unresolvedFieldSpec

	unresolvedSelfItems []unresolvedFieldSpec
}

func (f *BaseField) initialized() bool {
	return atomic.LoadUint32(&f._initialized) == 1
}

type tagSpec struct {
	// yamlKey is empty when field is unexported or ignored
	yamlKey string

	inline    bool
	omitempty bool
	inlineMap bool
	disableRS bool
}

// parseFieldTags
//
// return value will be nil if the field is unexported or ignored by its data tag (e.g. `yaml:"-"`)
func (f *BaseField) parseFieldTags(sf *reflect.StructField, dataTagNS string) (ret tagSpec, err error) {
	if len(sf.PkgPath) != 0 {
		// unexported
		return
	}

	yTags := strings.Split(sf.Tag.Get(dataTagNS), ",")
	yamlKey := yTags[0]

	if yamlKey == "-" {
		// ignored explicitly
		return
	}

	if len(yamlKey) == 0 {
		yamlKey = strings.ToLower(sf.Name)
	}

	ret.yamlKey = yamlKey

	for _, t := range yTags[1:] {
		switch t {
		case "inline":
			if sf.Type.Kind() != reflect.Map {
				ret.inline = true
				continue
			}

			// inline map, MUST have string key
			if sf.Type.Key() != type_string {
				err = fmt.Errorf(
					"inline option not applicable to %s.%s: "+
						"inline map MUST have string key",
					f._parentValue.Type().String(), sf.Name,
				)
				return
			}

			// inline map is equivalent to `rs:"other"`
			ret.inlineMap = true
		case "omitempty":
			ret.omitempty = true
		default:
			// TBD: currently we just ignore unknown tag, shall we panic?
		}
	}

	// rs tag is used to extend yaml tag
	for _, t := range strings.Split(sf.Tag.Get(TagNameRS), ",") {
		switch t {
		case "other":
			// other is used to match unhandled values
			// only supports map[string]Any
			ret.inlineMap = true
		case "":
		case "disabled":
			ret.disableRS = true
		default:
			err = fmt.Errorf(
				"unknown rs tag value %q for %s.%s",
				t, f._parentValue.Type().String(), sf.Name,
			)
			return
		}
	}

	return ret, nil
}

func (f *BaseField) init(
	parentVal reflect.Value,
	opts *Options,
) error {
	if !atomic.CompareAndSwapUint32(&f._initialized, 0, 1) {
		// already initialized
		return nil
	}

	f._parentValue = parentVal
	parentType := parentVal.Type()
	f._opts = opts

	var dataTagNS string
	if opts != nil && len(opts.DataTagNamespace) != 0 {
		dataTagNS = opts.DataTagNamespace
	} else {
		dataTagNS = "yaml"
	}

	// get known fields for unmarshaling, skip the first field (the BaseField itself)
	for i := 1; i < parentType.NumField(); i++ {
		sf := parentType.Field(i)

		ts, err := f.parseFieldTags(&sf, dataTagNS)
		if err != nil {
			return err
		}

		if len(ts.yamlKey) == 0 {
			continue
		}

		fieldValue := f._parentValue.Field(i)

		// initialize struct fields accepted by Init(), in case being used later
		// DO NOT USE tryInit, that will only init current field, which will cause
		// error when user try to resolve data not unmarshaled from yaml
		InitRecursively(fieldValue, opts)

		if !ts.inline {
			if !f.addField(f, sf.Name, ts, fieldValue) {
				return fmt.Errorf(
					"duplicate yaml key %q in struct %s.%s",
					ts.yamlKey, parentType.String(), sf.Name,
				)
			}

			continue
		}

		// handle inline fields

		err = f.collectInlineFields(&sf, fieldValue, dataTagNS)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *BaseField) collectInlineFields(
	sf *reflect.StructField,
	fieldValue reflect.Value,
	dataTagNS string,
) error {
	// try to find BaseField in inline field
	// if exists, let it manage dynamnic fields
	var innerBaseF reflect.Value
	switch kind := sf.Type.Kind(); {
	case kind == reflect.Struct:
		// embedded struct
		if reflect.PtrTo(sf.Type).Implements(typeEface_Field) || sf.Type.Implements(typeEface_Field) {
			innerBaseF = fieldValue.Field(0)
		}
	case kind == reflect.Ptr:
		typ := sf.Type
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			return fmt.Errorf(
				"invalid inline tag applied to pointer of non struct %s.%s",
				f._parentValue.Type().String(), sf.Name,
			)
		}

		// TODO: fix BaseField lookup
		// embedded ptr to struct
		// if sf.Type.Implements(fieldInterfaceType) || typ.Implements(fieldInterfaceType) {
		// 	innerBaseF = fieldValue.Elem().Field(0)
		// }
	default:
		// TBD: support interface type inline?
		return fmt.Errorf(
			"invalid inline tag applied to non struct nor struct pointer field %s.%s",
			f._parentValue.Type().String(), sf.Name,
		)
	}

	base := f
	start := 0
	switch innerBaseF.Kind() {
	case reflect.Struct:
		if innerBaseF.Addr().Type() == typePtr_BaseField {
			base = innerBaseF.Addr().Interface().(*BaseField)
			start = 1
		}
	case reflect.Ptr:
		if innerBaseF.Type() == typePtr_BaseField {
			base = innerBaseF.Interface().(*BaseField)
			start = 1
		}
	}

	// using fieldValue.NumField() instead of sf.Type.NumField()
	// to support already initialized interface value
	for i := start; i < fieldValue.NumField(); i++ {
		sf := sf.Type.Field(i)

		ts, err := f.parseFieldTags(&sf, dataTagNS)
		if err != nil {
			return err
		}

		if len(ts.yamlKey) == 0 {
			continue
		}

		if !ts.inline {
			if !f.addField(base, sf.Name, ts, fieldValue.Field(i)) {
				return fmt.Errorf(
					"duplicate yaml key %q in inline field %s.%s",
					ts.yamlKey, f._parentValue.Type().String(), sf.Name,
				)
			}

			continue
		}

		// handle inline fields

		err = f.collectInlineFields(&sf, fieldValue.Field(i), dataTagNS)
		if err != nil {
			return err
		}
	}

	return nil
}

type fieldRef struct {
	base *BaseField

	// tagName is the first value in selected tag namespace
	tagName string

	// fieldName is the struct field name
	fieldName string

	fieldValue reflect.Value

	omitempty bool

	// this field is only set to true for map type field with
	// `rs:"other"` or `yaml:",inline"` struct field tag
	isInlineMap bool

	// disable rendering suffix support
	disableRS bool
}

func (f *fieldRef) Elem() fieldRef {
	return f.clone(f.fieldValue.Elem())
}

func (f *fieldRef) clone(v reflect.Value) (ret fieldRef) {
	ret = *f
	ret.fieldValue = v
	return
}

// addField adds one field identified by its yamlKey
// it may be a catch-other field
func (f *BaseField) addField(
	base *BaseField,
	fieldName string,
	ts tagSpec,
	fieldValue reflect.Value,
) bool {
	if ts.inlineMap {
		if f.inlineMap != nil {
			panic(fmt.Errorf(
				"bad field tags in %s: only one map in the struct can have `rs:\"other\"` or `yaml:\",inline\"` tag",
				f._parentValue.Type().String(),
			))
		}

		f.inlineMap = &fieldRef{
			tagName:    ts.yamlKey,
			fieldName:  fieldName,
			fieldValue: fieldValue,
			base:       base,

			// TODO: handle omitempty for inline map in marshal
			// omitempty:       false,
			isInlineMap: true,
			disableRS:   ts.disableRS,
		}

		return true
	}

	if f.normalFields == nil {
		f.normalFields = make(map[string]fieldRef)
	}

	// handle normal field

	if _, exists := f.normalFields[ts.yamlKey]; exists {
		return false
	}

	f.normalFields[ts.yamlKey] = fieldRef{
		tagName:   ts.yamlKey,
		fieldName: fieldName,

		fieldValue: fieldValue,
		base:       base,

		omitempty:   ts.omitempty,
		isInlineMap: false,
		disableRS:   ts.disableRS,
	}

	return true
}

func (f *BaseField) getField(yamlKey string) (ret fieldRef, ok bool) {
	ret, ok = f.normalFields[yamlKey]
	return
}

type unresolvedFieldSpec struct {
	// fieldName is struct field name when isInlineMapItem = false
	// otherwise it's inline map item key
	ref *fieldRef

	rawData   *yaml.Node
	renderers []rendererSpec
}

// nolint:revive
func (f *BaseField) addUnresolvedField_self(suffix string, n *yaml.Node) error {
	f.unresolvedSelfItems = append(f.unresolvedSelfItems, unresolvedFieldSpec{
		ref: &fieldRef{
			tagName:     "",
			fieldName:   "",
			fieldValue:  f._parentValue,
			base:        f,
			omitempty:   false,
			isInlineMap: false,
			disableRS:   false,
		},

		rawData:   n,
		renderers: parseRenderingSuffix(suffix),
	})

	return nil
}

func (f *BaseField) addUnresolvedField(
	// value part
	ref *fieldRef,
	rawData *yaml.Node,
	// key part
	yamlKey string,
	suffix string,
	resolvedSuffix []rendererSpec,
) error {
	if resolvedSuffix == nil {
		resolvedSuffix = parseRenderingSuffix(suffix)
	}

	if f._opts != nil && f._opts.AllowedRenderers != nil {
		n := len(resolvedSuffix)
		var name string
		for i := 0; i < n; i++ {
			name = resolvedSuffix[i].name
			_, ok := f._opts.AllowedRenderers[name]
			if !ok {
				return fmt.Errorf("renderer %q is not allowed", name)
			}
		}
	}

	if !ref.isInlineMap {
		f.addUnresolvedNormalField(yamlKey, resolvedSuffix, ref, rawData)
		return nil
	}

	if f.unresolvedInlineMapItems == nil {
		f.unresolvedInlineMapItems = make(map[string][]unresolvedFieldSpec)
	}

	f.unresolvedInlineMapItems[yamlKey] = append(f.unresolvedInlineMapItems[yamlKey], unresolvedFieldSpec{
		ref: f.inlineMap,

		rawData:   rawData,
		renderers: resolvedSuffix,
	})

	return nil
}

func (f *BaseField) addUnresolvedNormalField(
	// key part
	yamlKey string,
	renderers []rendererSpec,

	// value part
	ref *fieldRef,
	rawData *yaml.Node,
) {
	if f.unresolvedNormalFields == nil {
		f.unresolvedNormalFields = make(map[string]unresolvedFieldSpec)
	}

	// it can have multiple values only when it's an inline map item
	if old, exists := f.unresolvedNormalFields[yamlKey]; exists {
		old.rawData = rawData
		f.unresolvedNormalFields[yamlKey] = old
		return
	}

	f.unresolvedNormalFields[yamlKey] = unresolvedFieldSpec{
		ref:       ref,
		rawData:   rawData,
		renderers: renderers,
	}
}

type rendererSpec struct {
	name string

	patchSpec bool
	typeHint  TypeHint
}

func parseRenderingSuffix(rs string) (ret []rendererSpec) {
	var (
		i      int
		idx    int
		szPart int
		sz     = len(rs)
		part   string
	)

	for i = 0; i < sz; i += idx + 1 {
		idx = strings.IndexByte(rs[i:], '|')
		if idx == -1 {
			part = rs[i:]
			szPart = sz - i
			idx = szPart
		} else {
			part = rs[i : i+idx]
			szPart = idx
		}

		if szPart == 0 {
			continue
		}

		spec := rendererSpec{
			patchSpec: part[szPart-1] == '!',
		}

		if spec.patchSpec {
			part = part[:szPart-1]
			// size-- // size not used any more
		}

		if idx := strings.LastIndexByte(part, '?'); idx >= 0 {
			// TODO: do we really want to panic when type hint is not valid?
			var err error
			spec.typeHint, err = ParseTypeHint(part[idx+1:])
			if err != nil {
				panic(err)
			}

			part = part[:idx]
		}

		spec.name = part

		ret = append(ret, spec)
	}

	return ret
}

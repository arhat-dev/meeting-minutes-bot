package rs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"arhat.dev/pkg/stringhelper"
	"gopkg.in/yaml.v3"
)

func (f *BaseField) HasUnresolvedField() bool {
	return len(f.unresolvedNormalFields) != 0
}

func (f *BaseField) ResolveFields(rc RenderingHandler, depth int, names ...string) (err error) {
	if !f.initialized() {
		err = fmt.Errorf("rs: struct not intialized before resolving")
		return
	}

	if depth == 0 {
		return nil
	}

	if len(f.unresolvedSelfItems) != 0 {
		err = resolveOverlappedItems(1, &f._parentValue, "", f.unresolvedSelfItems, rc)
		if err != nil {
			err = fmt.Errorf("rs: resolve value from virtual key: %w", err)
			return
		}
	}

	if len(names) == 0 {
		// resolve all

		for name, v := range f.normalFields {
			err = v.base.resolveNormalField(depth, &v.fieldValue, name, rc)
			if err != nil {
				err = fmt.Errorf(
					"rs: resolve field %s.%s: %w",
					f._parentValue.Type().String(), v.fieldName, err,
				)
				return
			}
		}

		for k, list := range f.unresolvedInlineMapItems {
			err = resolveOverlappedItems(depth, &f.inlineMap.fieldValue, k, list, rc)
			if err != nil {
				return
			}
		}

		if f.inlineMap != nil {
			// the inline map has been resolved above, so let's go to
			// values of these inline map entries
			err = handleResolvedField(depth, &f.inlineMap.fieldValue, rc)
			return
		}

		return
	}

	// resolve specified fileds by tag names

	var (
		ref fieldRef
		ok  bool
	)
	for _, name := range names {
		ref, ok = f.normalFields[name]
		if !ok {
			// can be targeting inline map field name
			switch {
			case len(name) == 0:
				// resolve itself (added by virtual key `__`)

				err = resolveOverlappedItems(1, &f._parentValue, "", f.unresolvedSelfItems, rc)
				if err != nil {
					return
				}

				continue
			case f.inlineMap != nil && f.inlineMap.fieldName == name:
				for k, list := range f.unresolvedInlineMapItems {
					err = resolveOverlappedItems(depth, &f.inlineMap.fieldValue, k, list, rc)
					if err != nil {
						return
					}
				}

				continue
			default:
				err = fmt.Errorf(
					"rs: no such field %q in struct %q",
					name, f._parentValue.Type().String(),
				)
				return
			}
		}

		err = ref.base.resolveNormalField(depth, &ref.fieldValue, name, rc)
		if err != nil {
			err = fmt.Errorf(
				"rs: resolve requested field %s of %s: %w",
				name, f._parentValue.Type().String(), err,
			)

			return
		}
	}

	return nil
}

func resolveOverlappedItems(
	depth int,
	fVal *reflect.Value,
	k string,
	list []unresolvedFieldSpec,
	rc RenderingHandler,
) error {
	var (
		itemCache *reflect.Value
		err       error
	)

	n := len(list)
	for i := 0; i < n; i++ {
		itemCache, err = handleUnresolvedField(
			depth,
			&list[i],
			itemCache,
			// flush existing data with the same key on first pair
			i != 0,
			k,
			rc,
		)
		if err != nil {
			return err
		}
	}

	return handleResolvedField(depth, fVal, rc)
}

func (f *BaseField) resolveNormalField(
	depth int,
	fieldValue *reflect.Value,
	yamlKey string,
	rc RenderingHandler,
) error {
	v, ok := f.unresolvedNormalFields[yamlKey]
	if !ok {
		return handleResolvedField(depth, fieldValue, rc)
	}

	_, err := handleUnresolvedField(
		depth, &v, nil, false, yamlKey, rc,
	)
	return err
}

func handleResolvedField(
	depth int,
	field *reflect.Value,
	rc RenderingHandler,
) (err error) {
	if depth == 0 {
		return
	}

	switch fk := field.Kind(); fk {
	case reflect.Map:
		if field.IsNil() {
			return
		}

		iter := field.MapRange()
		var val reflect.Value
		for iter.Next() {
			val = iter.Value()
			err = handleResolvedField(depth-1, &val, rc)
			if err != nil {
				return
			}
		}
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if fk == reflect.Slice && field.IsNil() {
			// this is a resolved field, slice/array empty means no value
			return
		}

		var (
			val reflect.Value
			n   = field.Len()
		)

		for i := 0; i < n; i++ {
			val = field.Index(i)
			err = handleResolvedField(depth-1, &val, rc)
			if err != nil {
				return
			}
		}
	case reflect.Struct:
		// handled after switch
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		if !field.IsValid() || field.IsZero() || field.IsNil() {
			return
		}
		// handled after switch
	default:
		// scalar types, no action required
		return
	}

	return tryResolve(depth-1, field, rc)
}

func tryResolve(depth int, targetField *reflect.Value, rc RenderingHandler) error {
	if targetField.CanInterface() {
		fVal, canCallResolve := targetField.Interface().(Field)
		if canCallResolve {
			// fVal can have underlying data nil, but with Field type
			if targetField.IsNil() {
				return nil
			}

			return fVal.ResolveFields(rc, depth)
		}
	}

	if !targetField.CanAddr() {
		return nil
	}

	ptrTargetField := targetField.Addr()
	if !ptrTargetField.CanInterface() {
		return nil
	}

	fVal, canCallResolve := ptrTargetField.Interface().(Field)
	if canCallResolve {
		if fVal == nil || ptrTargetField.IsNil() {
			return nil
		}

		return fVal.ResolveFields(rc, depth)
	}

	// no more field to resolve
	return nil
}

func handleUnresolvedField(
	depth int,
	v *unresolvedFieldSpec,
	inlineMapItemCache *reflect.Value,
	keepOld bool,
	yamlKey string,
	rc RenderingHandler,
) (_ *reflect.Value, err error) {
	target := v.ref

	toResolve := v.rawData
	if target.isInlineMap {
		// unwrap map data for resolving
		toResolve = toResolve.Content[1]
	}

	n := len(v.renderers)
	for i := 0; i < n; i++ {
		toResolve, err = tryRender(
			toResolve,
			&v.renderers[i],
			rc,
		)
		if err != nil {
			err = fmt.Errorf("render value for %q: %w", yamlKey, err)
			return
		}
	}

	resolved := toResolve
	if target.isInlineMap {
		var fm yaml.Node
		fakeMap(&fm, v.rawData.Content[0], resolved)

		inlineMapItemCache, err = unmarshalMap(
			&fm,
			target,
			inlineMapItemCache,
			&keepOld,
			yamlKey,
			rc,
		)
	} else {
		err = unmarshal(resolved, target, nil, nil, yamlKey, rc)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"unmarshal resolved value of %q to field %q: %w",
			yamlKey, target.fieldName, err,
		)
	}

	return inlineMapItemCache, handleResolvedField(depth, &target.fieldValue, rc)
}

func tryRender(
	toResolve *yaml.Node,
	rdr *rendererSpec,
	rc RenderingHandler,
) (_ *yaml.Node, err error) {
	if rdr.patchSpec {
		var (
			patchSpec  *PatchSpec
			patchedObj any
			dataBytes  []byte
		)

		patchSpec, err = resolvePatchSpec(rc, toResolve)
		if err != nil {
			err = fmt.Errorf("invalid patch spec: %w", err)
			return
		}

		patchedObj, err = patchSpec.Apply(rc)
		if err != nil {
			err = fmt.Errorf("apply patch: %w", err)
			return
		}

		// patch doc is generated from arbitrary yaml data from built-in `any` value
		// so we are able to marshal it into json with no error, and then parse the json data
		// as *yaml.Node for further processing
		// TODO: convert patchedObj to *yaml.Node directly

		dataBytes, err = json.Marshal(patchedObj)
		if err != nil {
			err = fmt.Errorf("marshal patched data: %w", err)
			return
		}

		// json data is deemed to be valid yaml, if not, we must have some
		// big problem. so we don't recover yaml.Unmarshal from panic here
		var tmp yaml.Node
		err = yaml.Unmarshal(dataBytes, &tmp)
		if err != nil {
			err = fmt.Errorf("prepare patched data: %w", err)
			return
		}

		toResolve = &tmp
	}

	if len(rdr.name) != 0 {
		var (
			renderedData []byte
			tag          string
		)

		renderedData, err = rc.RenderYaml(rdr.name, toResolve)
		if err != nil {
			err = fmt.Errorf("renderer %q render value: %w", rdr.name, err)
			return
		}

		// check type hinting before assuming it's valid yaml
		//
		// see TestResolve_yaml_unmarshal_invalid_but_no_error in resolve_test.go
		// for reasons this pre-type hint check exists

		// scalar types cannot be applied with patch spec
		// so the rendered data will be the final value for resolving

		var tmp yaml.Node

		switch rdr.typeHint.(type) {
		case TypeHintStr:
			tag = strTag
		case TypeHintInt:
			tag = intTag
		case TypeHintFloat:
			tag = floatTag
		default:
			// assume rendered data as yaml for further processing
			assumeValidYaml(renderedData, &tmp)
		}

		if len(tag) != 0 {
			tmp = yaml.Node{
				Style: guessYamlStringStyle(renderedData),
				Kind:  yaml.ScalarNode,
				Tag:   tag,
				Value: stringhelper.Convert[string, byte](renderedData),
			}
		}

		toResolve = &tmp
	}

	// apply hint after resolving (rendering)
	toResolve, err = applyHint(rdr.typeHint, toResolve)
	if err != nil {
		err = fmt.Errorf("ensure type hint %q: %w", rdr.typeHint, err)
		return
	}

	return toResolve, nil
}

func guessYamlStringStyle(s []byte) yaml.Style {
	switch {
	case bytes.IndexByte(s, '\n') != -1:
		return yaml.LiteralStyle
	default:
		return 0
	}
}

// assumeValidYaml tries its best to unmarshal data as yaml.Node
func assumeValidYaml(data []byte, out *yaml.Node) {
	err := yaml.Unmarshal(data, out)
	if err != nil {
		*out = yaml.Node{
			Style: guessYamlStringStyle(data),
			Kind:  yaml.ScalarNode,
			Tag:   strTag,
			Value: stringhelper.Convert[string, byte](data),
		}

		return
	}

	// unmarshal ok
	if prepared := prepareYamlNode(out); prepared != nil {
		*out = *prepared
	}

	switch {
	case isStrScalar(out), isBinaryScalar(out):
		// use original string instead of yaml unmarshaled string
		// yaml.Unmarshal may modify string content when it's not
		// valid yaml

		*out = yaml.Node{
			Style: guessYamlStringStyle(data),
			Kind:  yaml.ScalarNode,
			Tag:   strTag,
			Value: stringhelper.Convert[string, byte](data),
		}
	}

	return
}

// applyHint applies type hint to yaml.Node, default to TypeHintNone if hint is nil
// return value is set to input if the hinted result is nil
func applyHint(hint TypeHint, in *yaml.Node) (*yaml.Node, error) {
	if hint == nil {
		hint = TypeHintNone{}
	}

	// hintedToResolve can be nil
	hintedToResolve, err := hint.apply(in)
	if err != nil {
		return in, err
	}

	if hintedToResolve != nil {
		return hintedToResolve, nil
	}

	return in, nil
}

// resolve user provided data as patch spec
func resolvePatchSpec(
	rc RenderingHandler,
	toResolve *yaml.Node,
) (ret *PatchSpec, err error) {
	ret = Init(&PatchSpec{}, nil).(*PatchSpec)

	err = toResolve.Decode(ret)
	if err != nil {
		err = fmt.Errorf("decode patch spec: %w", err)
		return
	}

	err = ret.ResolveFields(rc, -1)
	if err != nil {
		err = fmt.Errorf("resolve patch spec: %w", err)
		return
	}

	return
}

func handleOptionalRenderingSuffixResolving(n *yaml.Node, resolve *bool, rc RenderingHandler) (any, error) {
	n = prepareYamlNode(n)
	if n == nil {
		return nil, nil
	}

	if resolve == nil || *resolve {
		any := new(AnyObject)
		err := n.Decode(any)
		if err != nil {
			return nil, err
		}

		err = any.ResolveFields(rc, -1)
		if err != nil {
			return nil, err
		}

		return any.NormalizedValue(), nil
	}

	var ret any
	err := n.Decode(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

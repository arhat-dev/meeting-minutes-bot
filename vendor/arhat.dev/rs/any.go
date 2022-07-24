package rs

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

type anyObjectKind uint8

const (
	_noData anyObjectKind = iota

	_mapData
	_sliceData
	_scalarData

	_unknown
)

// AnyObject is a `any` equivalent with rendering suffix support
type AnyObject struct {
	BaseField `yaml:"-" json:"-"`

	mapData    AnyObjectMap
	sliceData  []AnyObject
	scalarData any

	kind anyObjectKind
}

// NormalizedValue returns underlying value of the AnyObject with primitive types
// that is:
// 	for maps: map[string]any
// 	for slices: []any
//  and primitive types for scalar types
//
// should be called after fields being resolved
func (o *AnyObject) NormalizedValue() any {
	if o == nil {
		return nil
	}

	switch o.kind {
	case _noData:
		return nil
	case _mapData:
		return o.mapData.NormalizedValue()
	case _sliceData:
		n := len(o.sliceData)
		ret := make([]any, n)
		for i := 0; i < n; i++ {
			ret[i] = o.sliceData[i].NormalizedValue()
		}
		return ret
	default:
		return o.scalarData
	}
}

func (o *AnyObject) value() any {
	if o == nil {
		return nil
	}

	switch o.kind {
	case _noData:
		return nil
	case _mapData:
		return &o.mapData
	case _sliceData:
		n := len(o.sliceData)
		ret := make([]*AnyObject, n)
		for i := 0; i < n; i++ {
			ret[i] = &o.sliceData[i]
		}

		return ret
	default:
		return o.scalarData
	}
}

func (o *AnyObject) MarshalYAML() (any, error)    { return o.value(), nil }
func (o *AnyObject) MarshalJSON() ([]byte, error) { return json.Marshal(o.value()) }
func (o *AnyObject) UnmarshalJSON(p []byte) error { return yaml.Unmarshal(p, o) }

func (o *AnyObject) UnmarshalYAML(n *yaml.Node) error {
	if !o.BaseField.initialized() {
		_ = Init(o, nil)
	}

	switch n.Kind {
	case yaml.SequenceNode:
		o.kind = _sliceData
		o.sliceData = make([]AnyObject, len(n.Content))
		for i, vn := range n.Content {
			_ = Init(&o.sliceData[i], o._opts)

			err := prepareYamlNode(vn).Decode(&o.sliceData[i])
			if err != nil {
				return err
			}
		}

		return nil
	case yaml.MappingNode:
		o.kind = _mapData
		pairs, err := unmarshalYamlMap(n.Content)
		if err != nil {
			return err
		}

		var content []*yaml.Node
		for _, pair := range pairs {
			if suffix := strings.TrimPrefix(pair[0].Value, "__@"); suffix != pair[0].Value {
				// is virtual key
				err = o.BaseField.addUnresolvedField_self(suffix, pair[1])
				if err != nil {
					return err
				}

				continue
			}

			content = append(content, pair[0], pair[1])
		}

		if len(content) == 0 {
			return nil
		}

		n.Content = content
		if !o.mapData.initialized() {
			_ = Init(&o.mapData, o._opts)
		}

		return n.Decode(&o.mapData)
	case yaml.ScalarNode:
		o.kind = _scalarData
		switch n.ShortTag() {
		case strTag:
			o.scalarData = n.Value
		case binaryTag:
			o.scalarData = n.Value
		default:
			return n.Decode(&o.scalarData)
		}
		return nil
	default:
		// unreachable (but we accept it as scalar)
		o.kind = _unknown
		return n.Decode(&o.scalarData)
	}
}

func (o *AnyObject) ResolveFields(rc RenderingHandler, depth int, names ...string) error {
	if o == nil {
		return nil
	}

	// resolve self (__@)
	err := o.BaseField.ResolveFields(rc, depth)
	if err != nil {
		return err
	}

	switch o.kind {
	case _mapData:
		if !o.mapData.initialized() {
			Init(&o.mapData, o._opts)
		}

		return o.mapData.ResolveFields(rc, depth, names...)
	case _sliceData:
		n := len(o.sliceData)
		for i := 0; i < n; i++ {
			sd := &o.sliceData[i]
			if !sd.initialized() {
				Init(sd, o._opts)
			}

			err := sd.ResolveFields(rc, depth, names...)
			if err != nil {
				return err
			}
		}

		return nil
	default:
		// scalar type data doesn't need resolving
		return nil
	}
}

// AnyObjectMap is a `map[string]any` equivalent with rendering suffix support
type AnyObjectMap struct {
	BaseField `yaml:"-" json:"-"`

	Data map[string]*AnyObject `rs:"other"`
}

// NormalizedValue returns value of the AnyObjectMap with type map[string]any
func (aom *AnyObjectMap) NormalizedValue() map[string]any {
	if aom.Data == nil {
		return nil
	}

	ret := make(map[string]any, len(aom.Data))
	for k, v := range aom.Data {
		ret[k] = v.NormalizedValue()
	}

	return ret
}

func (aom *AnyObjectMap) MarshalYAML() (any, error)    { return aom.Data, nil }
func (aom *AnyObjectMap) MarshalJSON() ([]byte, error) { return json.Marshal(aom.Data) }
func (aom *AnyObjectMap) UnmarshalJSON(p []byte) error { return yaml.Unmarshal(p, aom) }

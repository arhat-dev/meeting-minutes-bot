package rs

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"arhat.dev/pkg/stringhelper"
	"gopkg.in/yaml.v3"
)

type TypeHint interface {
	apply(*yaml.Node) (*yaml.Node, error)

	String() string
}

type (
	TypeHintNone    struct{}
	TypeHintStr     struct{}
	TypeHintObject  struct{}
	TypeHintObjects struct{}
	TypeHintInt     struct{}
	TypeHintFloat   struct{}
	TypeHintBool    struct{}
)

// nolint:revive
const (
	typeHintName_None    = ""
	typeHintName_Str     = "str"
	typeHintName_Object  = "obj"
	typeHintName_Objects = "[]obj"
	typeHintName_Int     = "int"
	typeHintName_Float   = "float"
	typeHintName_Bool    = "bool"
)

func (TypeHintNone) String() string    { return typeHintName_None }
func (TypeHintStr) String() string     { return typeHintName_Str }
func (TypeHintObject) String() string  { return typeHintName_Object }
func (TypeHintObjects) String() string { return typeHintName_Objects }
func (TypeHintInt) String() string     { return typeHintName_Int }
func (TypeHintFloat) String() string   { return typeHintName_Float }
func (TypeHintBool) String() string    { return typeHintName_Bool }

func (TypeHintNone) apply(v *yaml.Node) (*yaml.Node, error) {
	return prepareYamlNode(v), nil
}

func applyStrHint(n *yaml.Node) (*yaml.Node, error) {
	switch {
	case isStrScalar(n), isBinaryScalar(n):
		// no need to convert
		return n, nil
	case isBoolScalar(n), isFloatScalar(n), isIntScalar(n):
		// TODO: shall we cast null to string directly?
		return castScalarNode(n, strTag)
	}

	data, err := yaml.Marshal(n)
	if err != nil {
		return nil, err
	}

	return &yaml.Node{
		Style: guessYamlStringStyle(data),
		Kind:  yaml.ScalarNode,
		Tag:   strTag,
		Value: strings.TrimSuffix(stringhelper.Convert[string, byte](data), "\n"),
	}, nil
}

func (thi TypeHintStr) apply(n *yaml.Node) (*yaml.Node, error) {
	return applyStrHint(n)
}

func applyObjectHint(n *yaml.Node) (ret *yaml.Node, err error) {
	n = prepareYamlNode(n)
	if isEmpty(n) {
		return nil, nil
	}

	switch n.Kind {
	case yaml.MappingNode:
		return n, nil
	case yaml.ScalarNode:
		// can be string, convert to object
		var dataBytes []byte
		switch {
		case isStrScalar(n):
			dataBytes = stringhelper.ToBytes[byte, byte](n.Value)
		case isBinaryScalar(n):
			dataBytes, err = base64.StdEncoding.DecodeString(n.Value)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf(
				"unsupported non str nor binary scalar node to object conversion",
			)
		}

		var tmp yaml.Node
		defer func() {
			errX := recover()

			if errX != nil {
				err = fmt.Errorf("%v", errX)
			}
		}()
		err = yaml.Unmarshal(dataBytes, &tmp)
		if err != nil {
			return nil, err
		}

		ret = prepareYamlNode(&tmp)
		switch {
		case ret == nil,
			ret.Kind == yaml.MappingNode:
			return
		default:
		}

		fallthrough
	default:
		return nil, fmt.Errorf("cannot convert %q to object", n.Kind)
	}
}

func (tho TypeHintObject) apply(n *yaml.Node) (ret *yaml.Node, err error) {
	return applyObjectHint(n)
}

func applyObjectsHint(n *yaml.Node) (ret *yaml.Node, err error) {
	n = prepareYamlNode(n)
	if n == nil || isEmpty(n) {
		return nil, nil
	}

	switch n.Kind {
	case yaml.SequenceNode:
		return n, nil
	case yaml.ScalarNode:
		// can be string, convert to object
		var dataBytes []byte
		switch {
		case isStrScalar(n):
			dataBytes = stringhelper.ToBytes[byte, byte](n.Value)
		case isBinaryScalar(n):
			dataBytes, err = base64.StdEncoding.DecodeString(n.Value)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf(
				"unsupported non str nor binary scalar node to objects conversion",
			)
		}

		ret = new(yaml.Node)
		defer func() {
			errX := recover()

			if errX != nil {
				err = fmt.Errorf("%v", errX)
			}
		}()

		err = yaml.Unmarshal(dataBytes, ret)
		if err != nil {
			return nil, err
		}

		ret = prepareYamlNode(ret)
		switch {
		case ret == nil,
			ret.Kind == yaml.SequenceNode:
			return
		default:
		}

		fallthrough
	default:
		return nil, fmt.Errorf("cannot convert %q to objects", n.Tag)
	}
}

func (tho TypeHintObjects) apply(n *yaml.Node) (ret *yaml.Node, err error) {
	return applyObjectsHint(n)
}

func (TypeHintFloat) apply(n *yaml.Node) (*yaml.Node, error) {
	return castScalarNode(n, floatTag)
}

func (TypeHintInt) apply(n *yaml.Node) (*yaml.Node, error) {
	return castScalarNode(n, intTag)
}

func (TypeHintBool) apply(n *yaml.Node) (*yaml.Node, error) {
	return castScalarNode(n, boolTag)
}

// cast scalar node directly by changing the tag of it
func castScalarNode(n *yaml.Node, newTag string) (*yaml.Node, error) {
	n = prepareYamlNode(n)
	if n == nil || isEmpty(n) {
		return nil, nil
	}

	if n.Kind != yaml.ScalarNode {
		return nil, fmt.Errorf(
			"unexpected non scalar node for %q", newTag,
		)
	}

	switch {
	case n.Tag == newTag, shortTag(n.Tag) == newTag:
		return n, nil
	}

	var ret yaml.Node
	cloneYamlNode(&ret, n, newTag, n.Value)
	switch newTag {
	case strTag, binaryTag:
	default:
		val, err := strconv.Unquote(n.Value)
		if err != nil {
			// was not quoted
			ret.Value = n.Value
		} else {
			ret.Value = val
		}
	}

	return &ret, nil
}

// cloneYamlNode creates a new yaml.Node by copying all values from n
// and override its tag and value
func cloneYamlNode(out, n *yaml.Node, tag, value string) {
	*out = *n
	out.Tag = tag
	out.Value = value
	return
}

func ParseTypeHint(h string) (TypeHint, error) {
	switch h {
	case typeHintName_None:
		return TypeHintNone{}, nil
	case typeHintName_Str:
		return TypeHintStr{}, nil
	case typeHintName_Objects:
		return TypeHintObjects{}, nil
	case typeHintName_Object:
		return TypeHintObject{}, nil
	case typeHintName_Int:
		return TypeHintInt{}, nil
	case typeHintName_Float:
		return TypeHintFloat{}, nil
	case typeHintName_Bool:
		return TypeHintBool{}, nil
	default:
		return nil, fmt.Errorf("unknown type hint %q", h)
	}
}

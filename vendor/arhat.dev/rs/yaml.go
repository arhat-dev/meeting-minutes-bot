package rs

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// nolint:revive
const (
	nullTag      = "!!null"
	boolTag      = "!!bool"
	strTag       = "!!str"
	intTag       = "!!int"
	floatTag     = "!!float"
	timestampTag = "!!timestamp"
	seqTag       = "!!seq"
	mapTag       = "!!map"
	binaryTag    = "!!binary"
	mergeTag     = "!!merge"

	nullTag_long      = "tag:yaml.org,2002:null"
	boolTag_long      = "tag:yaml.org,2002:bool"
	strTag_long       = "tag:yaml.org,2002:str"
	intTag_long       = "tag:yaml.org,2002:int"
	floatTag_long     = "tag:yaml.org,2002:float"
	timestampTag_long = "tag:yaml.org,2002:timestamp"
	seqTag_long       = "tag:yaml.org,2002:seq"
	mapTag_long       = "tag:yaml.org,2002:map"
	binaryTag_long    = "tag:yaml.org,2002:binary"
	mergeTag_long     = "tag:yaml.org,2002:merge"
)

const (
	longTagPrefix = "tag:yaml.org,2002:"
)

func shortTag(tag string) string {
	if strings.HasPrefix(tag, longTagPrefix) {
		switch tag {
		case nullTag_long:
			return nullTag
		case boolTag_long:
			return boolTag
		case strTag_long:
			return strTag
		case intTag_long:
			return intTag
		case floatTag_long:
			return floatTag
		case timestampTag_long:
			return timestampTag
		case seqTag_long:
			return seqTag
		case mapTag_long:
			return mapTag
		case binaryTag_long:
			return binaryTag
		case mergeTag_long:
			return mergeTag
		default:
			return "!!" + tag[len(longTagPrefix):]
		}
	}
	return tag
}

func isMerge(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode &&
		n.Value == "<<" &&
		(n.Tag == "" || n.Tag == "!" || shortTag(n.Tag) == mergeTag)
}

// is empty returns true when
// - n is nil
// - n is a document node and contains nothing
// - n is a scalar node and isNullScalar
// - n is a document node and its only child isEmpty
func isEmpty(n *yaml.Node) bool {
	for n != nil {
		switch n.Kind {
		case 0:
			// node not initialized
			return true
		case yaml.DocumentNode:
			switch len(n.Content) {
			case 0:
				return true
			case 1:
				n = n.Content[0]
			default:
				return false
			}
		case yaml.ScalarNode:
			return isNullScalar(n)
		default:
			return false
		}
	}

	return n == nil
}

func checkScalarType(n *yaml.Node, expectedTag string) bool {
	if n.Kind != yaml.ScalarNode {
		return false
	}

	if len(n.Value) == 0 && len(n.Tag) == 0 {
		n.Tag = nullTag
		return n.Tag == expectedTag
	}

	if len(n.Tag) == 0 || n.Tag == "!" {
		n.Tag = n.ShortTag()
		if len(n.Tag) != 0 {
			return n.Tag == expectedTag
		}

		// n.SchortTag may not generate valid tag value when value is not set
		if len(n.Tag) == 0 || n.Tag == "!" {
			n.Tag = nullTag
			return n.Tag == expectedTag
		}
	}

	return shortTag(n.Tag) == expectedTag
}

func isIntScalar(n *yaml.Node) bool    { return checkScalarType(n, intTag) }
func isFloatScalar(n *yaml.Node) bool  { return checkScalarType(n, floatTag) }
func isBoolScalar(n *yaml.Node) bool   { return checkScalarType(n, boolTag) }
func isNullScalar(n *yaml.Node) bool   { return checkScalarType(n, nullTag) }
func isBinaryScalar(n *yaml.Node) bool { return checkScalarType(n, binaryTag) }
func isStrScalar(n *yaml.Node) bool    { return checkScalarType(n, strTag) }

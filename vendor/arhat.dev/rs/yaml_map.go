package rs

import (
	"fmt"
	"unsafe"

	"gopkg.in/yaml.v3"
)

// prepareYamlNode finds first meaningful node (data node) for n
// result can be nil if it's an empty document (e.g. all comments)
func prepareYamlNode(n *yaml.Node) *yaml.Node {
	for n != nil {
		switch n.Kind {
		case yaml.DocumentNode:
			switch len(n.Content) {
			case 0:
				return nil
			case 1:
				n = n.Content[0]
				continue
			default:
				return nil
			}
		case yaml.AliasNode:
			n = n.Alias
			continue
		default:
			return n
		}
	}

	return n
}

// fakeMap constructs a single entry map node from two yaml nodes
func fakeMap(out, k, v *yaml.Node) {
	out.Kind = yaml.MappingNode
	out.Value = ""
	out.Content = []*yaml.Node{k, v}
}

func unmarshalYamlMap(content []*yaml.Node) (ret []*[2]*yaml.Node, err error) {
	var (
		sz         = len(content)
		szSub      int
		j          int
		subContent []*yaml.Node
	)

	for i := 0; i < sz; i += 2 {
		if !isMerge(content[i]) {
			ret = append(ret, (*[2]*yaml.Node)(unsafe.Pointer(&content[i])))
			continue
		}

		subContent, err = merge(content[i+1])
		if err != nil {
			return
		}

		szSub = len(subContent)
		for j = 0; j < szSub; j += 2 {
			ret = append(ret, (*[2]*yaml.Node)(unsafe.Pointer(&subContent[j])))
		}
	}

	return
}

func merge(n *yaml.Node) ([]*yaml.Node, error) {
	switch n.Kind {
	case yaml.AliasNode:
		if n.Alias != nil && n.Alias.Kind != yaml.MappingNode {
			return nil, fmt.Errorf(
				"invalid alias not a map for merging: got %q",
				n.Alias.ShortTag(),
			)
		}
		return n.Alias.Content, nil
	case yaml.MappingNode:
		return n.Content, nil
	case yaml.SequenceNode:
		// Step backwards as earlier nodes take precedence.
		var ret []*yaml.Node
		for i := len(n.Content) - 1; i >= 0; i-- {
			ni := n.Content[i]
			if ni.Kind == yaml.AliasNode {
				if ni.Alias != nil && ni.Alias.Kind != yaml.MappingNode {
					return nil, fmt.Errorf(
						"invalid alias in seq node not a map for merging: got %q",
						n.Alias.ShortTag(),
					)
				}
			} else if ni.Kind != yaml.MappingNode {
				return nil, fmt.Errorf(
					"invalid seq element not a map for merging: got %q",
					ni.ShortTag(),
				)
			}

			ret = append(ret, ni.Content...)
		}

		return ret, nil
	default:
		return nil, fmt.Errorf("unsupported merge source: %q", n.ShortTag())
	}
}

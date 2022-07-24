package rs

import (
	"encoding/json"
	"fmt"
	"reflect"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

type MergeSource struct {
	BaseField `yaml:"-" json:"-"`

	// Value for the source
	Value *yaml.Node `yaml:"value,omitempty"`

	// Resolve rendering suffix in value if any before being merged
	//
	// Defaults to `true`
	Resolve *bool `yaml:"resolve"`

	// Select some data from the source
	Select string `yaml:"select,omitempty"`
}

// PatchSpec is the input definition for renderers with a patching suffix
type PatchSpec struct {
	BaseField `yaml:"-" json:"-"`

	// Value for the renderer
	//
	// 	say we have a yaml list (`[bar]`) stored at https://example.com/bar.yaml
	//
	// 		foo@http!:
	// 		  value: https://example.com/bar.yaml
	// 		  merge: { value: [foo] }
	//
	// then the resolved value of foo will be `[bar, foo]`
	Value *yaml.Node `yaml:"value"`

	// Resolve rendering suffix in value if any before being patched
	//
	// Defaults to `true`
	Resolve *bool `yaml:"resolve"`

	// Merge additional data into Value
	//
	// this action happens first
	Merge []MergeSource `yaml:"merge,omitempty"`

	// Patch Value using standard rfc6902 json-patch
	//
	// this action happens after merge
	Patch []JSONPatchSpec `yaml:"patch"`

	// Select part of the data as final result
	//
	// this action happens after merge and patch
	// TODO: support jq variables
	Select string `yaml:"select"`

	// TODO: give following options proper names

	// Unique to make sure elements in the sequence is unique
	//
	// only effective when Value is yaml sequence
	Unique bool `yaml:"unique"`

	// MapListItemUnique to ensure items are unique in all merged lists respectively
	// lists with no merge data input are untouched
	MapListItemUnique bool `yaml:"map_list_item_unique"`

	// MapListAppend to append lists instead of replacing existing list
	MapListAppend bool `yaml:"map_list_append"`
}

func runJQ(query string, data any) (any, error) {
	q, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query: %w", err)
	}

	var (
		ret any
		n   int
		ok  bool
	)

	iter := q.Run(data)
	for {
		var v any
		v, ok = iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq query failed: %w", err)
		}

		switch n {
		case 0:
			ret = v
		case 1:
			ret = []any{ret, v}
		default:
			ret = append(ret.([]any), v)
		}

		n++
	}

	return ret, nil
}

func (s *PatchSpec) merge(rc RenderingHandler, valueData any) (any, error) {
	mergeSrc := make([]any, len(s.Merge))
	for i, m := range s.Merge {
		v, err := handleOptionalRenderingSuffixResolving(m.Value, m.Resolve, rc)
		if err != nil {
			return nil, err
		}

		if len(m.Select) != 0 {
			v, err = runJQ(m.Select, v)
			if err != nil {
				return nil, fmt.Errorf(
					"run select over merge#%d: %w",
					i, err,
				)
			}
		}

		mergeSrc[i] = v
	}

doMerge:
	switch dt := valueData.(type) {
	case []any:
		for _, merge := range mergeSrc {
			switch mt := merge.(type) {
			case []any:
				dt = append(dt, mt...)

				if s.Unique {
					dt = UniqueList(dt)
				}
			case nil:
				// no value to merge, skip
			default:
				// invalid type, not able to merge
				return nil, fmt.Errorf("unexpected non list value of merge, got %T", mt)
			}
		}

		return dt, nil
	case map[string]any:
		var err error
		for _, merge := range mergeSrc {
			switch mt := merge.(type) {
			case map[string]any:
				dt, err = MergeMap(dt, mt, s.MapListAppend, s.MapListItemUnique)
				if err != nil {
					return nil, fmt.Errorf("merge map value: %w", err)
				}
			case nil:
				// no value to merge, skip
			default:
				// invalid type, not able to merge
				return nil, fmt.Errorf("unexpected non map value of merge, got %T", mt)
			}
		}

		return dt, nil
	case nil:
		switch len(mergeSrc) {
		case 0:
			return nil, nil
		case 1:
			return mergeSrc[0], nil
		default:
			valueData = mergeSrc[0]
			mergeSrc = mergeSrc[1:]
			goto doMerge
		}
	default:
		// TODO: merge scalar data, how?
		if len(mergeSrc) != 0 {
			return nil, fmt.Errorf(
				"mergering scalar type value (%T) is not supported",
				valueData,
			)
		}

		// no merge source
		return valueData, nil
	}
}

// Apply Merge and Patch to Value, Unique is ensured if set to true
func (s *PatchSpec) Apply(rc RenderingHandler) (_ any, err error) {
	valueData, err := handleOptionalRenderingSuffixResolving(s.Value, s.Resolve, rc)
	if err != nil {
		return
	}

	data, err := s.merge(rc, valueData)
	if err != nil {
		return
	}

	type resolvedJSONPatchSpec struct {
		Operation string `json:"op"`
		Path      string `json:"path"`
		Value     any    `json:"value,omitempty"`
	}

	// apply select action to patches
	patchSrc := make([]resolvedJSONPatchSpec, len(s.Patch))
	for i, p := range s.Patch {
		var v any
		v, err = handleOptionalRenderingSuffixResolving(
			p.Value, p.Resolve, rc,
		)
		if err != nil {
			return
		}

		patchSrc[i] = resolvedJSONPatchSpec{
			Path:      p.Path,
			Operation: p.Operation,
			Value:     v,
		}

		if len(p.Select) != 0 {
			patchSrc[i].Value, err = runJQ(p.Select, patchSrc[i].Value)
			if err != nil {
				return nil, fmt.Errorf(
					"run select over patch#%d: %w",
					i, err,
				)
			}
		}
	}

	if len(patchSrc) == 0 {
		if len(s.Select) != 0 {
			data, err = runJQ(s.Select, data)
			if err != nil {
				return
			}
		}

		return data, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	patchData, err := json.Marshal(patchSrc)
	if err != nil {
		return
	}

	patch, err := jsonpatch.DecodePatch(patchData)
	if err != nil {
		return
	}

	options := jsonpatch.ApplyOptions{
		SupportNegativeIndices:   true,
		EnsurePathExistsOnAdd:    false,
		AccumulatedCopySizeLimit: 0,
		AllowMissingPathOnRemove: true,
	}

	patchedDoc, err := patch.ApplyIndentWithOptions(jsonData, "", &options)
	if err != nil {
		return
	}

	var ret any
	err = json.Unmarshal(patchedDoc, &ret)
	if err != nil {
		err = fmt.Errorf("unmarshal patched value: %w", err)
		return
	}

	if len(s.Select) == 0 {
		return ret, nil
	}

	ret, err = runJQ(s.Select, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func MergeMap(
	original, additional map[string]any,

	// options
	appendList bool,
	uniqueInListItems bool,
) (map[string]any, error) {
	out := make(map[string]any, len(original))
	for k, v := range original {
		out[k] = v
	}

	var err error
	for k, v := range additional {
		switch newVal := v.(type) {
		case map[string]any:
			if originalVal, ok := out[k]; ok {
				if orignalMap, ok := originalVal.(map[string]any); ok {
					out[k], err = MergeMap(orignalMap, newVal, appendList, uniqueInListItems)
					if err != nil {
						return nil, err
					}

					continue
				} else {
					return nil, fmt.Errorf("unexpected non map data %q: %v", k, orignalMap)
				}
			} else {
				out[k] = newVal
			}
		case []any:
			if originalVal, ok := out[k]; ok {
				if originalList, ok := originalVal.([]any); ok {
					if appendList {
						originalList = append(originalList, newVal...)
					} else {
						originalList = newVal
					}

					if uniqueInListItems {
						originalList = UniqueList(originalList)
					}

					out[k] = originalList

					continue
				} else {
					return nil, fmt.Errorf("unexpected non list data %q: %v", k, originalList)
				}
			} else {
				out[k] = newVal
			}
		default:
			out[k] = newVal
		}
	}

	return out, nil
}

func UniqueList(dt []any) []any {
	var ret []any
	dupAt := make(map[int]struct{})
	for i := range dt {
		_, isDup := dupAt[i]
		if isDup {
			continue
		}

		for j := i; j < len(dt); j++ {
			if reflect.DeepEqual(dt[i], dt[j]) {
				dupAt[j] = struct{}{}
			}
		}

		ret = append(ret, dt[i])
	}

	return ret
}

// JSONPatchSpec per rfc6902
type JSONPatchSpec struct {
	BaseField `yaml:"-" json:"-"`

	Operation string `yaml:"op"`

	Path string `yaml:"path"`

	Value *yaml.Node `yaml:"value,omitempty"`

	// Resolve rendering suffix in value before being applied
	//
	// Defaults to `true`
	Resolve *bool `yaml:"resolve"`

	// Select part of the value for patching
	//
	// this action happens before patching
	Select string `yaml:"select"`
}

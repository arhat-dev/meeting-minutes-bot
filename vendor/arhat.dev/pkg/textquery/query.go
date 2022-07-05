package textquery

import (
	"fmt"
	"io"
	"strconv"

	"github.com/itchyny/gojq"
)

// QueryResultHandleFunc handles each query result
// NOTE: result can be a empty/nil slice when there was no matched content or query returned nothing
type QueryResultHandleFunc func(data any, result []any, queryErr error) error

// DocIterFunc unmarshals next available doc into golang any object (map, slice)
// return true when there is doc available, otherwise false
type DocIterFunc func() (any, bool)

// Query is a generic jq query wrapper
func Query(
	query string,
	extraValues []any,
	nextDoc DocIterFunc,
	handleResult QueryResultHandleFunc,
	options ...gojq.CompilerOption,
) (err error) {
	q, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("textquery: parse %q: %w", query, err)
	}

	for {
		data, ok := nextDoc()
		if !ok {
			break
		}

		result, err2 := RunQuery(q, data, extraValues, options...)
		err = handleResult(data, result, err2)
		if err != nil {
			return
		}
	}

	return
}

// RunQuery runs jq query over an object (map, slice)
//
// if options contains gojq.WithVariables, length of extraValues should match variable count
func RunQuery(
	query *gojq.Query,
	object any,
	extraValues []any,
	options ...gojq.CompilerOption,
) (ret []any, err error) {
	code, err := gojq.Compile(query, options...)
	if err != nil {
		return nil, fmt.Errorf("compile query with variables: %w", err)
	}

	iter := code.Run(object, extraValues...)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok = v.(error); ok {
			return nil, err
		}

		ret = append(ret, v)
	}

	return
}

func CreateResultToTextHandleFuncForJsonOrYaml(
	output io.StringWriter,
	marshalFunc func(in any) ([]byte, error),
) QueryResultHandleFunc {
	notWrote := true
	return func(data any, result []any, queryErr error) error {
		if len(result) == 0 {
			return queryErr
		}

		if notWrote {
			notWrote = false
		} else {
			output.WriteString("\n")
		}

		output.WriteString(MarshalJsonOrYamlQueryResult(result, marshalFunc))
		return nil
	}
}

// MarshalJsonOrYamlQueryResult from RunQuery
func MarshalJsonOrYamlQueryResult(
	result []any,
	marshalFunc func(in any) ([]byte, error),
) string {
	switch len(result) {
	case 0:
		return ""
	case 1:
		switch r := result[0].(type) {
		case string:
			return r
		case []byte:
			return string(r)
		case []any, map[string]any:
			res, _ := marshalFunc(r)
			return string(res)
		case int64:
			return strconv.FormatInt(r, 10)
		case int32:
			return strconv.FormatInt(int64(r), 10)
		case int16:
			return strconv.FormatInt(int64(r), 10)
		case int8:
			return strconv.FormatInt(int64(r), 10)
		case int:
			return strconv.FormatInt(int64(r), 10)
		case uint64:
			return strconv.FormatUint(r, 10)
		case uint32:
			return strconv.FormatUint(uint64(r), 10)
		case uint16:
			return strconv.FormatUint(uint64(r), 10)
		case uint8:
			return strconv.FormatUint(uint64(r), 10)
		case uint:
			return strconv.FormatUint(uint64(r), 10)
		case float64:
			return strconv.FormatFloat(r, 'f', -1, 64)
		case float32:
			return strconv.FormatFloat(float64(r), 'f', -1, 64)
		case bool:
			return strconv.FormatBool(r)
		case nil:
			return "null"
		default:
			return fmt.Sprint(r)
		}
	default:
		res, _ := marshalFunc(result)
		return string(res)
	}
}

package mime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMIME(t *testing.T) {
	for _, test := range []struct {
		input           string
		expectedType    string
		expectedSubtype string
	}{
		{"", "", ""},
		{"foo", "foo", ""},
		{"/bar", "", "bar"},
		{"foo/bar", "foo", "bar"},
	} {
		t.Run(test.input, func(t *testing.T) {
			m := New(test.input)
			assert.Equal(t, test.expectedType, m.Type())
			assert.Equal(t, test.expectedSubtype, m.Subtype())
			assert.Equal(t, test.input, m.Value)
		})
	}
}

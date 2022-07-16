package telegraph

import (
	"bytes"
	"testing"

	"arhat.dev/mbot/pkg/rt"
	"github.com/stretchr/testify/assert"
)

func TestHTMLToNodes(t *testing.T) {
	var rd bytes.Reader
	rd.Reset([]byte(`<p>test</p>bar<h3>foo</h3>`))
	nodes, err := htmlToNodes(&rd)
	assert.NoError(t, err)
	assert.EqualValues(t, []telegraphNode{
		{
			Elm: rt.NewOptionalValue(telegraphNodeElement{
				Tag: "p",
				Children: []telegraphNode{
					{
						Text: "test",
					},
				},
			}),
		},
		{
			Text: "bar",
		},
		{
			Elm: rt.NewOptionalValue(telegraphNodeElement{
				Tag: "h3",
				Children: []telegraphNode{
					{
						Text: "foo",
					},
				},
			}),
		},
	}, nodes)
}

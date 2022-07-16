package telegraph

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func htmlToNodes(r io.Reader) (ret []telegraphNode, err error) {
	var ctx html.Node
	ctx.Type = html.ElementNode
	nodes, err := html.ParseFragment(r, &ctx)
	if err != nil {
		return
	}

	for _, n := range nodes {
		ret = append(ret, domToNode(n))
	}

	return
}

func domToNode(dom *html.Node) (ret telegraphNode) {
	if dom.Type == html.TextNode {
		ret.Text = dom.Data
		return
	}

	if dom.Type != html.ElementNode {
		return
	}

	var elem telegraphNodeElement
	elem.Tag = dom.Data
	if len(dom.Attr) != 0 {
		elem.Attrs = make(map[string]string)
	}

	for _, attr := range dom.Attr {
		elem.Attrs[attr.Key] = attr.Val
	}

	for child := dom.FirstChild; child != nil; child = child.NextSibling {
		elem.Children = append(elem.Children, domToNode(child))
	}

	ret.Elm.Set(elem)
	return
}

func formatJSONStrings[T ~string](s []T) string {
	var (
		sz int
		sb strings.Builder
	)

	for _, str := range s {
		sz += len(str) + 3 // `"str",`
	}

	last := len(s) - 1

	sb.Grow(sz + 2 - 1 /* `[...]` with last comma removed */)
	sb.WriteByte('[')
	for i := range s {
		sb.WriteByte('"')
		sb.WriteString(string(s[i]))
		sb.WriteByte('"')
		if i == last {
			break
		}

		sb.WriteByte(',')
	}
	sb.WriteByte(']')

	return sb.String()
}

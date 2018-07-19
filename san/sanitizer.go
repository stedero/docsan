package san

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Sanitize comments out unwanted elements in HTML
func Sanitize(writer io.Writer, data string) error {
	doc, err := html.Parse(strings.NewReader(data))
	if err != nil {
		return err
	}
	toReplace := make([]*html.Node, 0, 10)
	check := func(n *html.Node) {
		if ignore(n) {
			toReplace = append(toReplace, n)
		}
	}
	scanTree(doc, check)
	replaceWithComments(toReplace)
	return html.Render(writer, doc)
}

func scanTree(n *html.Node, check func(*html.Node)) {
	check(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		scanTree(c, check)
	}
}

func replaceWithComments(nodes []*html.Node) {
	for _, node := range nodes {
		parent := node.Parent
		parent.InsertBefore(toComment(node), node)
		parent.RemoveChild(node)
	}
}

func toComment(n *html.Node) *html.Node {
	var b bytes.Buffer
	html.Render(&b, n)
	m := &html.Node{
		Type:     html.CommentNode,
		DataAtom: n.DataAtom,
		Data:     b.String(),
		Attr:     make([]html.Attribute, len(n.Attr)),
	}
	copy(m.Attr, n.Attr)
	return m
}

func ignore(node *html.Node) bool {
	return node.Type == html.ElementNode && (node.Data == "script" || isStylesheetLink(node))
}

func isStylesheetLink(node *html.Node) bool {
	if node.Data == "link" {
		for _, attr := range node.Attr {
			if attr.Key == "rel" && attr.Val == "stylesheet" {
				return true
			}
		}
	}
	return false
}

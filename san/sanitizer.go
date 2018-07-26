package san

import (
	"io"

	"golang.org/x/net/html"
	"ibfd.org/docsan/node"
)

// SanitizeHTML comments out unwanted elements in HTML
func SanitizeHTML(r io.Reader) (*html.Node, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	node.ReplaceWithComments(doc, accept())
	return doc, nil
}

func accept() node.Check {
	isScript := node.Element("script")
	isStylesheetLink := node.And(node.Element("link"), node.Attr("rel", "stylesheet"))
	return node.Or(isScript, isStylesheetLink)
}

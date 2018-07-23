package render

import (
	"golang.org/x/net/html"
	"ibfd.org/docsan/doc"
	"ibfd.org/docsan/node"
)

// ToJSON transforms a HTML node to JSON
func ToJSON(htmlDoc *html.Node) ([]byte, error) {
	head := node.FindFirst(htmlDoc, accept("head"))
	body := node.FindFirst(htmlDoc, accept("body"))
	title := node.FindFirst(head, accept("title"))
	metas := node.FindAll(head, accept("meta"))
	document := doc.NewDocument(title, metas, body)
	return document.ToJSON()
}

func accept(element string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == element
	}
}

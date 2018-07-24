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
	node.ReplaceWithComments(doc, accept)
	return doc, nil
}

func accept(node *html.Node) bool {
	return isScriptElement(node) || isStylesheetLink(node)
}

func isScriptElement(node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "script"
}

func isStylesheetLink(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "link" {
		for _, attr := range node.Attr {
			if attr.Key == "rel" && attr.Val == "stylesheet" {
				return true
			}
		}
	}
	return false
}

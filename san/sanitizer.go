package san

import (
	"io"
	"strings"

	"golang.org/x/net/html"
	"ibfd.org/docsan/node"
)

// Sanitize comments out unwanted elements in HTML
func Sanitize(writer io.Writer, data string) error {
	doc, err := html.Parse(strings.NewReader(data))
	if err != nil {
		return err
	}
	checker := node.NewChecker(accept)
	checker.ScanTree(doc)
	checker.ReplaceWithComments()
	return html.Render(writer, doc)
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

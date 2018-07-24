package node

import (
	"bytes"

	"golang.org/x/net/html"
)

// ToString renders a node to a string
func ToString(node *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, node)
	return b.String()
}

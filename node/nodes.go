package node

import (
	"bytes"

	"golang.org/x/net/html"
)

// Check defines functions for checking nodes
type Check func(*html.Node) bool

// CheckAttrs defines functions for checking node attributes
type CheckAttrs func(attrs map[string]string) bool

// checker holds nodes that are accepted by the accept function
type checker struct {
	accept Check
	nodes  []*html.Node
}

func newChecker(accept Check) *checker {
	return &checker{accept, make([]*html.Node, 0, 32)}
}

// Check determines whether a node must be accepted
func (chkr *checker) check(node *html.Node) {
	if chkr.accept(node) {
		chkr.nodes = append(chkr.nodes, node)
	}
}

// ReplaceWithComments replaces all nodes that are accepted with comment nodes.
// Please note that the same node is returned that was passed to this function
// because that node is transformed by this function.
func ReplaceWithComments(node *html.Node, accept Check) *html.Node {
	for _, n := range FindAll(node, accept) {
		parent := n.Parent
		parent.InsertBefore(toComment(n), n)
		parent.RemoveChild(n)
	}
	return node
}

func toComment(n *html.Node) *html.Node {
	return &html.Node{Type: html.CommentNode, DataAtom: n.DataAtom, Data: Render(n)}
}

// FindFirst finds the first node that is accepted
func FindFirst(n *html.Node, accept Check) *html.Node {
	nodes := FindAll(n, accept)
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// FindAll finds all nodes that are accepted
func FindAll(n *html.Node, accept Check) []*html.Node {
	chkr := newChecker(accept)
	chkr.walk(n)
	return chkr.nodes
}

// walk the node tree
func (chkr *checker) walk(n *html.Node) {
	chkr.check(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		chkr.walk(c)
	}
}

// Content gets the first child node as an unescaped HTML string.
func Content(n *html.Node) string {
	return html.UnescapeString(Render(n.FirstChild))
}

// Render a node to a string
func Render(n *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, n)
	return b.String()
}

// And returns a function that applies a logical AND operation
// to the results of two functions
func And(f1, f2 Check) Check {
	return func(n *html.Node) bool {
		return f1(n) && f2(n)
	}
}

// Or returns a function that applies a logical OR operation
// to the results of two functions
func Or(f1, f2 Check) Check {
	return func(n *html.Node) bool {
		return f1(n) || f2(n)
	}
}

// Not returns a function that negates the output of the supplied function.
func Not(f Check) Check {
	return func(n *html.Node) bool {
		return !f(n)
	}
}

// Element returns a function that checks whether a node is an element of the specified type
func Element(element string) Check {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == element
	}
}

// Attr returns a function that checks whether a node has a specific attribute
func Attr(key, value string) Check {
	return func(n *html.Node) bool {
		for _, attr := range n.Attr {
			if attr.Key == key && attr.Val == value {
				return true
			}
		}
		return false
	}
}

// ToMapArray creates an array of node attributes as a map
func ToMapArray(nodes []*html.Node) []map[string]string {
	return ToMapArrayFiltered(nodes, acceptAll)
}

// ToMapArrayFiltered creates a filtered array of node attributes as a map
func ToMapArrayFiltered(nodes []*html.Node, accept CheckAttrs) []map[string]string {
	metas := make([]map[string]string, 0, len(nodes))
	for _, n := range nodes {
		m := AttrsAsMap(n)
		if accept(m) {
			metas = append(metas, m)
		}
	}
	return metas
}

// AttrsAsMap creates a map from node attributes.
func AttrsAsMap(n *html.Node) map[string]string {
	m := make(map[string]string)
	for _, attr := range n.Attr {
		m[attr.Key] = html.UnescapeString(attr.Val)
	}
	return m
}

func acceptAll(attrs map[string]string) bool {
	return true
}

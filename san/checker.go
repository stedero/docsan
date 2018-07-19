package san

import (
	"bytes"

	"golang.org/x/net/html"
)

// Checker keeps track of nodes that need to be replaced
type Checker struct {
	accept func(*html.Node) bool
	nodes  []*html.Node
}

func newChecker(accept func(*html.Node) bool) *Checker {
	return &Checker{accept, make([]*html.Node, 0, 10)}
}

// Check determines whether a node needs to be replaced
func (checker *Checker) Check(node *html.Node) {
	if checker.accept(node) {
		checker.nodes = append(checker.nodes, node)
	}
}

// ReplaceWithComments replaces all collected nodes with comment nodes
func (checker *Checker) ReplaceWithComments() {
	for _, node := range checker.nodes {
		parent := node.Parent
		parent.InsertBefore(toComment(node), node)
		parent.RemoveChild(node)
	}
}

func toComment(n *html.Node) *html.Node {
	var b bytes.Buffer
	html.Render(&b, n)
	return &html.Node{Type: html.CommentNode, DataAtom: n.DataAtom, Data: b.String()}
}

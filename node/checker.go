package node

import (
	"bytes"

	"golang.org/x/net/html"
)

// Checker keeps track of nodes that need to be replaced
type Checker struct {
	accept func(*html.Node) bool
	nodes  []*html.Node
}

// NewChecker creates a new node checker.
func NewChecker(accept func(*html.Node) bool) *Checker {
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

// FindAll locates all nodes with the specified element name
func (checker *Checker) FindAll(n *html.Node) []*html.Node {
	checker.ScanTree(n)
	return checker.nodes
}

// ScanTree walks the node tree
func (checker *Checker) ScanTree(n *html.Node) {
	checker.Check(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		checker.ScanTree(c)
	}
}

package node

import (
	"golang.org/x/net/html"
)

// checker holds nodes that are accepted by the accept function
type checker struct {
	accept func(*html.Node) bool
	nodes  []*html.Node
}

func newChecker(accept func(*html.Node) bool) *checker {
	return &checker{accept, make([]*html.Node, 0, 32)}
}

// Check determines whether a node must be accepted
func (chkr *checker) check(node *html.Node) {
	if chkr.accept(node) {
		chkr.nodes = append(chkr.nodes, node)
	}
}

// ReplaceWithComments replaces all nodes that are accepted with comment nodes
func ReplaceWithComments(node *html.Node, accept func(*html.Node) bool) {
	chkr := newChecker(accept)
	chkr.walk(node)
	for _, n := range chkr.nodes {
		parent := n.Parent
		parent.InsertBefore(toComment(n), n)
		parent.RemoveChild(n)
	}
}

func toComment(n *html.Node) *html.Node {
	return &html.Node{Type: html.CommentNode, DataAtom: n.DataAtom, Data: ToString(n)}
}

// FindFirst finds the first node that is accepted
func FindFirst(n *html.Node, accept func(*html.Node) bool) *html.Node {
	nodes := FindAll(n, accept)
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// FindAll finds all nodes that are accepted
func FindAll(n *html.Node, accept func(*html.Node) bool) []*html.Node {
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

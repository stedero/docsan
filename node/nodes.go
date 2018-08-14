package node

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// Check defines functions for checking nodes
type Check func(*html.Node) bool

// Transform creates a node from a node
type Transform func(*html.Node) *html.Node

// CheckAttrs defines functions for checking node attributes
type CheckAttrs func(attrs map[string]string) bool

// collector collects nodes that are accepted by the accept function
type collector struct {
	accept Check
	nodes  []*html.Node
}

func newCollector(accept Check) *collector {
	return &collector{accept, make([]*html.Node, 0, 32)}
}

// Check determines whether a node must be accepted
func (coll *collector) check(node *html.Node) {
	if coll.accept(node) {
		coll.nodes = append(coll.nodes, node)
	}
}

// ReplaceWithComments replaces all nodes that are accepted with comment nodes.
// Please note that the same node is returned that was passed to this function
// because that node is transformed by this function.
func ReplaceWithComments(node *html.Node, accept Check) *html.Node {
	return replace(node, accept, toComment)
}

// ReplaceWithContent replaces all nodes that are accepted with the content of that nodes.
// Please note that the same node is returned that was passed to this function
// because that node is transformed by this function.
func ReplaceWithContent(node *html.Node, accept Check) *html.Node {
	return replace(node, accept, toContent)
}

// Replace replaces every node that is accepted by a transformation of that node.
func replace(node *html.Node, accept Check, transform Transform) *html.Node {
	for _, n := range FindAll(node, accept) {
		parent := n.Parent
		parent.InsertBefore(transform(n), n)
		parent.RemoveChild(n)
	}
	return node
}

// AttrToComment removes an attribute from a node and places
// that attribute in a comment node after its parent.
func AttrToComment(node *html.Node, key string) *html.Node {
	for _, n := range FindAll(node, And(AnyElement(), HasAttr(key))) {
		var found = -1
		for ai, attr := range n.Attr {
			if attr.Key == key {
				found = ai
			}
		}
		if found > -1 {
			// TODO: attribute from Attr slice and create as comment node
		}
	}
	return node
}

func toComment(n *html.Node) *html.Node {
	return &html.Node{Type: html.CommentNode, DataAtom: n.DataAtom, Data: Render(n)}
}

func toContent(n *html.Node) *html.Node {
	return clone(n.FirstChild)
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
	coll := newCollector(accept)
	coll.walk(n)
	return coll.nodes
}

// walk the node tree
func (coll *collector) walk(n *html.Node) {
	coll.check(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		coll.walk(c)
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

// AnyElement returns a function that checks whether a node is an element node
func AnyElement() Check {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode
	}
}

// AttrEquals returns a function that checks whether a node attribute has specific value
func AttrEquals(key, value string) Check {
	return attrCheck(key, attrEqual(value))
}

// HasAttr returns a function that checks whether a node has a specifi attribute.
func HasAttr(key string) Check {
	return attrCheck(key, attrAnyValue())
}

// AttrPrefix returns a function that checks whether a node attribute has a value
// with the specified prefix
func AttrPrefix(key, prefix string) Check {
	return attrCheck(key, attrPrefix(prefix))
}

// AttrNotPrefix returns a function that checks whether a node attribute
// does NOT have a value with the specified prefix
func AttrNotPrefix(key, prefix string) Check {
	return attrCheck(key, attrNot(attrPrefix(prefix)))
}

func attrCheck(key string, accept func(string) bool) Check {
	return func(n *html.Node) bool {
		for _, attr := range n.Attr {
			if attr.Key == key && accept(attr.Val) {
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
	list := make([]map[string]string, 0, len(nodes))
	for _, n := range nodes {
		m := AttrsAsMap(n)
		if accept(m) {
			list = append(list, m)
		}
	}
	return list
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

func attrNot(f func(string) bool) func(string) bool {
	return func(value string) bool {
		return !f(value)
	}
}

func attrEqual(value string) func(string) bool {
	return func(str string) bool {
		return str == value
	}
}

func attrAnyValue() func(string) bool {
	return func(str string) bool {
		return true
	}
}

func attrPrefix(prefix string) func(string) bool {
	return func(str string) bool {
		return strings.HasPrefix(str, prefix)
	}
}

// clone returns a new node with the same type, data and attributes.
// The clone has no parent, no siblings and no children.
func clone(n *html.Node) *html.Node {
	m := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     make([]html.Attribute, len(n.Attr)),
	}
	copy(m.Attr, n.Attr)
	return m
}

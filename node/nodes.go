package node

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
	a "golang.org/x/net/html/atom"
	log "ibfd.org/docsan/log4u"
)

// Action defines a action tat modifies the HTML structure of a document.
type Action struct {
	DocID string
}

// NewAction creates a document action.
func NewAction(docID string) *Action {
	return &Action{docID}
}

// Check defines functions to filter nodes.
type Check func(*html.Node) bool

// Transform creates a node from a node. If the transformation
// cannot be performed then nil must be returned.
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

// collect determines whether a node must be accepted
// and adds it to this collector if so.
func (coll *collector) collect(node *html.Node) {
	if coll.accept(node) {
		coll.nodes = append(coll.nodes, node)
	}
}

// ReplaceWithComments replaces all nodes that are accepted with comment nodes.
// Returns the modified node.
func ReplaceWithComments(node *html.Node, accept Check) *html.Node {
	return replace(node, accept, toComment)
}

// ReplaceWithContent replaces all nodes that are accepted with the content of that nodes.
// Returns the modified node.
func ReplaceWithContent(node *html.Node, accept Check) *html.Node {
	return replace(node, accept, toContent)
}

// Replace replaces every node that is accepted by a transformation of that node.
func replace(node *html.Node, accept Check, transform Transform) *html.Node {
	for _, n := range FindAll(node, accept) {
		parent := n.Parent
		result := transform(n)
		if result != nil {
			parent.InsertBefore(result, n)
			parent.RemoveChild(n)
		}
	}
	return node
}

// DisableAttribute prefixes an attribute key with 'xxx' to disable it.
// To be used for JavaScript events such as 'onclick'.
func (action *Action) DisableAttribute(node *html.Node, key string, accept Check) *html.Node {
	nodes := FindAll(node, accept)
	if len(nodes) > 0 {
		action.Log(fmt.Sprintf("disabling %d %s events", len(nodes), key))
	}
	for _, n := range nodes {
		var found = -1
		for ai, attr := range n.Attr {
			if attr.Key == key {
				found = ai
			}
		}
		if found > -1 {
			n.Attr[found].Key = "xxx" + n.Attr[found].Key
		}
	}
	return node
}

func toComment(n *html.Node) *html.Node {
	return &html.Node{Type: html.CommentNode, DataAtom: n.DataAtom, Data: Render(n)}
}

// toContent fetches the node contents assuming one child node.
func toContent(n *html.Node) *html.Node {
	if n.FirstChild != nil {
		return clone(n.FirstChild)
	}
	return nil
}

// AddNoticePlaceholders add placeholders for the action bars
func (action *Action) AddNoticePlaceholders(node *html.Node, accept Check) *html.Node {
	nodes := FindAll(node, accept)
	if len(nodes) > 0 {
		action.Log(fmt.Sprintf("adding %d notice placeholders", len(nodes)))
	}
	for _, n := range nodes {
		attrMap := AttrsAsMap(n)
		id, _ := attrMap["id"]
		attr1 := html.Attribute{Key: "id", Val: "notice_" + id}
		attr2 := html.Attribute{Key: "class", Val: "ib-notice"}
		attr3 := html.Attribute{Key: "data-generator", Val: "docsan"}
		attrs := []html.Attribute{attr1, attr2, attr3}
		div := newDiv(attrs)
		n.InsertBefore(div, n.FirstChild)
	}
	return node
}

// AddSeeAlsoPlaceholders add placeholders for the see also sections.
func (action *Action) AddSeeAlsoPlaceholders(node *html.Node, accept Check) *html.Node {
	nodes := FindAll(node, accept)
	if len(nodes) > 0 {
		action.Log(fmt.Sprintf("adding %d seealso placeholders", len(nodes)))
	}
	for _, n := range nodes {
		attrMap := AttrsAsMap(n)
		id, _ := attrMap["id"]
		attr1 := html.Attribute{Key: "id", Val: "seealso_" + id}
		attr2 := html.Attribute{Key: "class", Val: "ib-seealso"}
		attr3 := html.Attribute{Key: "data-generator", Val: "docsan"}
		attrs := []html.Attribute{attr1, attr2, attr3}
		div := newDiv(attrs)
		n.AppendChild(div)
	}
	return node
}

// WrapTables wraps nodes in a div
func (action *Action) WrapTables(node *html.Node, accept Check) *html.Node {
	nodes := FindAll(node, accept)
	if len(nodes) > 0 {
		action.Log(fmt.Sprintf("wrapping %d tables", len(nodes)))
	}
	for _, n := range nodes {
		attr1 := html.Attribute{Key: "class", Val: "ib-table-wrapper"}
		attr2 := html.Attribute{Key: "data-generator", Val: "docsan"}
		attrs := []html.Attribute{attr1, attr2}
		div := newDiv(attrs)
		parent := n.Parent
		parent.InsertBefore(div, n)
		move(div, n)
	}
	return node
}

// Log logs an action.
func (action *Action) Log(msg string) {
	log.Errorf("DS1955 %s: %s", action.DocID, msg)
}

func newDiv(attrs []html.Attribute) *html.Node {
	return &html.Node{
		Type:     html.ElementNode,
		DataAtom: a.Div,
		Data:     a.Div.String(),
		Attr:     attrs,
	}
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

// Remove removes all selected nodes.
func Remove(node *html.Node, accept Check) *html.Node {
	for _, n := range FindAll(node, accept) {
		parent := n.Parent
		parent.RemoveChild(n)
	}
	return node
}

// walk the node tree
func (coll *collector) walk(n *html.Node) {
	coll.collect(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		coll.walk(c)
	}
}

// Content gets the childdren of a node as an unescaped HTML string.
func Content(n *html.Node) string {
	if n == nil {
		return ""
	}
	return html.UnescapeString(RenderChildren(n))
}

// RenderChildrenCommentParent renders all children of a node
// and surrounds that with the parent start and end element as comment.
func RenderChildrenCommentParent(n *html.Node) string {
	var b bytes.Buffer
	startElm, endElm := asCommentElements(n)
	b.WriteString(startElm)
	renderChildren(&b, n)
	b.WriteString(endElm)
	return b.String()
}

// asCommentElements transforms the start and end elements of
// a node to comment elements.
func asCommentElements(n *html.Node) (string, string) {
	parts := strings.SplitAfterN(Render(clone(n)), ">", 2)
	return asCommentElement(parts[0]), asCommentElement(parts[1])
}

func asCommentElement(str string) string {
	return "<!--" + str + "-->"
}

// RenderChildren renders all children of a node to a string.
func RenderChildren(n *html.Node) string {
	var b bytes.Buffer
	renderChildren(&b, n)
	return b.String()
}

// Render a node to a string
func Render(n *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, n)
	return b.String()
}

// renderChildren renders all children of a node.
func renderChildren(w io.Writer, n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		html.Render(w, c)
	}
}

// And returns a function that applies a logical AND operation to
// the results of all supplied functions.
func And(checks ...Check) Check {
	return func(n *html.Node) bool {
		for _, chk := range checks {
			if !chk(n) {
				return false
			}
		}
		return true
	}
}

// Or returns a function that applies a logical OR operation to
// the results of all supplied functions.
func Or(checks ...Check) Check {
	return func(n *html.Node) bool {
		for _, chk := range checks {
			if chk(n) {
				return true
			}
		}
		return false
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

// HasAttr returns a function that checks whether a node has a specific attribute.
func HasAttr(key string) Check {
	return attrCheck(key, attrAnyValue())
}

// AttrPrefix returns a function that checks whether a node attribute has a value
// with the specified prefix
func AttrPrefix(key, prefix string) Check {
	return attrCheck(key, attrPrefix(prefix))
}

// AttrContains returns a function that checks whether a node attribute
// value contains a substring
func AttrContains(key, substr string) Check {
	return attrCheck(key, attrContains(substr))
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

func attrContains(value string) func(string) bool {
	return func(str string) bool {
		return strings.Contains(str, value)
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

// move moves a node from one parent to another.
func move(dst *html.Node, c *html.Node) {
	parent := c.Parent
	n := clone(c)
	reparentChildren(n, c)
	parent.RemoveChild(c)
	dst.AppendChild(n)
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

// reparentChildren reparents all of src's child nodes to dst.
func reparentChildren(dst, src *html.Node) {
	for {
		child := src.FirstChild
		if child == nil {
			break
		}
		src.RemoveChild(child)
		dst.AppendChild(child)
	}
}

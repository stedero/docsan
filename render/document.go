package render

import (
	"encoding/json"

	"golang.org/x/net/html"
	"ibfd.org/docsan/node"
)

// Document defines a document to render as JSON
type document struct {
	Title string              `json:"title"`
	Metas []map[string]string `json:"metas"`
	Body  string              `json:"body"`
}

// ToJSON transforms a HTML node to JSON
func ToJSON(htmlDoc *html.Node) ([]byte, error) {
	head := node.FindFirst(htmlDoc, accept("head"))
	body := node.FindFirst(htmlDoc, accept("body"))
	title := node.FindFirst(head, accept("title"))
	metas := node.FindAll(head, accept("meta"))
	d := newDocument(title, metas, body)
	return d.toJSON()
}

func (d *document) toJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

func accept(element string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == element
	}
}

// newDocument create a new document
func newDocument(titleNode *html.Node, metaNodes []*html.Node, bodyNode *html.Node) *document {
	return &document{node.ToString(titleNode), toMetas(metaNodes), node.ToString(bodyNode)}
}

func toMetas(nodes []*html.Node) []map[string]string {
	metas := make([]map[string]string, 0, len(nodes))
	for _, node := range nodes {
		m := make(map[string]string)
		for _, attr := range node.Attr {
			m[attr.Key] = attr.Val
		}
		metas = append(metas, m)
	}
	return metas
}

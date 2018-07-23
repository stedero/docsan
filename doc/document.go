package doc

import (
	"bytes"
	"encoding/json"

	"golang.org/x/net/html"
)

// Document defines a document to render as JSON
type Document struct {
	Title string              `json:"title"`
	Metas []map[string]string `json:"metas"`
	Body  string              `json:"body"`
}

// NewDocument create a new document
func NewDocument(titleNode *html.Node, metaNodes []*html.Node, bodyNode *html.Node) *Document {
	return &Document{toString(titleNode), toMetas(metaNodes), toString(bodyNode)}
}

// ToJSON transforms a document to JSON
func (document *Document) ToJSON() ([]byte, error) {
	return json.MarshalIndent(document, "", "    ")
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

func toString(node *html.Node) string {
	var b bytes.Buffer
	html.Render(&b, node)
	return b.String()
}

package render

import (
	"encoding/json"
	"log"

	"golang.org/x/net/html"
	"ibfd.org/docsan/node"
)

// JSON defines pre-rendered JSON
type JSON struct {
	json string
}

// document defines a document to render as JSON
type document struct {
	Generated string              `json:"generated"`
	Title     string              `json:"title"`
	Outline   *JSON               `json:"outline"`
	Metas     []map[string]string `json:"metas"`
	Body      string              `json:"body"`
}

// ToJSON transforms a HTML node to JSON
func ToJSON(htmlDoc *html.Node, version string) ([]byte, error) {
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	outline := node.FindFirst(head, node.And(node.Element("script"), node.Attr("id", "outline")))
	metas := node.FindAll(head, node.Element("meta"))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	sanitizedBody := node.ReplaceWithComments(body, commentTargetSelector())
	d := newDocument(version, title, outline, metas, sanitizedBody)
	return d.toJSON()
}

// newDocument create a new document
func newDocument(version string, titleNode *html.Node, outline *html.Node, metaNodes []*html.Node, bodyNode *html.Node) *document {
	return &document{"docsan " + version, node.Content(titleNode), formatOutline(outline), toMetas(metaNodes), node.Render(bodyNode)}
}

func (d *document) toJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

// MarshalJSON marshals a pre-rendered JSON object
func (j JSON) MarshalJSON() ([]byte, error) {
	log.Printf("JSON: %s", string(j.json))
	return []byte(j.json), nil
}

func toMetas(nodes []*html.Node) []map[string]string {
	metas := make([]map[string]string, 0, len(nodes))
	for _, node := range nodes {
		m := make(map[string]string)
		for _, attr := range node.Attr {
			m[attr.Key] = html.UnescapeString(attr.Val)
		}
		metas = append(metas, m)
	}
	return metas
}

func commentTargetSelector() node.Check {
	isScript := node.Element("script")
	isStylesheetLink := node.And(node.Element("link"), node.Attr("rel", "stylesheet"))
	return node.Or(isScript, isStylesheetLink)
}

func formatOutline(n *html.Node) *JSON {
	var data string
	if n == nil {
		data = "{}"
	} else {
		data = n.FirstChild.Data
	}
	return &JSON{data}
}

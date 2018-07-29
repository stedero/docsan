package render

import (
	"encoding/json"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
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
	Links     []map[string]string `json:"links"`
	Scripts   []map[string]string `json:"scripts"`
	Body      string              `json:"body"`
}

// ToJSON transforms a HTML node to JSON
func ToJSON(htmlDoc *html.Node, generated string) ([]byte, error) {
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	outline := node.FindFirst(head, node.And(node.Element("script"), node.Attr("id", "outline")))
	metas := node.FindAll(head, node.Element("meta"))
	links := node.FindAll(head, node.Element("link"))
	scripts := node.FindAll(head, node.Element("script"))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	sanitizedBody := node.ReplaceWithComments(body, commentTargetSelector())
	d := newDocument(generated, title, outline, metas, links, scripts, sanitizedBody)
	return d.toJSON()
}

// newDocument create a new document
func newDocument(generated string, titleNode *html.Node, outline *html.Node, metaNodes []*html.Node, linkNodes []*html.Node, scriptNodes []*html.Node, bodyNode *html.Node) *document {
	return &document{
		Generated: "docsan " + generated,
		Title:     node.Content(titleNode),
		Outline:   formatOutline(outline),
		Metas:     toMetas(metaNodes),
		Links:     node.ToMapArray(linkNodes),
		Scripts:   node.ToMapArray(scriptNodes),
		Body:      node.Render(bodyNode)}
}

func (d *document) toJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "    ")
}

// MarshalJSON marshals a pre-rendered JSON object
func (j JSON) MarshalJSON() ([]byte, error) {
	return []byte(j.json), nil
}

func toMetas(nodes []*html.Node) []map[string]string {
	metaNameAccept := metaAccept(config.MetaNameAccept())
	metas := node.ToMapArrayFiltered(nodes, metaNameAccept)
	return metas
}

func commentTargetSelector() node.Check {
	isScript := node.Element("script")
	isStylesheetLink := node.And(node.Element("link"), node.Attr("rel", "stylesheet"))
	return node.Or(isScript, isStylesheetLink)
}

func metaAccept(acceptMetaName func(string) bool) node.CheckAttrs {
	return func(attrs map[string]string) bool {
		name, present := attrs["name"]
		return !present || acceptMetaName(name)
	}
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

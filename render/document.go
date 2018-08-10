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
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	anchorSelector := node.Element("a")
	anchorTypeSelector := node.AttrNotPrefix("href", "#")
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	outline := node.FindFirst(head, node.And(scriptSelector, outLineAttrChecker))
	metas := node.FindAll(head, node.Element("meta"))
	links := node.FindAll(head, node.Element("link"))
	scripts := node.FindAll(head, node.And(scriptSelector, node.Not(outLineAttrChecker)))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	san1Body := node.ReplaceWithComments(body, commentTargetSelector())
	san2Body := node.ReplaceWithContent(san1Body, node.And(anchorSelector, anchorTypeSelector))
	d := newDocument(generated, title, outline, metas, links, scripts, san2Body)
	return d.toJSON()
}

// newDocument create a new document
func newDocument(generated string, title *html.Node, outline *html.Node, metas []*html.Node, links []*html.Node, scripts []*html.Node, sanitizedBody *html.Node) *document {
	return &document{
		Generated: "docsan " + generated,
		Title:     node.Content(title),
		Outline:   formatOutline(outline),
		Metas:     toMetas(metas),
		Links:     node.ToMapArray(links),
		Scripts:   node.ToMapArray(scripts),
		Body:      node.Render(sanitizedBody)}
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
	isStylesheetLink := node.And(node.Element("link"), node.AttrEquals("rel", "stylesheet"))
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

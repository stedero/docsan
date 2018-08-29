package render

import (
	"encoding/json"
	"io"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	"ibfd.org/docsan/node"
)

// JSON defines pre-rendered JSON
type JSON struct {
	json string
}

// Document defines a document to render as JSON
type Document struct {
	Generated string              `json:"generated"`
	Title     string              `json:"title"`
	Outline   *JSON               `json:"outline"`
	Metas     []map[string]string `json:"metas"`
	Links     []map[string]string `json:"links"`
	Scripts   []map[string]string `json:"scripts"`
	Body      string              `json:"body"`
}

// Transform transforms a HTML node to a document structure for JSON output.
func Transform(htmlDoc *html.Node, generated string) *Document {
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	jsonOutline := formatOutline(node.FindFirst(htmlDoc, node.And(scriptSelector, outLineAttrChecker)))
	metas := node.FindAll(head, node.Element("meta"))
	links := node.FindAll(head, node.Element("link"))
	scripts := node.FindAll(head, node.And(scriptSelector, node.Not(outLineAttrChecker)))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	san1Body := node.ReplaceWithComments(body, commentTargetSelector())
	san2Body := node.ReplaceWithContent(san1Body, notInternalLinkSelector())
	san3Body := node.DisableAttribute(san2Body, "onclick")
	return newDocument(generated, title, jsonOutline, metas, links, scripts, san3Body)
}

// ToJSON renders a document to JSON.
func (document *Document) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

// newDocument create a new document
func newDocument(generated string, title *html.Node, jsonOutline *JSON, metas []*html.Node, links []*html.Node, scripts []*html.Node, sanitizedBody *html.Node) *Document {
	return &Document{
		Generated: "docsan " + generated,
		Title:     node.Content(title),
		Outline:   jsonOutline,
		Metas:     toMetas(metas),
		Links:     node.ToMapArray(links),
		Scripts:   node.ToMapArray(scripts),
		Body:      node.RenderChildrenCommentParent(sanitizedBody)}
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

func notInternalLinkSelector() node.Check {
	anchorSelector := node.Element("a")
	anchorTypeSelector := node.AttrNotPrefix("href", "#")
	return node.And(anchorSelector, anchorTypeSelector)
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

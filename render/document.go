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
	Sumtab    *JSON               `json:"sumtab"`
	Metas     []map[string]string `json:"metas"`
	Links     []map[string]string `json:"links"`
	Scripts   []map[string]string `json:"scripts"`
	Body      string              `json:"body"`
}

// Transform transforms a HTML node to a document structure for JSON output.
func Transform(htmlDoc *html.Node, generated string) *Document {
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	sumtabAttrChecker := node.AttrEquals("id", "sumtab")
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	jsonOutline := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, outLineAttrChecker)))
	jsonSumtab := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, sumtabAttrChecker)))
	metas := node.FindAll(head, node.Element("meta"))
	links := node.FindAll(head, node.Element("link"))
	scripts := node.FindAll(head, node.And(scriptSelector, node.Not(node.Or(outLineAttrChecker, sumtabAttrChecker))))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	san1Body := addNoticePlaceholdersIfNeeded(body)
	san2Body := node.ReplaceWithComments(san1Body, commentTargetSelector())
	san3Body := node.DisableAttribute(san2Body, "onclick", disableAtributeSelector())
	return newDocument(generated, title, jsonOutline, jsonSumtab, metas, links, scripts, san3Body)
}

// ToJSON renders a document to JSON.
func (document *Document) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

// newDocument create a new document
func newDocument(generated string, title *html.Node, jsonOutline *JSON, jsonSumtab *JSON, metas []*html.Node, links []*html.Node, scripts []*html.Node, sanitizedBody *html.Node) *Document {
	return &Document{
		Generated: "docsan " + generated,
		Title:     node.Content(title),
		Outline:   jsonOutline,
		Sumtab:    jsonSumtab,
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
	isCompareParagraph := node.And(node.Element("p"), node.AttrEquals("class", "compare-to"))
	return node.Or(isScript, isStylesheetLink, isCompareParagraph)
}

func noticePlaceholderTargetSelector() node.Check {
	hasAnnotatableClass := node.AttrEquals("class", "annotatable")
	hasID := node.HasAttr("id")
	return node.And(node.AnyElement(), hasAnnotatableClass, hasID)
}

func notInternalLinkSelector() node.Check {
	anchorSelector := node.Element("a")
	anchorTypeSelector := node.AttrNotPrefix("href", "#")
	return node.And(anchorSelector, anchorTypeSelector)
}

func disableAtributeSelector() node.Check {
	clickEvents := node.HasAttr("onclick")
	simultaxButton := node.AttrEquals("class", "dyncal-button")
	return node.And(node.AnyElement(), clickEvents, node.Not(simultaxButton))
}

func addNoticePlaceholdersIfNeeded(body *html.Node) *html.Node {
	noticePlaceholders := node.FindFirst(body, noticePlaceholder())
	if noticePlaceholders == nil {
		return node.AddNoticePlaceholders(body, noticePlaceholderTargetSelector())
	}
	return body
}

func noticePlaceholder() node.Check {
	isDiv := node.Element("div")
	isNoticePlaceholder := node.AttrPrefix("id", "notice_")
	return node.And(isDiv, isNoticePlaceholder)
}

func metaAccept(acceptMetaName func(string) bool) node.CheckAttrs {
	return func(attrs map[string]string) bool {
		name, present := attrs["name"]
		return !present || acceptMetaName(name)
	}
}

func formatJSON(n *html.Node) *JSON {
	var data string
	if n == nil {
		data = "{}"
	} else {
		data = n.FirstChild.Data
	}
	return &JSON{data}
}

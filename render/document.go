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
	DocLinks  *JSON               `json:"links"`
	Metas     []map[string]string `json:"metas"`
	Scripts   []map[string]string `json:"scripts"`
	Body      string              `json:"body"`
}

// Transform transforms a HTML node to a document structure for JSON output.
func Transform(htmlDoc *html.Node, generated string) *Document {
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	sumtabAttrChecker := node.AttrEquals("id", "sumtab")
	linksAttrChecker := node.AttrEquals("id", "links")
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	jsonOutline := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, outLineAttrChecker)))
	jsonSumtab := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, sumtabAttrChecker)))
	jsonLinks := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, linksAttrChecker)))
	metas := node.FindAll(head, node.Element("meta"))
	scripts := node.FindAll(head, node.And(scriptSelector, node.Not(node.Or(outLineAttrChecker, sumtabAttrChecker, linksAttrChecker))))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	san1Body := addNoticePlaceholdersIfNeeded(body)
	san2Body := addSeeAlsoPlaceholdersIfNeeded(san1Body)
	san3Body := node.ReplaceWithComments(san2Body, commentTargetSelector())
	san4Body := node.WrapTables(san3Body, chapterTableSelector())
	san5Body := node.DisableAttribute(san4Body, "onclick", disableAtributeSelector())
	return newDocument(generated, title, jsonOutline, jsonSumtab, jsonLinks, metas, scripts, san5Body)
}

// ToJSON renders a document to JSON.
func (document *Document) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

// newDocument create a new document
func newDocument(generated string, title *html.Node, jsonOutline *JSON, jsonSumtab *JSON, jsonLinks *JSON, metas []*html.Node, scripts []*html.Node, sanitizedBody *html.Node) *Document {
	return &Document{
		Generated: "docsan " + generated,
		Title:     node.Content(title),
		Outline:   jsonOutline,
		Sumtab:    jsonSumtab,
		DocLinks:  jsonLinks,
		Metas:     toMetas(metas),
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
	isCompareParagraph := node.And(node.Element("p"), node.AttrContains("class", "compare-to"))
	return node.Or(isScript, isStylesheetLink, isCompareParagraph)
}

func notInternalLinkSelector() node.Check {
	anchorSelector := node.Element("a")
	anchorTypeSelector := node.AttrNotPrefix("href", "#")
	return node.And(anchorSelector, anchorTypeSelector)
}

func disableAtributeSelector() node.Check {
	clickEvents := node.HasAttr("onclick")
	simultaxButton := node.AttrContains("class", "dyncal-button")
	return node.And(node.AnyElement(), clickEvents, node.Not(simultaxButton))
}

func addNoticePlaceholdersIfNeeded(body *html.Node) *html.Node {
	noticePlaceholders := node.FindFirst(body, noticePlaceholder())
	if noticePlaceholders == nil {
		return node.AddNoticePlaceholders(body, placeholderTargetSelector())
	}
	return body
}

func addSeeAlsoPlaceholdersIfNeeded(body *html.Node) *html.Node {
	seeAlsoPlaceholders := node.FindFirst(body, seeAlsoPlaceholder())
	if seeAlsoPlaceholders == nil {
		return node.AddSeeAlsoPlaceholders(body, placeholderTargetSelector())
	}
	return body
}

func placeholderTargetSelector() node.Check {
	hasAnnotatableClass := node.AttrContains("class", "annotatable")
	hasID := node.HasAttr("id")
	return node.And(node.AnyElement(), hasAnnotatableClass, hasID)
}

func noticePlaceholder() node.Check {
	isDiv := node.Element("div")
	isNoticePlaceholder := node.AttrPrefix("id", "notice_")
	return node.And(isDiv, isNoticePlaceholder)
}

func seeAlsoPlaceholder() node.Check {
	isDiv := node.Element("div")
	isNoticePlaceholder := node.AttrPrefix("id", "seealso_")
	return node.And(isDiv, isNoticePlaceholder)
}

func chapterTableSelector() node.Check {
	isTable := node.Element("table")
	isChapterType := node.AttrContains("class", "chapter-table")
	return node.And(isTable, isChapterType)
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

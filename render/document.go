package render

import (
	"encoding/json"
	"io"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	"ibfd.org/docsan/node"
)

type jsonType int

const (
	jsonArray jsonType = iota
	jsonObject
)

// JSON defines pre-rendered JSON
type JSON struct {
	json string
}

// Document defines a document to render as JSON
type Document struct {
	Generated  string              `json:"generated"`
	Title      string              `json:"title"`
	Outline    *JSON               `json:"outline"`
	Sumtab     *JSON               `json:"sumtab"`
	DocLinks   *JSON               `json:"links"`
	References *JSON               `json:"references"`
	Metas      []map[string]string `json:"metas"`
	Scripts    []map[string]string `json:"scripts"`
	Body       string              `json:"body"`
}

// Transform transforms a HTML node to a document structure for JSON output.
func Transform(htmlDoc *html.Node, generated string) *Document {
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	sumtabAttrChecker := node.AttrEquals("id", "sumtab")
	linksAttrChecker := node.AttrEquals("id", "links")
	refsAttrChecker := node.AttrEquals("id", "references")
	tocAttrChecker := node.AttrEquals("id", "script_toc")
	head := node.FindFirst(htmlDoc, node.Element("head"))
	title := node.FindFirst(head, node.Element("title"))
	jsonOutline := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, outLineAttrChecker)), jsonObject)
	jsonSumtab := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, sumtabAttrChecker)), jsonObject)
	jsonLinks := formatJSON(node.FindFirst(htmlDoc, node.And(scriptSelector, linksAttrChecker)), jsonObject)
	jsonRefs := formatJSONRefs(node.FindFirst(htmlDoc, node.And(scriptSelector, refsAttrChecker)))
	metas := node.FindAll(head, node.Element("meta"))
	scriptsToDelete := node.And(scriptSelector, node.Or(outLineAttrChecker, sumtabAttrChecker, linksAttrChecker, refsAttrChecker, tocAttrChecker))
	scripts := node.FindAll(head, node.And(scriptSelector, node.Not(scriptsToDelete)))
	body := node.FindFirst(htmlDoc, node.Element("body"))
	san1Body := addNoticePlaceholdersIfNeeded(body)
	san2Body := addSeeAlsoPlaceholdersIfNeeded(san1Body)
	san3Body := node.Remove(san2Body, scriptsToDelete)
	san4Body := node.ReplaceWithComments(san3Body, commentTargetSelector())
	san5Body := node.WrapTables(san4Body, chapterTableSelector())
	san6Body := node.DisableAttribute(san5Body, "onclick", disableAtributeSelector())
	return newDocument(generated, title, jsonOutline, jsonSumtab, jsonLinks, jsonRefs, metas, scripts, san6Body)
}

// ToJSON renders a document to JSON.
func (document *Document) ToJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

// newDocument create a new document
func newDocument(generated string, title *html.Node, jsonOutline *JSON, jsonSumtab *JSON, jsonLinks *JSON, jsonRefs *JSON, metas []*html.Node, scripts []*html.Node, sanitizedBody *html.Node) *Document {
	return &Document{
		Generated:  "docsan " + generated,
		Title:      node.Content(title),
		Outline:    jsonOutline,
		Sumtab:     jsonSumtab,
		DocLinks:   jsonLinks,
		References: jsonRefs,
		Metas:      toMetas(metas),
		Scripts:    node.ToMapArray(scripts),
		Body:       node.RenderChildrenCommentParent(sanitizedBody)}
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

func formatJSONRefs(n *html.Node) *JSON {
	if n == nil {
		return formatJSON(n, jsonArray)
	}
	return formatJSON(n.FirstChild, jsonArray)
}

func formatJSON(n *html.Node, jtype jsonType) *JSON {
	var data string
	if n == nil {
		switch jtype {
		case jsonArray:
			data = "[]"
		case jsonObject:
			data = "{}"
		}
	} else {
		data = n.FirstChild.Data
	}
	return &JSON{data}
}

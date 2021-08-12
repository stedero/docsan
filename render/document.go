package render

import (
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
	"ibfd.org/docsan/config"
	log "ibfd.org/docsan/log4u"
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

// DocumentFactory defines a document factory.
type DocumentFactory struct {
	generated                 string
	outlineSelector           node.Check
	sumtabSelector            node.Check
	linksSelector             node.Check
	refsSelector              node.Check
	tablesSelector            node.Check
	lookupSelector            node.Check
	tocSelector               node.Check
	specialCopyrightsSelector node.Check
	scriptsToDeleteSelector   node.Check
	scriptsToKeepSelector     node.Check
	commentTargetSelector     node.Check
	chapterTableSelector      node.Check
	disableAtributeSelector   node.Check
	noticePlaceholder         node.Check
	seeAlsoPlaceholder        node.Check
	placeholderTargetSelector node.Check
}

// Document defines a document to render as JSON
type Document struct {
	DocID             string              `json:"-"`
	Generated         string              `json:"generated"`
	Title             string              `json:"title"`
	Metas             []map[string]string `json:"metas"`
	Outline           *JSON               `json:"outline"`
	Sumtab            *JSON               `json:"sumtab"`
	DocLinks          *JSON               `json:"links"`
	SeeAlso           *JSON               `json:"seealso"`
	Tables            *JSON               `json:"tables"`
	Lookup            *JSON               `json:"lookup"`
	SpecialCopyrights *JSON               `json:"specialcopyrights"`
	Scripts           []map[string]string `json:"scripts"`
	Body              string              `json:"body"`
}

// NewDocumentFactory creates a document factory.
func NewDocumentFactory(appName string) *DocumentFactory {
	scriptSelector := node.Element("script")
	outLineAttrChecker := node.AttrEquals("id", "outline")
	sumtabAttrChecker := node.AttrEquals("id", "sumtab")
	linksAttrChecker := node.AttrEquals("id", "links")
	refsAttrChecker := node.AttrEquals("id", "references")
	tablesAttrChecker := node.AttrEquals("id", "tables")
	lookupAttrChecker := node.AttrEquals("id", "lookup")
	tocAttrChecker := node.AttrEquals("id", "script_toc")
	specialCopyrightsAttrChecker := node.AttrEquals("id", "specialcopyrights")
	scriptsToDeleteSelector := node.And(scriptSelector, node.Or(outLineAttrChecker, sumtabAttrChecker,
		linksAttrChecker, refsAttrChecker, tablesAttrChecker, lookupAttrChecker, tocAttrChecker, specialCopyrightsAttrChecker))
	return &DocumentFactory{
		generated:                 appName,
		outlineSelector:           node.And(scriptSelector, outLineAttrChecker),
		sumtabSelector:            node.And(scriptSelector, sumtabAttrChecker),
		linksSelector:             node.And(scriptSelector, linksAttrChecker),
		refsSelector:              node.And(scriptSelector, refsAttrChecker),
		tablesSelector:            node.And(scriptSelector, tablesAttrChecker),
		lookupSelector:            node.And(scriptSelector, lookupAttrChecker),
		tocSelector:               node.And(scriptSelector, tocAttrChecker),
		specialCopyrightsSelector: node.And(scriptSelector, specialCopyrightsAttrChecker),
		scriptsToDeleteSelector:   scriptsToDeleteSelector,
		scriptsToKeepSelector:     node.And(scriptSelector, node.Not(scriptsToDeleteSelector)),
		commentTargetSelector:     commentTargetSelector(),
		chapterTableSelector:      chapterTableSelector(),
		disableAtributeSelector:   disableAtributeSelector(),
		noticePlaceholder:         noticePlaceholder(),
		seeAlsoPlaceholder:        seeAlsoPlaceholder(),
		placeholderTargetSelector: placeholderTargetSelector()}
}

// Transform transforms a HTML node to a document structure for JSON output.
func (df *DocumentFactory) Transform(htmlDoc *html.Node) *Document {
	head := node.FindFirst(htmlDoc, node.Element("head"))
	metas := toMetas(node.FindAll(head, node.Element("meta")))
	docID := getDocID(metas)
	action := node.NewAction(docID)
	return &Document{
		DocID:             docID,
		Generated:         df.generated,
		Title:             node.Content(node.FindFirst(head, node.Element("title"))),
		Metas:             metas,
		Outline:           formatJSON(node.FindFirst(htmlDoc, df.outlineSelector), jsonObject),
		Sumtab:            formatJSON(node.FindFirst(htmlDoc, df.sumtabSelector), jsonObject),
		DocLinks:          formatJSON(node.FindFirst(htmlDoc, df.linksSelector), jsonObject),
		SeeAlso:           formatJSON(node.FindFirst(htmlDoc, df.refsSelector), jsonObject),
		Tables:            formatJSON(node.FindFirst(htmlDoc, df.tablesSelector), jsonArray),
		Lookup:            formatJSON(node.FindFirst(htmlDoc, df.lookupSelector), jsonArray),
		SpecialCopyrights: formatJSON(node.FindFirst(htmlDoc, df.specialCopyrightsSelector), jsonObject),
		Scripts:           node.ToMapArray(node.FindAll(head, df.scriptsToKeepSelector)),
		Body:              df.renderBody(htmlDoc, action)}
}

func (df *DocumentFactory) renderBody(htmlDoc *html.Node, action *node.Action) string {
	body1 := node.FindFirst(htmlDoc, node.Element("body"))
	body2 := df.addNoticePlaceholdersIfNeeded(action, body1)
	body3 := df.addSeeAlsoPlaceholdersIfNeeded(action, body2)
	body4 := node.Remove(body3, df.scriptsToDeleteSelector)
	body5 := node.ReplaceWithComments(body4, df.commentTargetSelector)
	body6 := action.WrapTables(body5, df.chapterTableSelector)
	body7 := action.DisableAttribute(body6, "onclick", df.disableAtributeSelector)
	return node.RenderChildren(body7)
}

// ToJSON renders a document to JSON.
func (document *Document) ToJSON() ([]byte, error) {
	if config.JSONPretty() {
		return json.MarshalIndent(document, "", "  ")
	} else {
		return json.Marshal(document)
	}
}

// MarshalJSON marshals a pre-rendered JSON object
func (j JSON) MarshalJSON() ([]byte, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(j.json), &result)
	if err != nil {
		log.Errorf("invalid JSON ignored: %s", j.json)
		return []byte("{}"), nil
	}
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

func (df *DocumentFactory) addNoticePlaceholdersIfNeeded(action *node.Action, body *html.Node) *html.Node {
	noticePlaceholders := node.FindFirst(body, df.noticePlaceholder)
	if noticePlaceholders == nil {
		return action.AddNoticePlaceholders(body, df.placeholderTargetSelector)
	}
	return body
}

func (df *DocumentFactory) addSeeAlsoPlaceholdersIfNeeded(action *node.Action, body *html.Node) *html.Node {
	seeAlsoPlaceholders := node.FindFirst(body, df.seeAlsoPlaceholder)
	if seeAlsoPlaceholders == nil {
		return action.AddSeeAlsoPlaceholders(body, df.placeholderTargetSelector)
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

func getDocID(metas []map[string]string) string {
	for _, meta := range metas {
		if strings.EqualFold(meta["name"], "docid") {
			return meta["content"]
		}
	}
	return "unknown"
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

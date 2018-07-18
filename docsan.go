package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"io/ioutil"

	"golang.org/x/net/html"
)

func main() {
	s := readFile("evdeudir_2006_112.html")
	doc := parse(s)
	toReplace := make([]*html.Node, 0)
	check := func(n *html.Node) {
		if ignore(n) {
			toReplace = append(toReplace, n)
		}
	}
	scanTree(doc, check)
	for _, node := range toReplace {
		parent := node.Parent
		parent.InsertBefore(toComment(node), node)
		parent.RemoveChild(node)
	}
	html.Render(os.Stdout, doc)
}

func parse(str string) *html.Node {
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func scanTree(n *html.Node, check func(*html.Node)) {
	check(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		scanTree(c, check)
	}
}

func toComment(n *html.Node) *html.Node {
	var b bytes.Buffer
	html.Render(&b, n)
	m := &html.Node{
		Type:     html.CommentNode,
		DataAtom: n.DataAtom,
		Data:     b.String(),
		Attr:     make([]html.Attribute, len(n.Attr)),
	}
	copy(m.Attr, n.Attr)
	return m
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("fail to read file %s: %v", filename, err)
	}
	return string(data)
}

func ignore(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "script" || (n.Data == "link" && isStylesheetLink(n.Attr))
}

func isStylesheetLink(attrs []html.Attribute) bool {
	for _, attr := range attrs {
		if attr.Key == "rel" && attr.Val == "stylesheet" {
			return true
		}
	}
	return false
}

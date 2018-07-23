package render

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/html"
	"ibfd.org/docsan/doc"
	"ibfd.org/docsan/node"
	"ibfd.org/docsan/san"
)

// ToJSON sanitizes HTML and renders to JSON
func ToJSON(w http.ResponseWriter, r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err == nil {
		log.Printf("read succesfully %d bytes", len(data))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		htmlDoc, err := san.SanitizeHTML(string(data))
		if err != nil {
			return err
		}
		head := node.FindFirst(htmlDoc, accept("head"))
		body := node.FindFirst(htmlDoc, accept("body"))
		title := node.FindFirst(head, accept("title"))
		metas := node.FindAll(head, accept("meta"))
		document := doc.NewDocument(title, metas, body)
		w.Write(document.ToJSON())
	}
	return nil
}

func accept(element string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == element
	}
}
